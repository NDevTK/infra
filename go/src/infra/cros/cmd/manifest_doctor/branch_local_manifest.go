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

	"infra/cros/internal/branch"
	"infra/cros/internal/git"
	"infra/cros/internal/manifestutil"
	"infra/cros/internal/osutils"
	"infra/cros/internal/repo"
	"infra/cros/internal/shared"

	"cloud.google.com/go/firestore"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	luciflag "go.chromium.org/luci/common/flag"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// Default location of manifest-internal project.
	manifestInternalProjectPath = "manifest-internal"

	firestoreProject    = "chromeos-bot"
	firestoreCollection = "LocalManifestBranchMetadatas"
)

type localManifestBrancher struct {
	subcommands.CommandRunBase
	authFlags            authcli.Flags
	chromeosCheckoutPath string
	minMilestone         int
	projectList          string
	projects             []string
	push                 bool
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
			b.Flags.BoolVar(&b.push, "push", false,
				"Whether or not to push changes to the remote.")
			return b
		}}
}

func (b *localManifestBrancher) validate() error {
	if b.minMilestone == -1 {
		return fmt.Errorf("--min_milestone required")
	}

	if b.chromeosCheckoutPath == "" {
		return fmt.Errorf("--chromeos_checkout required")
	} else if _, err := os.Stat(b.chromeosCheckoutPath); gerrs.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %s does not exist", b.chromeosCheckoutPath)
	} else if err != nil {
		return fmt.Errorf("error validating --chromeos_checkout=%s", b.chromeosCheckoutPath)
	}

	if len(b.projects) == 0 {
		return fmt.Errorf("at least one project is required")
	}

	return nil
}

func (b *localManifestBrancher) authToken(ctx context.Context) (oauth2.TokenSource, error) {
	authOpts, err := b.authFlags.Options()
	if err != nil {
		return nil, err
	}
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)
	return authenticator.TokenSource()
}

func (b *localManifestBrancher) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	// Common setup (argument validation, logging, etc.)
	ret := SetUp(b, a, args, env)
	if ret != 0 {
		return ret
	}

	ctx := context.Background()
	authToken, err := b.authToken(ctx)
	if err != nil {
		LogErr(err.Error())
		return 2
	}
	client, err := firestore.NewClient(ctx, firestoreProject, option.WithTokenSource(authToken))
	if err != nil {
		LogErr(err.Error())
		return 3
	}
	defer client.Close()

	if err := BranchLocalManifests(ctx, client, b.chromeosCheckoutPath, b.projects, b.minMilestone, !b.push); err != nil {
		LogErr(err.Error())
		return 4
	}

	return 0
}

// pinLocalManifest returns whether or not local_manifest.xml in the specified
// the project/branch is up to date (false if the file does not exist), and
// a potential error.
func pinLocalManifest(ctx context.Context, checkout, path, branch string, referenceManifest *repo.Manifest, dryRun bool) (bool, error) {
	// Checkout appropriate branch of project.
	projectPath := filepath.Join(checkout, path)
	if !osutils.PathExists(projectPath) {
		return false, fmt.Errorf("project path %s does not exist", projectPath)
	}

	if hasBranch, err := git.RemoteHasBranch(projectPath, "cros-internal", branch); err != nil {
		return false, errors.Annotate(err, "failed to ls-remote branch %s from remote for project %s", branch, path).Err()
	} else if !hasBranch {
		LogOut("branch %s does not exist for project %s, skipping...", branch, path)
		return false, nil
	}
	if err := git.Fetch(projectPath, "cros-internal", branch); err != nil {
		return false, errors.Annotate(err, "failed to fetch branch %s from remote for project %s", branch, path).Err()
	}
	if err := git.Checkout(projectPath, branch); err != nil {
		return false, errors.Annotate(err, "failed to checkout branch %s for project %s", branch, path).Err()
	}

	// Repair local manifest.
	localManifestPath := filepath.Join(projectPath, "local_manifest.xml")
	if _, err := os.Stat(localManifestPath); os.IsNotExist(err) {
		LogOut("local_manifest.xml does not exist for project %s, branch %s, skipping...", path, branch)
		return false, nil
	}

	localManifest, err := repo.LoadManifestFromFile(localManifestPath)
	if err != nil {
		return false, errors.Annotate(err, "failed to load local_manifest.xml from project %s, branch %s", path, branch).Err()
	}

	if err := manifestutil.PinManifestFromManifest(&localManifest, referenceManifest); err != nil {
		return false, errors.Annotate(err, "failed to pin local_manifest.xml from reference manifest for project %s, branch %s", path, branch).Err()
	}
	hasChanges, err := repo.UpdateManifestElementsInFile(localManifestPath, &localManifest)
	if err != nil {
		return false, errors.Annotate(err, "failed to write changes to local_manifest.xml for project %s, branch %s", path, branch).Err()
	}

	// If the manifest actually changed, commit and push those changes.
	if !hasChanges {
		LogOut("no changes needed for project %s, branch %s\n", path, branch)
		return true, nil
	}

	var commitMsg string
	commitMsg += fmt.Sprintf("Repair local_manifest.xml for branch %s\n\n", branch)
	commitMsg += "This CL was created by the Manifest Doctor.\n"
	if _, err := git.CommitAll(projectPath, commitMsg); err != nil {
		return false, errors.Annotate(err, "failed to commit changes for project %s, branch %s", path, branch).Err()
	}

	remotes, err := git.GetRemotes(projectPath)
	if err != nil {
		return false, errors.Annotate(err, "failed to get remotes for checkout of project %s", path).Err()
	}
	if len(remotes) > 1 {
		return false, fmt.Errorf("project %s has more than one remote, don't know which to push to", path)
	}
	if len(remotes) == 0 {
		return false, fmt.Errorf("project %s has no remotes", path)
	}

	remoteRef := git.RemoteRef{
		Remote: remotes[0],
		Ref:    fmt.Sprintf("refs/for/%s", branch) + "%submit",
	}
	pushFunc := func() error {
		return git.PushRef(projectPath, "HEAD", remoteRef, git.DryRunIf(dryRun))
	}
	if err := shared.DoWithRetry(ctx, shared.LongerOpts, pushFunc); err != nil {
		return false, errors.Annotate(err, "failed to push/upload changes for project %s, branch %s", path, branch).Err()
	}
	if !dryRun {
		LogOut("committed changes for project %s, branch %s\n", path, branch)
	} else {
		LogOut("would have committed changes (dry run) for project %s, branch %s\n", path, branch)
	}

	return true, nil
}

