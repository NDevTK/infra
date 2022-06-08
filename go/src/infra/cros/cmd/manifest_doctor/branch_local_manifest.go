// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	gerrs "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"infra/cros/internal/branch"
	"infra/cros/internal/gerrit"
	"infra/cros/internal/git"
	"infra/cros/internal/manifestutil"
	"infra/cros/internal/osutils"
	"infra/cros/internal/repo"
	"infra/cros/internal/shared"

	"cloud.google.com/go/firestore"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	lucigerrit "go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/errors"
	luciflag "go.chromium.org/luci/common/flag"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// Default location of manifest-internal project.
	manifestInternalProjectPath = "manifest-internal"

	firestoreProject    = "chromeos-bot"
	firestoreCollection = "LocalManifestBranchMetadatas"

	internalHost = "https://chrome-internal-review.googlesource.com"
)

type localManifestBrancher struct {
	subcommands.CommandRunBase
	authFlags            authcli.Flags
	bbid                 int
	chromeosCheckoutPath string
	minMilestone         int
	specificBranches     []string
	projectList          string
	projects             []string
	push                 bool
	workerCount          int
}

func cmdLocalManifestBrancher(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "branch-local-manifest --chromeos_checkout ~/chromiumos " +
			" --min_milestone 90 --projects chromeos/project/foo,chromeos/project/bar",
		ShortDesc: "Repair local_manifest.xml on specified non-ToT branches.",
		CommandRun: func() subcommands.CommandRun {
			b := &localManifestBrancher{}
			b.authFlags = authcli.Flags{}
			b.authFlags.Register(b.GetFlags(), authOpts)
			b.Flags.StringVar(&b.chromeosCheckoutPath, "chromeos_checkout", "",
				"Path to full ChromeOS checkout.")
			b.Flags.IntVar(&b.minMilestone, "min_milestone", -1,
				"Minimum milestone of branches to consider. Used directly "+
					"in selecting release branches and indirectly for others.")
			b.Flags.Var(luciflag.CommaList(&b.projects), "projects",
				"Comma-separated list of project paths to consider. "+
					"At least one project is required.")
			b.Flags.Var(luciflag.CommaList(&b.specificBranches), "branches",
				"Comma-separated list of branches to process. "+
					"If set, minMilestone will be ignored and only these "+
					"branches will be processed.")
			b.Flags.BoolVar(&b.push, "push", false,
				"Whether or not to push changes to the remote.")
			b.Flags.IntVar(&b.workerCount, "j", 1, "Number of jobs to run for parallel operations.")
			b.Flags.IntVar(&b.bbid, "bbid", 0,
				"LUCI buildbucket ID which launched this manifest_doctor. If not set, no build URL will be referenced in the git commit.")
			return b
		}}
}

func (b *localManifestBrancher) validate() error {
	if b.minMilestone == -1 && len(b.specificBranches) == 0 {
		return gerrs.New("--min_milestone or --branches required")
	}

	if b.chromeosCheckoutPath == "" {
		return gerrs.New("--chromeos_checkout required")
	} else if _, err := os.Stat(b.chromeosCheckoutPath); gerrs.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %s does not exist", b.chromeosCheckoutPath)
	} else if err != nil {
		return fmt.Errorf("error validating --chromeos_checkout=%s", b.chromeosCheckoutPath)
	}

	if len(b.projects) == 0 {
		return gerrs.New("at least one project is required")
	}
	return nil
}

func (b *localManifestBrancher) newAuthenticator(ctx context.Context) (*auth.Authenticator, error) {
	authOpts, err := b.authFlags.Options()
	if err != nil {
		return nil, err
	}
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)
	return authenticator, nil
}

