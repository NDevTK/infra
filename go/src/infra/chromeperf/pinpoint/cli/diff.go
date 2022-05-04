package cli

import (
	"context"
	"fmt"
	"strings"

	"infra/chromeperf/pinpoint"
	pppb "infra/chromeperf/pinpoint/proto"

	"github.com/google/go-cmp/cmp"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/testing/protocmp"
)

// customDiffReporter is a simple custom reporter that only records differences
// detected during comparison.
type customDiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *customDiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *customDiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", r.path, vx, vy))
	}
}

func (r *customDiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *customDiffReporter) String() string {
	return strings.Join(r.diffs, "\n")
}

type diffReportLine struct {
	label, message string
}

type diffCmd struct {
	baseCommandRun
	params Param
	name   string
}

func (dc *diffCmd) getJob(ctx context.Context, name string) (*pppb.Job, error) {
	c, err := dc.pinpointClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &pppb.GetJobRequest{Name: pinpoint.LegacyJobName(name)}
	j, err := c.GetJob(ctx, req)
	if err != nil {
		return nil, errors.Annotate(err, "failed during GetJob").Err()
	}
	return j, nil
}

func (dc *diffCmd) Run(ctx context.Context, a subcommands.Application, args []string) error {
	job, err := dc.getJob(ctx, dc.name)
	if err != nil {
		return err
	}
	e := job.GetJobSpec().GetExperiment()
	if e == nil {
		return errors.Reason("unsupported job kind: %+v", job).Err()
	}
	diffs, err := diffJob(job)
	if err != nil {
		return err
	}
	for _, c := range diffs {
		fmt.Printf("%s diff results (-base, +experiment):\n%v\n", c.label, c.message)
	}
	return nil
}

func diffJob(job *pppb.Job) ([]diffReportLine, error) {
	e := job.GetJobSpec().GetExperiment()
	if e == nil {
		return nil, errors.Reason("unsupported job kind: %+v", job).Err()
	}
	comps := []struct {
		label       string
		left, right interface{}
	}{
		{"Commit", e.GetBaseCommit(), e.GetExperimentCommit()},
		{"Patch", e.GetBasePatch(), e.GetExperimentPatch()},
	}
	ret := []diffReportLine{}
	for _, c := range comps {
		r := &customDiffReporter{}
		cmp.Equal(c.left, c.right, protocmp.Transform(), protocmp.IgnoreEmptyMessages(), cmp.Reporter(r))
		ret = append(ret, diffReportLine{c.label, r.String()})
	}
	return ret, nil
}

func (dc *diffCmd) RegisterFlags(p Param) {
	dc.baseCommandRun.RegisterFlags(p)
	dc.Flags.StringVar(&dc.name, "name", "", text.Doc(`
		Required; the name of the job to diff.
		Example: "-name=XXXXXXXXXXXXXX"
	`))
}

func cmdDiff(p Param) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "diff",
		ShortDesc: "Show differences between job specifications for each arm of an experiment.",
		CommandRun: wrapCommand(p, func() pinpointCommand {
			return &diffCmd{params: p}
		}),
	}
}