type localManifestBranchMetadata struct {
	PathToPrevSHA map[string]string `firestore:"prevshas"`
}

// BranchLocalManifests is responsible for doing the actual work of local manifest branching.
func BranchLocalManifests(ctx context.Context, dsClient *firestore.Client, checkout string, projects []string, minMilestone int, dryRun bool) error {
	branches, err := branch.BranchesFromMilestone(checkout, minMilestone)
	if err != nil {
		return errors.Annotate(err, "BranchesFromMilestone failure").Err()
	}

	manifestInternalPath := filepath.Join(checkout, manifestInternalProjectPath)
	if !osutils.PathExists(manifestInternalPath) {
		return fmt.Errorf("manifest-internal checkout not found at %s", manifestInternalPath)
	}

	errs := []error{}
	for _, branch := range branches {
		// Checkout appropriate branch in sentinel project.
		if err := git.Checkout(manifestInternalPath, branch); err != nil {
			errs = append(errs, errors.Annotate(err, "failed to checkout branch %s in %s", branch, manifestInternalProjectPath).Err())
			continue
		}

		// Read reference manifest.
		referencePath := filepath.Join(manifestInternalPath, "default.xml")
		referenceManifest, err := repo.LoadManifestFromFileWithIncludes(referencePath)
		if err != nil {
			errs = append(errs, errors.Annotate(err, "failed to load reference manifest for branch %s", branch).Err())
			continue
		}

		// Get SHA of reference manifest.
		output, err := git.RunGit(manifestInternalPath, []string{"rev-parse", "HEAD"})
		if err != nil {
			errs = append(errs, errors.Annotate(err, "failed to rev-parse branch %s in %s", branch, manifestInternalProjectPath).Err())
			continue
		}
		currentSHA := strings.TrimSpace(output.Stdout)

		// Load optimization data from Firestore.
		bm := localManifestBranchMetadata{
			PathToPrevSHA: make(map[string]string),
		}
		var bmDoc *firestore.DocumentRef
		docExists := true
		if dsClient != nil {
			bmDoc = dsClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, branch))
			if docsnap, err := bmDoc.Get(ctx); err != nil {
				errorCode, ok := status.FromError(err)
				if ok && errorCode.Code() == codes.NotFound {
					docExists = false
					LogErr("no history for branch %s, not skipping", branch)
				} else {
					LogErr(errors.Annotate(err, "failed to get history, attempting all branch/project combos").Err().Error())
				}
			} else {
				docsnap.DataTo(&bm)
			}
		}

		for _, path := range projects {
			// If the SHA for the reference manifest hasn't changed since the last update, no need to reprocess this
			// particular project/branch combo.
			previousSHA, ok := bm.PathToPrevSHA[path]
			if !ok {
				LogErr("no history for project %s, branch %s, not skipping", path, branch)
			} else if previousSHA == currentSHA {
				LogOut("no change in reference manifest since last pin for project %s, branch %s, skipping...", path, branch)
				continue
			}

			if didWork, err := pinLocalManifest(ctx, checkout, path, branch, referenceManifest, dryRun); err != nil {
				errs = append(errs, err)
				continue
			} else if !dryRun && didWork {
				// Update optimization data.
				bm.PathToPrevSHA[path] = currentSHA
			}
		}

		// Write optimization data.
		if !dryRun && dsClient != nil {
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
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.NewMultiError(errs...)
}
