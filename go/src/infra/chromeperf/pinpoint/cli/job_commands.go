package cli

import (
	"fmt"
	"infra/chromeperf/pinpoint"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
)

type listJobs struct {
	baseCommandRun
	filter string
}

func cmdListJobs(p Param) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "list-jobs [--filter='']",
		ShortDesc: "Lists jobs tracked by Pinpoint",
		// TODO(dberris): Link to documentation about supported fields in the filter.
		LongDesc: text.Doc(`
			Prints out a list of jobs tracked by Pinpoint to stdout, possibly
			constrained by the filter. See https://aip.dev/160 for details on the
			filter syntax.
		`),
		CommandRun: func() subcommands.CommandRun {
			lj := &listJobs{}
			lj.RegisterDefaultFlags(p)
			lj.Flags.StringVar(&lj.filter, "filter", "", text.Doc(`
				Optional filter to apply to restrict the set of jobs listed.
			`))
			return lj
		},
	}
}

func (lj *listJobs) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, lj, env)
	c, err := lj.pinpointClient(ctx)
	if err != nil {
		fmt.Fprintf(a.GetErr(), "ERROR: Failed to create a Pinpoint client: %s\n", err)
		return 1
	}

	req := &pinpoint.ListJobsRequest{Filter: lj.filter}
	resp, err := c.ListJobs(ctx, req)
	if err != nil {
		fmt.Fprintf(a.GetErr(), "ERROR: Failed during ListJobs: %v\n", err)
		return 1
	}
	// TODO(chowski): have a nicer output format.
	fmt.Println(resp)
	return 0
}