func (b *localManifestBrancher) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// Common setup (argument validation, logging, etc.)
	ret := SetUp(b, a, args, env)
	if ret != 0 {
		return ret
	}

	ctx := context.Background()
	authenticator, err := b.newAuthenticator(ctx)
	if err != nil {
		LogErr(err.Error())
		return 2
	}
	authToken, err := authenticator.TokenSource()
	if err != nil {
		LogErr(err.Error())
		return 5
	}
	client, err := firestore.NewClient(ctx, firestoreProject, option.WithTokenSource(authToken))
	if err != nil {
		LogErr(err.Error())
		return 3
	}
	defer client.Close()

	fsClient := &prodFirestoreClient{
		dsClient: client,
	}

	authClient, err := authenticator.Client()
	if err != nil {
		LogErr(err.Error())
		return 6
	}
	gerritClient, err := gerrit.NewClient(authClient)
	if err != nil {
		LogErr(err.Error())
		return 7
	}
	if err := b.BranchLocalManifests(ctx, fsClient, gerritClient); err != nil {
		LogErr(err.Error())
		return 4
	}

	return 0
}

type projectInfo struct {
	name string
	path string
}

type firestoreClient interface {
	readFirestoreData(ctx context.Context, branch string) (localManifestBranchMetadata, bool)
	writeFirestoreData(ctx context.Context, branch string, docExists bool, bm localManifestBranchMetadata)
}

type prodFirestoreClient struct {
	dsClient *firestore.Client
}

// readFirestoreData returns metadata for the given branch, and a boolean indicating whether
// or not the doc exists.
func (p *prodFirestoreClient) readFirestoreData(ctx context.Context, branch string) (localManifestBranchMetadata, bool) {
	bm := localManifestBranchMetadata{
		PathToPrevSHA: make(map[string]string),
	}
	docExists := true
	bmDoc := p.dsClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, branch))
	if docsnap, err := bmDoc.Get(ctx); err != nil {
		// If we know the failure is because the data doesn't exist, we can
		// still proceed.
		errorCode, ok := status.FromError(err)
		if ok && (errorCode.Code() == codes.NotFound) {
			docExists = false
			LogErr("no history for branch %s, not skipping", branch)
		} else {
			LogErr(errors.Annotate(err, "failed to get history, attempting all branch/project combos").Err().Error())
		}
	} else {
		// If sucessful, parse the data.
		docsnap.DataTo(&bm)
	}
	return bm, docExists
}

func (p *prodFirestoreClient) writeFirestoreData(ctx context.Context, branch string, docExists bool, bm localManifestBranchMetadata) {
	bmDoc := p.dsClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, branch))
	if docExists {
		if _, err := bmDoc.Set(ctx, bm); err != nil {
			LogErr(errors.Annotate(err, "failed to store optimization data for branch %s", branch).Err().Error())
		}
	} else {
		if _, err := bmDoc.Create(ctx, bm); err != nil {
			LogErr(errors.Annotate(err, "failed to store optimization data for branch %s", branch).Err().Error())
		}
	}
}

