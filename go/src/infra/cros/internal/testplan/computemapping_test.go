package testplan

import (
	"context"
	"infra/cros/internal/cmd"
	"infra/cros/internal/gerrit"
	"infra/cros/internal/git"
	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"
	"infra/tools/dirmd/proto/chromeos"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.chromium.org/chromiumos/config/go/test/plan"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestComputeProjectMappingInfos(t *testing.T) {
	ctx := context.Background()
	// Two changes from testprojectA, one from testprojectB.
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
		{
			ChangeRevKey: gerrit.ChangeRevKey{
				Host:      "chromium-review.googlesource.com",
				ChangeNum: 456,
			},
			Project: "chromium/testprojectA",
			Ref:     "refs/changes/45/456/2",
			Files:   []string{"DIR_METADATA"},
		},
		{
			ChangeRevKey: gerrit.ChangeRevKey{
				Host:      "chromium-review.googlesource.com",
				ChangeNum: 789,
			},
			Project: "chromium/testprojectB",
			Ref:     "refs/changes/78/789/5",
			Files:   []string{"test.c", "test.h"},
		},
	}

	// The newest change for each project should be checked out.
	git.CommandRunnerImpl = &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			{
				ExpectedCmd: []string{"git", "clone", "https://chromium.googlesource.com/chromium/testprojectA", "testdata"},
			},
			{
				ExpectedCmd: []string{"git", "fetch", "https://chromium.googlesource.com/chromium/testprojectA", "refs/changes/45/456/2"},
			},
			{
				ExpectedCmd: []string{"git", "checkout", "FETCH_HEAD"},
			},
			{
				ExpectedCmd: []string{"git", "clone", "https://chromium.googlesource.com/chromium/testprojectB", "testdata"},
			},
			{
				ExpectedCmd: []string{"git", "fetch", "https://chromium.googlesource.com/chromium/testprojectB", "refs/changes/78/789/5"},
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

	projectMappingInfos, err := computeProjectMappingInfos(ctx, changeRevs)
	if err != nil {
		t.Fatalf("computeProjectMappingInfos(%v) failed: %s", changeRevs, err)
	}

	// Both projects read the same DIR_METADATA, so have the same expected
	// Mapping.
	expectedMapping := &dirmd.Mapping{
		Dirs: map[string]*dirmdpb.Metadata{
			".": {
				Chromeos: &chromeos.ChromeOS{
					Cq: &chromeos.ChromeOS_CQ{
						SourceTestPlans: []*plan.SourceTestPlan{
							{
								EnabledTestEnvironments: []plan.SourceTestPlan_TestEnvironment{
									plan.SourceTestPlan_HARDWARE,
								},
								KernelVersions: &plan.SourceTestPlan_KernelVersions{},
							},
						},
					},
				},
			},
		},
	}
	expectedAffectedFiles := [][]string{
		{"a/b/test1.txt", "a/b/test2.txt", "DIR_METADATA"},
		{"test.c", "test.h"},
	}

	for i, pmi := range projectMappingInfos {
		if diff := cmp.Diff(
			expectedMapping.Proto(), pmi.Mapping.Proto(), protocmp.Transform(),
		); diff != "" {
			t.Errorf(
				"computeProjectMappingInfos returned unexpected diff in mappings at index %d (-want +got):\n%s",
				i, diff,
			)
		}

		sort.Strings(expectedAffectedFiles[i])
		sort.Strings(pmi.AffectedFiles)

		if !reflect.DeepEqual(expectedAffectedFiles[i], pmi.AffectedFiles) {
			t.Errorf(
				"computeProjectMappingInfos returned affectedFiles %v, expected %v",
				pmi.AffectedFiles,
				expectedAffectedFiles[i],
			)
		}
	}
}
