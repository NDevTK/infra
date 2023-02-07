// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary golangbuild is a luciexe binary that builds and tests the code for the
// Go project. It supports building and testing go.googlesource.com/go as well as
// Go project subrepositories (e.g. go.googlesource.com/net) and on different branches.
//
// To build and run this locally end-to-end, follow these steps:
//
//	luci-auth login -scopes "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/gerritcodereview"
//	cat > build.jsonpb <<EOF
//	{
//		"builder": {
//			"project": "go",
//			"bucket": "ci",
//			"builder": "linux-amd64"
//		},
//		"input": {
//			"properties": {
//				"project": "go"
//			},
//			"gitiles_commit": {
//				"host": "go.googlesource.com",
//				"project": "go",
//				"id": "27301e8247580e456e712a07d68890dc1e857000",
//				"ref": "refs/heads/master"
//			}
//		}
//	}
//	EOF
//	LUCIEXE_FAKEBUILD=./build.jsonpb golangbuild
//
// Modify `build.jsonpb` as needed in order to try different paths. The format of
// `build.jsonpb` is a JSON-encoded protobuf with schema `go.chromium.org/luci/buildbucket/proto.Build`.
// The input.properties field of this protobuf follows the `infra/experimental/golangbuildpb.Inputs`
// schema which represents input parameters that are specific to this luciexe, but may also contain
// namespaced properties that are injected by different services. For instance, CV uses the
// "$recipe_engine/cq" namespace.
//
// As an example, to try out a "try bot" path, try the following `build.jsonpb`:
//
//	{
//		"builder": {
//			"project": "go",
//			"bucket": "try",
//			"builder": "linux-amd64"
//		},
//		"input": {
//			"properties": {
//				"project": "go",
//				"$recipe_engine/cq": {
//					"active": true,
//					"runMode": "DRY_RUN"
//				}
//			},
//			"gerrit_changes": [
//				{
//					"host": "go-review.googlesource.com",
//					"project": "go",
//					"change": 460376,
//					"patchset": 1
//				}
//			]
//		}
//	}
//
// NOTE: by default, a luciexe fake build will discard the temporary directory created to run
// the build. If you'd like to retain the contents of the directory, specify a working directory
// to the golangbuild luciexe via the `--working-dir` flag. Be careful about where this working
// directory lives; particularly, make sure it isn't a subdirectory of a Go module a directory
// containing a go.mod file.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/errors"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/luciexe/build"
	"go.chromium.org/luci/luciexe/build/cv"

	"infra/experimental/golangbuild/golangbuildpb"
)

func main() {
	inputs := new(golangbuildpb.Inputs)
	build.Main(inputs, nil, nil, func(ctx context.Context, args []string, st *build.State) error {
		return run(ctx, args, st, inputs)
	})
}

func run(ctx context.Context, args []string, st *build.State, inputs *golangbuildpb.Inputs) (err error) {
	authOpts := chromeinfra.SetDefaultAuthOptions(auth.Options{
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gerritcodereview",
		},
	})
	httpClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts).Client()
	if err != nil {
		return err
	}

	// Install some tools we'll need, including a bootstrap toolchain.
	gorootBootstrap, err := installTools(ctx)
	if err != nil {
		return err
	}

	// Define working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Annotate(err, "Get CWD").Err()
	}
	workdir := filepath.Join(cwd, inputs.Project)
	gocacheDir := filepath.Join(cwd, "gocache")

	// Set up environment.
	env := environ.FromCtx(ctx)
	env.Load(inputs.Env)
	env.Set("GOROOT_BOOTSTRAP", gorootBootstrap)
	env.Set("GOBIN", "")
	env.Set("GOCACHE", gocacheDir)
	env.Set("GO_BUILDER_NAME", st.Build().GetBuilder().GetBuilder()) // TODO(mknyszek): This is underspecified. We may need Project and Bucket.
	ctx = env.SetInCtx(ctx)

	inputPb := st.Build().GetInput()

	// Fetch the repository into workdir.
	isDryRun := false
	if mode, err := cv.RunMode(ctx); err == nil {
		isDryRun = strings.HasSuffix(mode, "DRY_RUN")
	} else if err != cv.ErrNotActive {
		return err
	}
	if err := fetchRepo(ctx, httpClient, inputs.Project, workdir, inputPb.GetGitilesCommit(), inputPb.GetGerritChanges(), isDryRun); err != nil {
		return err
	}

	if inputs.Project == "go" {
		// Build Go.
		//
		// TODO(mknyszek): Support Windows and plan9 by changing the extension.
		// TODO(mknyszek): Support cross-compile-only modes, perhaps by having CompileGOOS
		// and CompileGOARCH repeated fields in the input proto to identify what to build.
		// TODO(mknyszek): Support split make/run and sharding.
		scriptExt := ".bash"
		allScript := "all"
		if inputs.RaceMode {
			allScript = "race"
		}
		if err := runGoScript(ctx, workdir, allScript+scriptExt); err != nil {
			return err
		}
	} else {
		// TODO(mknyszek): Add support for running subrepository tests. This needs to
		// somehow obtain a Go toolchain to test against.
		return fmt.Errorf("subrepository build/test is unimplemented")
	}
	return nil
}

