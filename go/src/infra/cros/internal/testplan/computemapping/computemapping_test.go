package computemapping_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"infra/cros/internal/assert"
	"infra/cros/internal/cmd"
	"infra/cros/internal/gerrit"
	"infra/cros/internal/git"
	"infra/cros/internal/testplan/computemapping"
	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"
	"infra/tools/dirmd/proto/chromeos"

	"go.chromium.org/chromiumos/config/go/test/plan"
)

func TestComputeProjectMappingInfos(t *testing.T) {
	ctx := context.Background()
	// Two changes from testprojectA, one from testprojectB.
	changeRevs := []*gerrit.ChangeRev{
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

	// Changes should be cherry-picked, ordered by project number.
	git.CommandRunnerImpl = &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			{
				ExpectedCmd: []string{
					"git", "clone",
					"https://chromium.googlesource.com/chromium/testprojectA", "good_dirmd",
					"--depth", "1", "--no-tags",
				},
			},
			{
				ExpectedCmd: []string{
					"git", "fetch",
					"https://chromium.googlesource.com/chromium/testprojectA", "refs/changes/23/123/5",
					"--no-tags",
				},
			},
			{
				ExpectedCmd: []string{"git", "cherry-pick", "FETCH_HEAD"},
			},
			{
				ExpectedCmd: []string{
					"git", "fetch",
					"https://chromium.googlesource.com/chromium/testprojectA", "refs/changes/45/456/2",
					"--no-tags",
				},
			},
			{
				ExpectedCmd: []string{"git", "cherry-pick", "FETCH_HEAD"},
			},
			{
				ExpectedCmd: []string{
					"git", "clone",
					"https://chromium.googlesource.com/chromium/testprojectB", "good_dirmd",
					"--depth", "1", "--no-tags",
				},
			},
			{
				ExpectedCmd: []string{
					"git", "fetch",
					"https://chromium.googlesource.com/chromium/testprojectB", "refs/changes/78/789/5",
					"--no-tags",
				},
			},
			{
				ExpectedCmd: []string{"git", "cherry-pick", "FETCH_HEAD"},
			},
		},
	}

	// Set workdirFn so the CommandRunners can know where commands are run,
	// and the DIR_METADATA in testdata/good_dirmd is read. Don't cleanup the testdata.
	workdirFn := func() (string, func() error, error) {
		cleanup := func() error { return nil }
		return "../testdata/good_dirmd", cleanup, nil
	}

	projectMappingInfos, err := computemapping.ProjectInfos(ctx, changeRevs, workdirFn)
	if err != nil {
		t.Fatalf("computeProjectMappingInfos(%v) failed: %s", changeRevs, err)
	}

	// Both projects read the same DIR_METADATA, so have the same expected
	// Mapping.
	expectedMapping := &dirmd.Mapping{
		Dirs: map[string]*dirmdpb.Metadata{
			"go/src/infra/cros/internal/testplan/testdata/good_dirmd": {
				Chromeos: &chromeos.ChromeOS{
					Cq: &chromeos.ChromeOS_CQ{
						SourceTestPlans: []*plan.SourceTestPlan{
							{
								TestPlanStarlarkFiles: []*plan.SourceTestPlan_TestPlanStarlarkFile{
									{
										Host:    "chromium.googlesource.com",
										Project: "repo1",
										Path:    "test1.star",
									},
									{
										Host:    "chromium.googlesource.com",
										Project: "repo2",
										Path:    "test2.star",
									},
								},
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
			expectedMapping.Dirs, pmi.Mapping.Dirs, protocmp.Transform(),
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

func TestComputeProjectMappingInfosBadDirmd(t *testing.T) {
	ctx := context.Background()
	// One changes from testprojectA.
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

	// The change for testprojectA should be cherry-picked.
	git.CommandRunnerImpl = &cmd.FakeCommandRunnerMulti{
		CommandRunners: []cmd.FakeCommandRunner{
			{
				ExpectedCmd: []string{
					"git", "clone",
					"https://chromium.googlesource.com/chromium/testprojectA", "bad_dirmd",
					"--depth", "1", "--no-tags",
				},
			},
			{
				ExpectedCmd: []string{
					"git", "fetch",
					"https://chromium.googlesource.com/chromium/testprojectA", "refs/changes/23/123/5",
					"--no-tags",
				},
			},
			{
				ExpectedCmd: []string{"git", "cherry-pick", "FETCH_HEAD"},
			},
		},
	}

	// Set workdirFn so the CommandRunners can know where commands are run,
	// and the DIR_METADATA in testdata/bad_dirmd is read. Don't cleanup the testdata.
	workdirFn := func() (string, func() error, error) {
		cleanup := func() error { return nil }
		return "../testdata/bad_dirmd", cleanup, nil
	}

	_, err := computemapping.ProjectInfos(ctx, changeRevs, workdirFn)
	if err == nil {
		t.Fatalf("expected error from computeProjectMappingInfos(%v)", changeRevs)
	}

	assert.ErrorContains(t, err, "failed to read DIR_METADATA")
}