// pinLocalManifest returns whether or not local_manifest.xml in the specified
// the project/branch is up to date (false if the file does not exist), and
// a potential error.
func pinLocalManifest(ctx context.Context,
	checkout, path, branch string,
	referenceManifest *repo.Manifest, bbid int, dryRun bool,
	gerritClient gerrit.Client) (bool, error) {
	// Checkout appropriate branch of project.
	projectPath := filepath.Join(checkout, path)
	if !osutils.PathExists(projectPath) {
		return false, fmt.Errorf("project path %s does not exist", projectPath)
	}

	logPrefix := fmt.Sprintf("%s, %s", branch, path)

	var hasBranch bool
	if err := shared.DoWithRetry(ctx, shared.DefaultOpts, func() error {
		var err error
		hasBranch, err = git.RemoteHasBranch(projectPath, "cros-internal", branch)
		return err
	}); err != nil {
		return false, errors.Annotate(err, "%s: failed to ls-remote branch from remote", logPrefix).Err()
	} else if !hasBranch {
		LogOut("%s: branch does not exist for project, skipping...", logPrefix)
		return false, nil
	}
	if err := git.Fetch(projectPath, "cros-internal", branch); err != nil {
		return false, errors.Annotate(err, "%s: failed to fetch branch from remote", logPrefix).Err()
	}
	if err := git.Checkout(projectPath, branch); err != nil {
		return false, errors.Annotate(err, "%s: failed to checkout branch", logPrefix).Err()
	}

	// Repair local manifest.
	localManifestPath := filepath.Join(projectPath, "local_manifest.xml")
	if _, err := os.Stat(localManifestPath); os.IsNotExist(err) {
		LogOut("%s: local_manifest.xml does not exist, skipping...", logPrefix)
		return false, nil
	}

	localManifest, err := manifestutil.LoadManifestFromFile(localManifestPath)
	if err != nil {
		return false, errors.Annotate(err, "%s: failed to load local_manifest.xml", logPrefix).Err()
	}

	if err := manifestutil.PinManifestFromManifest(localManifest, referenceManifest); err != nil {
		if ferr, ok := err.(manifestutil.MissingProjectsError); ok {
			return false, errors.Annotate(err, "%s: failed to pin local_manifest.xml from reference manifest (missing projects %s)",
				logPrefix, ferr.MissingProjects).Err()
		}
		return false, errors.Annotate(err, "%s: failed to pin local_manifest.xml from reference manifest", logPrefix).Err()
	}
	hasChanges, err := manifestutil.UpdateManifestElementsInFile(localManifestPath, localManifest)
	if err != nil {
		return false, errors.Annotate(err, "%s: failed to write changes to local_manifest.xml", logPrefix).Err()
	}

	// If the manifest actually changed, commit and push those changes.
	if !hasChanges {
		LogOut("%s: no changes needed\n", logPrefix)
		return true, nil
	}

	var commitMsg string
	commitMsg += fmt.Sprintf("Repair local_manifest.xml for branch %s\n\n", branch)
	commitMsg += "This CL was created by the Manifest Doctor.\n"
	if bbid != 0 {
		commitMsg += fmt.Sprintf("See original build: http://go/bbid/%d\n", bbid)
	}
	// Need to acquire lock to make sure ChangeId is unique.
	commitHash, commitErr := git.CommitAll(projectPath, commitMsg)
	if commitErr != nil {
		return false, errors.Annotate(commitErr, "%s: failed to commit changes", logPrefix).Err()
	}

	remotes, err := git.GetRemotes(projectPath)
	if err != nil {
		return false, errors.Annotate(err, "%s: failed to get remotes for checkout of project", path).Err()
	}
	if len(remotes) > 1 {
		return false, fmt.Errorf("%s: project has more than one remote, don't know which to push to", path)
	}
	if len(remotes) == 0 {
		return false, fmt.Errorf("%s: project has no remotes", path)
	}

	remoteRef := git.RemoteRef{
		Remote: remotes[0],
		Ref:    fmt.Sprintf("refs/for/%s", branch),
	}
	pushFunc := func() error {
		return git.PushRef(projectPath, "HEAD", remoteRef, git.DryRunIf(dryRun))
	}
	if err := shared.DoWithRetry(ctx, shared.LongerOpts, pushFunc); err != nil {
		return false, errors.Annotate(err, "%s: failed to push/upload changes", logPrefix).Err()
	}
	if !dryRun {
		LogOut("%s: committed changes\n", logPrefix)

		// Look up gerrit change-id by commit hash.
		query := lucigerrit.ChangeQueryParams{
			Query: "commit:" + commitHash,
		}
		change, err := gerritClient.QueryChanges(ctx, internalHost, query)
		if err != nil {
			return false, errors.Annotate(err, "failed to get change-id for commit %s", commitHash).Err()
		}
		changeID := fmt.Sprintf("%d", change[0].ChangeNumber)
		LogOut("%s: committed changes\n", logPrefix)

		LogOut("%s: applying labels to changelist %s/q/%s", logPrefix, internalHost, changeID)
		reviewInput := &lucigerrit.ReviewInput{
			Labels: map[string]int{
				"Bot-Commit":      1,
				"Owners-Override": 1,
			},
		}
		_, err = gerritClient.SetReview(ctx, internalHost, changeID, reviewInput)
		if err != nil {
			return false, errors.Annotate(err, "%s: failed to apply labels to change", logPrefix).Err()
		}
		LogOut("%s: submitting changelist %s/q/%s", logPrefix, internalHost, changeID)
		if err := gerritClient.SubmitChange(ctx, internalHost, changeID); err != nil {
			return false, errors.Annotate(err, "%s: failed to submit changes", logPrefix).Err()
		}
	} else {
		LogOut("%s: would have committed changes (dry run)\n", logPrefix)
	}

	return true, nil
}