// cipdDeps is an ensure file that describes all our CIPD dependencies.
//
// N.B. We assume a few tools are already available on the machine we're
// running on. Namely:
// - git
// - C/C++ toolchain
//
// TODO(mknyszek): Make sure Go 1.17 still works as the bootstrap toolchain since
// it's our published minimum.
var cipdDeps = `
@Subdir go_bootstrap
infra/3pp/tools/go/${platform} version:2@1.19.3
`

func installTools(ctx context.Context) (gorootBootstrap string, err error) {
	step, ctx := build.StartStep(ctx, "install tools")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
	}()

	io.WriteString(step.Log("ensure file"), cipdDeps)

	toolsRoot, err := os.Getwd()
	if err != nil {
		return "", err
	}
	toolsRoot = filepath.Join(toolsRoot, "tools")

	// Install packages.
	cmd := exec.CommandContext(ctx, "cipd",
		"ensure", "-root", toolsRoot, "-ensure-file", "-",
		"-json-output", filepath.Join(os.TempDir(), "go_bootstrap_ensure_results.json"))
	cmd.Stdin = strings.NewReader(cipdDeps)
	if err := runCommandAsStep(ctx, "cipd ensure", cmd, true); err != nil {
		return "", err
	}
	return filepath.Join(toolsRoot, "go_bootstrap"), nil
}

const (
	goHost       = "go.googlesource.com"
	goReviewHost = "go-review.googlesource.com"
	// N.B. Unfortunately Go still calls the main branch "master" due to technical issues.
	tipBranch = "master" // nocheck
)

func fetchRepo(ctx context.Context, hc *http.Client, project, dst string, commit *bbpb.GitilesCommit, changes []*bbpb.GerritChange, isDryRun bool) (err error) {
	step, ctx := build.StartStep(ctx, "fetch repo")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
	}()

	// Get the GerritChange.
	var change *bbpb.GerritChange
	if len(changes) > 1 {
		return fmt.Errorf("no support for multiple GerritChanges")
	} else if len(changes) != 0 {
		change = changes[0]
	}

	// Validate change and commit.
	if change != nil {
		if change.Host != goReviewHost {
			return fmt.Errorf("unsupported host %q, want %q", change.Host, goReviewHost)
		}
		if change.Project != project {
			return fmt.Errorf("subrepo tests do not support cross-project triggers for trybots: triggered by %q", project)
		}
	}
	if commit != nil {
		if commit.Host != goHost {
			return fmt.Errorf("unsupported host %q, want %q", commit.Host, goHost)
		}
		if commit.Project != project {
			if commit.Project != "go" {
				return fmt.Errorf("unsupported trigger project for subrepo tests: %s", commit.Project)
			}
			// Subrepo test triggered by a change from a different project. Fetch at HEAD
			// and download Go toolchain for this commit.
			return fmt.Errorf("subrepo tests unimplemented")
		}
	}
	switch {
	case change != nil && isDryRun:
		return fetchRepoForTry(ctx, hc, project, dst, change)
	case change != nil && !isDryRun:
		return fetchRepoForSubmit(ctx, hc, dst, change)
	case commit != nil:
		return fetchRepoForCI(ctx, hc, project, dst, commit)
	}
	// TODO(mknyszek): Fetch repo at HEAD here for subrepo tests.
	return fmt.Errorf("no commit or change specified for build and test")
}

func fetchRepoForTry(ctx context.Context, hc *http.Client, project, dst string, change *bbpb.GerritChange) (err error) {
	// TODO(mknyszek): We probably shouldn't use the /+archive endpoint if we can help it (i.e. if we have git).
	tarURL := fmt.Sprintf("https://%s/%s/+archive/refs/changes/%d/%d/%d.tar.gz", goHost, change.Project, change.Change%100, change.Change, change.Patchset)
	if err := fetchRepoTar(ctx, hc, project, dst, tarURL); err != nil {
		return err
	}
	return writeVersionFile(ctx, dst, fmt.Sprintf("%d/%d", change.Change, change.Patchset))
}

