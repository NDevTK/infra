package computemapping

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"infra/cros/internal/gerrit"
	"infra/cros/internal/git"
	"infra/tools/dirmd"
	dirmdpb "infra/tools/dirmd/proto"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/logging"
)

// MappingInfo groups a computed Mapping and affected files for a set
// of ChangeRevs in a project.
type MappingInfo struct {
	Mapping       *dirmd.Mapping
	AffectedFiles []string
}

// WorkdirCreation is a function signature that returns a path to a workdir,
// a cleanup function, and an error if one occurred.
type WorkdirCreation func() (string, func() error, error)

// cherryPickChangeRevs cherry picks changeRevs to dir.
//
// changeRevs must all have the same project.
func cherryPickChangeRevs(ctx context.Context, dir string, changeRevs []*gerrit.ChangeRev) error {
	for i, changeRev := range changeRevs {
		if i > 0 && changeRev.Project != changeRevs[0].Project {
			// Change revs are sorted by project in the callers.
			panic(
				"all changeRevs passed to checkoutChangeRevs must have the same Project",
			)
		}
	}

	sort.Slice(changeRevs, func(i, j int) bool {
		return changeRevs[i].ChangeNum < changeRevs[j].ChangeNum
	})

	// All changeRevs must have the same Host and Project, as checked above.
	googlesourceHost := strings.Replace(changeRevs[0].Host, "-review", "", 1)
	remote := fmt.Sprintf("https://%s/%s", googlesourceHost, changeRevs[0].Project)

	logging.Debugf(ctx, "cloning repo %q", remote)

	if err := git.Clone(remote, dir, git.Depth(1), git.NoTags()); err != nil {
		return err
	}

	for _, changeRev := range changeRevs {
		logging.Debugf(ctx, "fetching ref %q from repo %q", changeRev.Ref, remote)

		if err := git.Fetch(dir, remote, changeRev.Ref, git.NoTags()); err != nil {
			return err
		}

		if err := git.CherryPick(ctx, dir, "FETCH_HEAD"); err != nil {
			return err
		}
	}

	return nil
}

// computeMappingForChangeRevs checks out a project with changeRevs applied and
// computes the Mapping.
//
// changeRevs must all have the same project.
func computeMappingForChangeRevs(
	ctx context.Context,
	changeRevs []*gerrit.ChangeRev,
	workdirFn WorkdirCreation,
) (mapping *dirmd.Mapping, err error) {
	workdir, cleanup, err := workdirFn()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cleanupErr := cleanup(); cleanupErr != nil {
			err = cleanupErr
		}
	}()

	if err = cherryPickChangeRevs(ctx, workdir, changeRevs); err != nil {
		return nil, err
	}

	mapping, err = dirmd.ReadMapping(ctx, dirmdpb.MappingForm_COMPUTED, workdir)
	if err != nil {
		return nil, fmt.Errorf("failed to read DIR_METADATA for change revs %q: %w", changeRevs, err)
	}

	if mapping == nil {
		return nil, fmt.Errorf("got nil mapping for change revs %q", changeRevs)
	}

	return mapping, nil
}

// computeProjectMappingInfos calculates a projectMappingInfo for each project
// in changeRevs.
func ProjectInfos(
	ctx context.Context,
	changeRevs []*gerrit.ChangeRev,
	workdirFn WorkdirCreation,
) ([]*MappingInfo, error) {
	projectToChangeRevs := make(map[string][]*gerrit.ChangeRev)
	projectToAffectedFiles := make(map[string]stringset.Set)

	for _, changeRev := range changeRevs {
		project := changeRev.Project

		if _, found := projectToChangeRevs[project]; !found {
			projectToChangeRevs[project] = make([]*gerrit.ChangeRev, 0)
		}

		projectToChangeRevs[project] = append(projectToChangeRevs[project], changeRev)

		if _, found := projectToAffectedFiles[project]; !found {
			projectToAffectedFiles[project] = stringset.New(0)
		}

		projectToAffectedFiles[project].AddAll(changeRev.Files)
	}

	projectMappingInfos := make([]*MappingInfo, 0, len(projectToChangeRevs))

	// Use a sorted list of keys from projectToChangeRevs, so iteration order is
	// deterministic.
	keys := make([]string, 0, len(projectToChangeRevs))
	for project := range projectToChangeRevs {
		keys = append(keys, project)
	}

	sort.Strings(keys)

	for _, project := range keys {
		changeRevs := projectToChangeRevs[project]

		logging.Infof(ctx, "computing metadata for project %q", project)

		mapping, err := computeMappingForChangeRevs(ctx, changeRevs, workdirFn)
		if err != nil {
			return nil, err
		}

		projectMappingInfos = append(projectMappingInfos, &MappingInfo{
			AffectedFiles: projectToAffectedFiles[project].ToSlice(),
			Mapping:       mapping,
		})
	}

	return projectMappingInfos, nil
}
