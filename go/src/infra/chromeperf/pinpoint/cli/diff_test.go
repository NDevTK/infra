package cli

import (
	"testing"

	pppb "infra/chromeperf/pinpoint/proto"

	"github.com/google/go-cmp/cmp"
)

func TestDiffJob(t *testing.T) {
	tests := []struct {
		job     *pppb.Job
		want    []diffReportLine
		wantErr bool
	}{
		{&pppb.Job{}, nil, true},
		{&pppb.Job{
			JobSpec: &pppb.JobSpec{
				JobKind: &pppb.JobSpec_Bisection{},
			},
		}, nil, true},
		{&pppb.Job{
			JobSpec: &pppb.JobSpec{
				JobKind: &pppb.JobSpec_Experiment{
					Experiment: &pppb.Experiment{},
				},
			},
		}, []diffReportLine{
			{"Commit", ""},
			{"Patch", ""},
		}, false},
		{&pppb.Job{
			JobSpec: &pppb.JobSpec{
				JobKind: &pppb.JobSpec_Experiment{
					Experiment: &pppb.Experiment{
						BaseCommit: &pppb.GitilesCommit{
							Project: "project",
						},
						ExperimentCommit: &pppb.GitilesCommit{
							Project: "project",
						},
						BasePatch: &pppb.GerritChange{
							Project: "project",
						},
						ExperimentPatch: &pppb.GerritChange{
							Project:  "project",
							Change:   123,
							Patchset: 4,
						},
					},
				},
			},
		}, []diffReportLine{
			{"Commit", ""},
			// TODO(seanmccullough): Make this test less brittle.
			{"Patch", "protocmp.Transform({*proto.GerritChange})[\"change\"]:\n\t-: <invalid reflect.Value>\n\t+: 123\n\nprotocmp.Transform({*proto.GerritChange})[\"patchset\"]:\n\t-: <invalid reflect.Value>\n\t+: 4\n"},
		}, false},
	}
	for i, test := range tests {
		got, err := diffJob(test.job)
		if test.wantErr && err == nil {
			t.Fatalf("%d: unexpected nil error", i)
		}
		if !test.wantErr && err != nil {
			t.Fatalf("%d: unexpected error: %#v", i, err)
		}
		if diff := cmp.Diff(got, test.want, cmp.AllowUnexported(diffReportLine{})); diff != "" {
			t.Fatalf("%d: got %#v, want %#v", i, got, test.want)
		}
	}

}