func fetchRepoForSubmit(ctx context.Context, hc *http.Client, dst string, change *bbpb.GerritChange) (err error) {
	// For submit, fetch HEAD for the branch this change is for, fetch the CL, and cherry-pick it.
	// TODO(mknyszek): We do a full git checkout. Consider caching.
	gc, err := gerrit.NewRESTClient(hc, change.Host, true)
	if err != nil {
		return err
	}
	changeInfo, err := gc.GetChange(ctx, &gerritpb.GetChangeRequest{
		Number:  change.Change,
		Project: change.Project,
	})
	if err != nil {
		return err
	}
	branch := changeInfo.Branch
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "--depth", "1", "-b", branch, "https://"+change.Host+"/"+change.Project, dst); err != nil {
		return err
	}
	ref := fmt.Sprintf("refs/changes/%d/%d/%d", change.Change%100, change.Change, change.Patchset)
	if err := runGit(ctx, "git fetch", "-C", dst, "fetch", "https://"+change.Host+"/"+change.Project, ref); err != nil {
		return err
	}
	return runGit(ctx, "git cherry-pick", "-C", dst, "cherry-pick", "FETCH_HEAD")
}

func fetchRepoForCI(ctx context.Context, hc *http.Client, project, dst string, commit *bbpb.GitilesCommit) (err error) {
	// TODO(mknyszek): We probably shouldn't use the /+archive endpoint if we can help it (i.e. if we have git).
	tarURL := "https://" + commit.Host + "/" + commit.Project + "/+archive/" + commit.Id + ".tar.gz"
	if err := fetchRepoTar(ctx, hc, project, dst, tarURL); err != nil {
		return err
	}
	return writeVersionFile(ctx, dst, commit.Id)
}

func writeVersionFile(ctx context.Context, dst, version string) (err error) {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("write VERSION (%s)", version))
	defer func() {
		// Any failure in this function is an infrastructure failure.
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
	}()
	contentsLog := step.Log("contents")

	f, err := os.Create(filepath.Join(dst, "VERSION"))
	if err != nil {
		return err
	}
	defer func() {
		r := f.Close()
		if err == nil {
			err = r
		} else {
			io.WriteString(step.Log("close error"), r.Error())
		}
	}()
	_, err = io.WriteString(io.MultiWriter(contentsLog, f), "devel "+version)
	return err
}

func runGit(ctx context.Context, stepName string, args ...string) (err error) {
	return runCommandAsStep(ctx, stepName, exec.CommandContext(ctx, "git", args...), true)
}

func runGoScript(ctx context.Context, goroot, script string) (err error) {
	dir := filepath.Join(goroot, "src")
	cmd := exec.CommandContext(ctx, "./"+script)
	cmd.Dir = dir
	return runCommandAsStep(ctx, script, cmd, false)
}

// runCommandAsStep runs the provided command as a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func runCommandAsStep(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (err error) {
	step, ctx := build.StartStep(ctx, stepName)
	defer func() {
		if infra {
			// Any failure in this function is an infrastructure failure.
			err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		}
		step.End(err)
	}()

	// Log the full command we're executing.
	var fullCmd bytes.Buffer
	envs := environ.FromCtx(ctx).Sorted()
	for _, env := range envs {
		fullCmd.WriteString(env)
		fullCmd.WriteString(" ")
	}
	if cmd.Dir != "" {
		fullCmd.WriteString("PWD=")
		fullCmd.WriteString(cmd.Dir)
		fullCmd.WriteString(" ")
	}
	fullCmd.WriteString(cmd.String())
	io.Copy(step.Log("commands"), &fullCmd)

	// Run the command.
	stdout := step.Log("stdout")
	stderr := step.Log("stderr")
	cmd.Env = envs
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
}