type localManifestBranchMetadata struct {
	PathToPrevSHA map[string]string `firestore:"prevshas"`
}

// BranchLocalManifests is responsible for doing the actual work of local manifest branching.
func (b *localManifestBrancher) BranchLocalManifests(ctx context.Context, fsClient firestoreClient, gerritClient gerrit.Client) error {
	checkout := b.chromeosCheckoutPath
	projects := b.projects
	minMilestone := b.minMilestone
	dryRun := !b.push

	var branches []string
	if len(b.specificBranches) != 0 {
		branches = b.specificBranches
	} else {
		var err error
		branches, err = branch.BranchesFromMilestone(checkout, minMilestone)
		if err != nil {
			return errors.Annotate(err, "BranchesFromMilestone failure").Err()
		}
	}

	manifestInternalPath := filepath.Join(checkout, manifestInternalProjectPath)
	if !osutils.PathExists(manifestInternalPath) {
		return fmt.Errorf("manifest-internal checkout not found at %s", manifestInternalPath)
	}

	errs := []error{}
	for _, branch := range branches {
		// Checkout appropriate branch in sentinel project.
		if err := git.Checkout(manifestInternalPath, branch); err != nil {
			err = errors.Annotate(err, "%s: failed to checkout branch in %s", branch, manifestInternalProjectPath).Err()
			LogErr("%s", err.Error())
			errs = append(errs, err)
			continue
		}

		// Read reference manifest.
		referencePath := filepath.Join(manifestInternalPath, "default.xml")
		referenceManifest, err := manifestutil.LoadManifestFromFileWithIncludes(referencePath)
		if err != nil {
			err = errors.Annotate(err, "failed to load reference manifest for branch %s", branch).Err()
			LogErr("%s", err.Error())
			errs = append(errs, err)
			continue
		}

		// Get SHA of reference manifest.
		output, err := git.RunGit(manifestInternalPath, []string{"rev-parse", "HEAD"})
		if err != nil {
			err = errors.Annotate(err, "failed to rev-parse branch %s in %s", branch, manifestInternalProjectPath).Err()
			LogErr("%s", err.Error())
			errs = append(errs, err)
			continue
		}
		currentSHA := strings.TrimSpace(output.Stdout)

		// Read optimization data from Firestore.
		bm, docExists := fsClient.readFirestoreData(ctx, branch)
		var wg sync.WaitGroup
		toProcess := make(chan string, len(projects))
		for _, path := range projects {

			// If the SHA for the reference manifest hasn't changed since the last update, no need to reprocess this
			// particular project/branch combo.
			previousSHA, ok := bm.PathToPrevSHA[path]
			if !ok {
				LogErr("%s, %s: no history, not skipping", branch, path)
			} else if previousSHA == currentSHA {
				LogOut("%s, %s: no change in reference manifest since last pin, skipping...", branch, path)
				continue
			}
			toProcess <- path
			wg.Add(1)
		}
		close(toProcess)

		optUpdates := &sync.Map{}
		for i := 1; i <= b.workerCount; i++ {
			go func(workerId int) {
				for path := range toProcess {
					if didWork, err := pinLocalManifest(ctx, checkout, path, branch, referenceManifest, b.bbid, dryRun, gerritClient); err != nil {
						LogErr("error: %s", err.Error())
						errs = append(errs, err)
						wg.Done()
						continue
					} else if !dryRun && didWork {
						// Update optimization data.
						optUpdates.Store(path, currentSHA)
					}
					wg.Done()
				}
			}(i)
		}
		wg.Wait()

		// Process optimization updates (can't do inline because map is not
		// thread-safe).
		optUpdates.Range(func(key, value interface{}) bool {
			bm.PathToPrevSHA[key.(string)] = value.(string)
			return true
		})

		// Write optimization data.
		if !dryRun {
			fsClient.writeFirestoreData(ctx, branch, docExists, bm)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.NewMultiError(errs...)
}
