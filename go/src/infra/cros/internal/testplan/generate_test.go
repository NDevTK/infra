package testplan

import (
	"context"
	"infra/cros/internal/cmd"
	"infra/cros/internal/gerrit"
	"infra/cros/internal/git"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	configpb "go.chromium.org/chromiumos/config/go/api"
	buildpb "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/payload"
)

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	changeRevs := []*gerrit.ChangeRev{
		{
			ChangeRevKey: gerrit.ChangeRevKey{
				Host:      "chromium-review.googlesource.com",
				ChangeNum: 123,
			},
			Project: "chromium/testprojectA",
			Ref:     "refs/changes/23/123/5",
			Files:   []string{"a/b/test1.txt", "a/b/test2.txt"},
		},
	}
	git.CommandRunnerImpl = &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			{
				ExpectedCmd: []string{
					"git", "clone",
					"https://chromium.googlesource.com/chromium/testprojectA", "testdata",
					"--depth", "1", "--no-tags",
				},
			},
			{
				ExpectedCmd: []string{"git", "fetch",
					"https://chromium.googlesource.com/chromium/testprojectA", "refs/changes/23/123/5",
					"--depth", "1", "--no-tags",
				},
			},
			{
				ExpectedCmd: []string{"git", "checkout", "FETCH_HEAD"},
			},
		},
	}

	// Set workdirFn so the CommandRunners can know where commands are run,
	// and the DIR_METADATA in testdata is read. Don't cleanup the testdata.
	workdirFn = func(_, _ string) (string, error) { return "./testdata", nil }
	workdirCleanupFn = func(_ string) error { return nil }

	buildSummaryList := &buildpb.SystemImage_BuildSummaryList{
		Values: []*buildpb.SystemImage_BuildSummary{
			buildSummary("project1", "4.14", "chipsetA", "P"),
			buildSummary("project2", "4.14", "chipsetB", "R"),
			buildSummary("project3", "5.4", "chipsetA", ""),
		},
	}

	flatConfigList := &payload.FlatConfigList{
		Values: []*payload.FlatConfig{
			flatConfig("config1", "project1", configpb.HardwareFeatures_Fingerprint_KEYBOARD_BOTTOM_LEFT),
			flatConfig("config2", "project2", configpb.HardwareFeatures_Fingerprint_NOT_PRESENT),
		},
	}

	outputs, err := Generate(ctx, changeRevs, buildSummaryList, flatConfigList)

	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	expectedOutputs := []*Output{
		{
			Name:         "kernel-4.14",
			BuildTargets: []string{"project1", "project2"},
		},
		{
			Name:         "kernel-5.4",
			BuildTargets: []string{"project3"},
		},
	}

	if diff := cmp.Diff(
		expectedOutputs,
		outputs,
		cmpopts.SortSlices(func(i, j *Output) bool {
			return i.Name < j.Name
		}),
		cmpopts.SortSlices(func(i, j string) bool {
			return i < j
		}),
		cmpopts.EquateEmpty(),
	); diff != "" {
		t.Errorf("generate returned unexpected diff (-want +got):\n%s", diff)
	}
}