func fetchRepoTar(ctx context.Context, hc *http.Client, project, dst, tarURL string) (err error) {
	step, ctx := build.StartStep(ctx, "fetch repo tar")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
	}()
	if _, err := step.Log("url").Write([]byte(tarURL)); err != nil {
		return err
	}

	// Bump the timeout for downloading the tar.
	oldTimeout := hc.Timeout
	defer func() {
		hc.Timeout = oldTimeout
	}()
	hc.Timeout = 30 * time.Second

	res, err := hc.Get(tarURL)
	if err != nil {
		return errors.Annotate(err, "Fetching repo from %s", tarURL).Err()
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		slurp, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
		return errors.Annotate(errors.New(string(slurp)), "Fetching Go from %s", tarURL).Err()
	}
	// See golang.org/issue/11224 for a discussion on tree filtering.
	b, err := io.ReadAll(io.LimitReader(res.Body, maxSize(project)+1))
	if int64(len(b)) > maxSize(project) && err == nil {
		return errors.Annotate(errors.New("too big"), "Fetching Go from %s", tarURL).Err()
	}
	if err != nil {
		return errors.Annotate(err, "Fetching Go from %s", tarURL).Err()
	}
	if err := os.Mkdir(dst, os.ModePerm); err != nil {
		return errors.Annotate(err, "Mkdir %s", dst).Err()
	}
	if err := untar(ctx, bytes.NewReader(b), dst); err != nil {
		return errors.Annotate(err, "Extracting %s", dst).Err()
	}
	return nil
}

// untar reads the gzip-compressed tar file from r and writes it into dir.
//
// Copied from https://go.googlesource.com/build/+/refs/heads/master/internal/untar/untar.go
func untar(ctx context.Context, r io.Reader, dir string) (err error) {
	step, ctx := build.StartStep(ctx, "untar")
	defer step.End(err)

	log := step.Log("log")
	t0 := time.Now()
	nFiles := 0
	madeDir := map[string]bool{}
	defer func() {
		fmt.Fprintf(log, "extracted tarball into %s: %d files, %d dirs\n", dir, nFiles, len(madeDir))
	}()
	zr, err := gzip.NewReader(r)
	if err != nil {
		return errors.Annotate(err, "requires gzip-compressed body").Err()
	}
	tr := tar.NewReader(zr)
	loggedChtimesError := false
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if f.Typeflag == tar.TypeXGlobalHeader {
			// golang.org/issue/22748: git archive exports
			// a global header ('g') which after Go 1.9
			// (for a bit?) contained an empty filename.
			// Ignore it.
			continue
		}
		if !validRelPath(f.Name) {
			return errors.Reason("tar file contained invalid name %q", f.Name).Err()
		}
		rel := filepath.FromSlash(f.Name)
		abs := filepath.Join(dir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			dir := filepath.Dir(abs)
			if !madeDir[dir] {
				if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
					return err
				}
				madeDir[dir] = true
			}
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			modTime := f.ModTime
			if modTime.After(t0) {
				// Clamp modtimes at system time. See
				// golang.org/issue/19062 when clock on
				// buildlet was behind the gitmirror server
				// doing the git-archive.
				modTime = t0
			}
			if !modTime.IsZero() {
				if err := os.Chtimes(abs, modTime, modTime); err != nil && !loggedChtimesError {
					// benign error. Gerrit doesn't even set the
					// modtime in these, and we don't end up relying
					// on it anywhere (the gomote push command relies
					// on digests only), so this is a little pointless
					// for now.
					fmt.Fprintf(log, "error changing modtime: %v (further Chtimes errors suppressed)\n", err)
					loggedChtimesError = true // once is enough
				}
			}
			nFiles++
		case mode.IsDir():
			if err := os.MkdirAll(abs, 0755); err != nil {
				return err
			}
			madeDir[abs] = true
		case mode&os.ModeSymlink != 0:
			// TODO: ignore these for now. They were breaking x/build tests.
			// Implement these if/when we ever have a test that needs them.
			// But maybe we'd have to skip creating them on Windows for some builders
			// without permissions.
		default:
			return errors.Reason("tar file entry %s contained unsupported file type %v", f.Name, mode).Err()
		}
	}
	return nil
}

// maxSize controls artificial limits on how big of a compressed source tarball
// this package is willing to accept. It's expected humans may need to manage
// these limits every couple of years for the evolving needs of the Go project,
// and ideally not much more often.
//
// repo is a go.googlesource.com repo ("go", "net", and so on).
//
// Copied from go.googlesource.com/build/internal/sourcecache/source.go
func maxSize(repo string) int64 {
	switch repo {
	default:
		// As of 2021-11-22, a compressed tarball of Go source is 23 MB,
		// x/net is 1.2 MB,
		// x/build is 1.1 MB,
		// x/tools is 2.9 MB.
		return 100 << 20
	case "website":
		// In 2021, all content in x/blog (52 MB) and x/talks (74 MB) moved
		// to x/website. This makes x/website an outlier, with a compressed
		// tarball size of 135 MB. Give it some room to grow from there.
		return 200 << 20
	}
}

// Copied from go.googlesource.com/build/internal/untar/untar.go
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}
