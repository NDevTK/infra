// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builder

import (
	"context"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/cmd/cloudbuildhelper/fileset"
	"infra/cmd/cloudbuildhelper/gitignore"
	"infra/cmd/cloudbuildhelper/godep"
)

// Names of Go sources roots in the bundle for GOPATH and modules mode.
const (
	goPathRoot = "_gopath"
	goModRoot  = "_gomod"
)

// Locations of files used to track dependencies in modules mode.
const (
	bundledGoModPath      = goModRoot + "/go.mod"
	bundledModulesTxtPath = goModRoot + "/vendor/modules.txt"
)

// What go dependency mechanisms the bundle should use.
type bundleMode string

const (
	bundleUnknown bundleMode = "unknown"
	bundleGoPath  bundleMode = "GOPATH"
	bundleModules bundleMode = "modules"
)

// runGoGAEBundleBuildStep executes manifest.GoGAEBundleBuildStep.
func runGoGAEBundleBuildStep(ctx context.Context, inv *stepRunnerInv) error {
	mode := bundleGoPath
	if inv.BuildStep.ModulesMode {
		mode = bundleModules
	}

	logging.Infof(ctx, "Bundling %q in %s mode", inv.BuildStep.GoGAEBundle, mode)

	// Hybrid bundles aren't allowed.
	if cur := currentMode(inv.Output); cur != bundleUnknown && cur != mode {
		return errors.Reason("the bundle is already in %s mode, but being extended using %s mode", cur, mode).Err()
	}

	yamlPath, err := filepath.Abs(inv.BuildStep.GoGAEBundle)
	if err != nil {
		return errors.Annotate(err, "failed to convert the path %q to absolute", inv.BuildStep.GoGAEBundle).Err()
	}

	// Read go runtime version from the YAML to know what Go build flags to use.
	//
	// It is either e.g. "go113" for GAE Standard or "go1.13" or just "go" for
	// GAE Flex.
	runtime, err := readRuntime(yamlPath)
	if err != nil {
		return err
	}
	logging.Infof(ctx, "Runtime is %q", runtime)
	if runtime != "go" && !strings.HasPrefix(runtime, "go1") {
		return errors.Reason("%q is not a supported go runtime", runtime).Err()
	}
	var goMinorVer int64
	if strings.HasPrefix(runtime, "go1") {
		runtime = strings.ReplaceAll(runtime, ".", "")
		if goMinorVer, err = strconv.ParseInt(runtime[3:], 10, 32); err != nil {
			return errors.Annotate(err, "can't parse %q", runtime).Err()
		}
	}

	// The directory with `main` package.
	mainDir := filepath.Dir(yamlPath)

	// Get a build.Context as if we are building for linux amd64. We primarily use
	// it to call its MatchFile method to check build tags.
	bc := buildContext(mainDir, int(goMinorVer))

	// Load the main package and all its transitive dependencies (they are stored
	// as a graph of packages.Package that can be accessed via pointer chasing
	// from the loaded root packages.Package).
	mainPkg, err := loadPackageTree(ctx, bc)
	if err != nil {
		return err
	}
	if mode == bundleModules && (mainPkg.Module == nil || !mainPkg.Module.Main) {
		return errors.Reason("the main package is not a main module").Err()
	}

	// In modules mode we should keep track of visited dependencies to build the
	// `go.mod` and `vendors/modules.txt` files describing packages from non-main
	// modules.
	//
	// If we are bundling multiple GAE apps via multiple GoGAEBundleBuildSteps, we
	// should keep adding dependencies additively. prepareModDeps(...) loads the
	// existing godep.Deps state (if any) from inv.Output to keep appending to it.
	var modDeps *godep.Deps
	if mode == bundleModules {
		modDeps, err = prepareModDeps(mainPkg.Module, inv.Output)
		if err != nil {
			return errors.Annotate(err, "preparing dependency tracker").Err()
		}
	}

	// In modules mode the main module goes into "_gomod" and all other modules
	// go under "_gomod/vendor" (where Go wants them). In GOPATH mode all packages
	// should be under a single GOPATH root "_gopath/src".
	var packageDest func(pkg *packages.Package) (string, error)
	if mode == bundleModules {
		packageDest = func(pkg *packages.Package) (string, error) {
			switch {
			case pkg.Module == nil:
				return "", errors.Reason("not in a module").Err()
			case pkg.Module.Main:
				var relToMod string
				switch {
				case pkg.PkgPath == pkg.Module.Path:
					relToMod = "."
				case !strings.HasPrefix(pkg.PkgPath, pkg.Module.Path+"/"):
					return "", errors.Reason("module %q doesn't match the package import path", pkg.Module.Path).Err()
				default:
					relToMod = pkg.PkgPath[len(pkg.Module.Path)+1:]
				}
				return filepath.Join(goModRoot, filepath.FromSlash(relToMod)), nil
			default:
				return filepath.Join(goModRoot, "vendor", filepath.FromSlash(pkg.PkgPath)), nil
			}
		}
	} else if mode == bundleGoPath {
		packageDest = func(pkg *packages.Package) (string, error) {
			return filepath.Join(goPathRoot, "src", filepath.FromSlash(pkg.PkgPath)), nil
		}
	} else {
		panic("impossible")
	}

	// Respect .gcloudignore files when traversing the GAE app directory to avoid
	// uploading unnecessary files as "static files". We don't care about any
	// other directories, since we pick only *.go files from them.
	excludedByIgnoreFile, err := gitignore.NewExcluder(mainDir, ".gcloudignore")
	if err != nil {
		return errors.Annotate(err, "when loading .gcloudignore files").Err()
	}

	// The directory inside the bundle that should contain the `main` package.
	mainPkgDestRel, err := packageDest(mainPkg)
	if err != nil {
		return errors.Annotate(err, "when finding where to put the main package").Err()
	}
	// Absolute path to it in the staging directory.
	mainPkgDestAbs := filepath.Join(inv.Manifest.ContextDir, mainPkgDestRel)

	// Copy all files that make up "main" package (they can be only at the root
	// of `mainDir`), and copy all non-go files recursively (they can potentially
	// be referenced by static_files in app.yaml). We'll deal with Go dependencies
	// separately.
	err = inv.addFilesToOutput(ctx, mainDir, mainPkgDestAbs, func(absPath string, isDir bool) bool {
		switch {
		case excludedByIgnoreFile(absPath, isDir):
			return true // respect .gcloudignore exclusions
		case isDir:
			return false // do not exclude directories, they may contain static files
		}
		rel, err := relPath(mainDir, absPath)
		if err != nil {
			panic(fmt.Sprintf("impossible: %s", err))
		}
		switch {
		// Do not exclude non-code files regardless of where they are.
		case !isGoSourceFile(rel):
			return false
		// Exclude code files not in the mainDir. If they are needed, they'll be
		// discovered by the next step that traverses Go dependencies.
		case rel != filepath.Base(rel):
			return true
		// For code files in the mainDir, pick up only ones matching the build
		// context (linux amd64).
		default:
			matches, err := bc.MatchFile(mainDir, rel)
			if err != nil {
				logging.Warningf(ctx, "Failed to check whether %q matches the build context, skipping it: %s", absPath, err)
				return true
			}
			return !matches
		}
	})
	if err != nil {
		return err
	}

	// Drop empty .gcloudignore in the main directory. We already skipped ignored
	// files, but gcloud wants some .gcloudignore anyway, creating the default one
	// otherwise.
	if err := inv.Output.AddFromMemory(filepath.Join(mainPkgDestRel, ".gcloudignore"), nil, nil); err != nil {
		return errors.Annotate(err, "failed to create .gcloudignore").Err()
	}

	// We moved the main package to be somewhere under "_gomod" or "_gopath" to
	// make the bundle be a self-contained Go tree. But the authors of the
	// manifest expect main package files (in particular various GAE YAMLs) be
	// reachable under their original names. They don't really "know" or care
	// about "_gomod" and "_gopath". Make a symlink that puts the main directory
	// to where it is really expected to make YAMLs addressable.
	linkName, err := relPath(inv.Manifest.ContextDir, inv.BuildStep.Dest)
	if err != nil {
		return err
	}
	linkTarget, err := relPath(filepath.Dir(inv.BuildStep.Dest), mainPkgDestAbs)
	if err != nil {
		return err
	}
	if err := inv.Output.AddSymlink(linkName, linkTarget); err != nil {
		return errors.Annotate(err, "failed to setup a symlink to the main package").Err()
	}

	// Packages for different go versions may have different files in them due to
	// filtering based on build tags. For each Go runtime we keep a separate map
	// of visited packages in this runtime. In practice it means if the GAE app
	// uses more than one runtime, all packages will be visited more than once.
	// Each separate visit may add more files to the output (or just revisit
	// already added ones, which is a noop).
	goDeps := inv.State.goDeps(runtime)

	errs := 0    // number of errors in packages.Visit below
	visited := 0 // number of packages actually visited
	copied := 0  // number of files copied

	reportErr := func(format string, args ...interface{}) {
		logging.Errorf(ctx, format, args...)
		errs++
	}

	// Copy all transitive dependencies into the bundle.
	logging.Infof(ctx, "Copying transitive dependencies...")
	packages.Visit([]*packages.Package{mainPkg}, nil, func(pkg *packages.Package) {
		switch {
		case errs != 0:
			return // failing already
		case !goDeps.Add(pkg.ID):
			return // added it already in some previous build step
		case isStdlib(bc, pkg):
			return // we are not bundling stdlib packages
		default:
			visited++
		}

		// List of absolute file paths to copy into the output. They all must be in
		// the same directory (the package directory). At least one *.go file is
		// expected there.
		var filesToAdd []string

		// We visit GoFiles and IgnoredFiles because we want to recheck the build
		// tags using bc.MatchFile: packages.Load *always* uses the current Go
		// version tags, but we want to apply bc.ReleaseTags instead. It means we
		// may need to pick up some files rejected by packages.Load (they end up in
		// IgnoredFiles list), or reject some files from GoFiles.
		addGoFiles := func(paths []string) {
			for _, p := range paths {
				switch match, err := bc.MatchFile(filepath.Split(p)); {
				case err != nil:
					reportErr("Failed to check build tags of %q: %s", p, err)
				case match:
					filesToAdd = append(filesToAdd, p)
				}
			}
		}
		addGoFiles(pkg.GoFiles)
		addGoFiles(pkg.IgnoredFiles)

		if errs != 0 {
			return
		}
		if len(filesToAdd) == 0 {
			logging.Warningf(ctx, "Skipping package %s: no relevant *.go files", pkg.PkgPath)
			return
		}

		// packages.Package doesn't tell the package directory path. Verify all *.go
		// files we discovered come from the same directory. It is the package
		// directory we are after.
		srcDir := filepath.Dir(filesToAdd[0])
		for _, path := range filesToAdd {
			if filepath.Dir(path) != srcDir {
				reportErr("Expected %q to be under %q", path, srcDir)
			}
		}
		if errs != 0 {
			return
		}

		// Add non-go files, like *.c or files embedded via "go:embed". They must
		// be under the package directory, but may be in a subdirectory (in case of
		// "go:embed").
		addNonGoFile := func(path string) {
			rel, err := filepath.Rel(srcDir, path)
			if err != nil {
				reportErr("Filed to get relative path of %q", path)
				return
			}
			if rel == "." || !filepath.IsLocal(rel) {
				reportErr("Expected %q to be under %q", path, srcDir)
				return
			}
			filesToAdd = append(filesToAdd, path)
		}
		for _, path := range pkg.OtherFiles {
			addNonGoFile(path)
		}
		for _, path := range pkg.EmbedFiles {
			addNonGoFile(path)
		}
		if errs != 0 {
			return
		}

		// Decide the destination directory in the bundle based on the module.
		dstDir, err := packageDest(pkg)
		if err != nil {
			reportErr("Cant decide where to put %v: %s", pkg.GoFiles, err)
			return
		}

		// Add all discovered files to the tarball.
		for _, path := range filesToAdd {
			name, err := filepath.Rel(srcDir, path)
			if err != nil {
				// We verified paths already above.
				panic(fmt.Sprintf("impossible filepath.Rel error: %s", err))
			}
			err = inv.Output.AddFromDisk(path, filepath.Join(dstDir, name), nil)
			if err != nil {
				reportErr("Failed to copy %q to the tarball: %s", path, err)
			} else {
				copied++
			}
		}
		if errs != 0 {
			return
		}

		// In modules mode record this package as a dependency of the main module
		// to make it show up in the generated go.mod. We don't need to do anything
		// for packages from the main module: they aren't tracked in go.mod.
		if mode == bundleModules && !pkg.Module.Main {
			if err := modDeps.Add(pkg.PkgPath, pkg.Module.Path, pkg.Module.GoVersion); err != nil {
				reportErr("Error adding %q as a module dependency: %s", pkg.PkgPath, err)
			}
		}
	})
	if errs != 0 {
		return errors.Reason("failed to add Go files to the tarball, see the log").Err()
	}
	logging.Infof(ctx, "Visited %d packages and copied %d files", visited, copied)

	// Generate go.mod and modules.txt describing bundled dependencies.
	if mode == bundleModules {
		logging.Infof(ctx, "Writing %s and %s", bundledGoModPath, bundledModulesTxtPath)
		state, err := modDeps.Save()
		if err != nil {
			return errors.Annotate(err, "generating bundled go.mod").Err()
		}
		err = inv.Output.AddFromMemory(bundledGoModPath, state.GoMod, nil)
		if err != nil {
			return errors.Annotate(err, "adding bundled go.mod").Err()
		}
		err = inv.Output.AddFromMemory(bundledModulesTxtPath, state.ModulesTxt, nil)
		if err != nil {
			return errors.Annotate(err, "adding bundled modules.txt").Err()
		}
	}

	// If app.yaml and go.mod are in different directories, copy static files into
	// the module root to workaround b/323980048. Conflicts are possible (though
	// unlikely), fail on any.
	if mode == bundleModules && mainPkg.PkgPath != mainPkg.Module.Path {
		err := hackStaticFiles(ctx,
			inv.Output,
			mainDir,
			goModRoot,
			excludedByIgnoreFile,
		)
		if err != nil {
			return errors.Annotate(err, "copying static files into the module root").Err()
		}
	}

	// Drop a script that can be used to manually test correctness of this bundle:
	//
	// $ cd _gomod
	// $ eval `./goenv`
	// $ go build -v ./...
	//
	// This script isn't supposed to be used for anything important though.
	var scriptPath string
	var scriptBody string
	switch mode {
	case bundleModules:
		scriptPath = filepath.Join(goModRoot, "goenv")
		scriptBody = envScriptModules
	case bundleGoPath:
		scriptPath = filepath.Join(goPathRoot, "goenv")
		scriptBody = envScriptGoPath
	default:
		panic("impossible")
	}
	return inv.Output.AddFromMemory(scriptPath, []byte(scriptBody), &fileset.File{
		Executable: true,
	})
}

// readRuntime reads `runtime` field in the YAML file.
func readRuntime(path string) (string, error) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Annotate(err, "failed to read %q", path).Err()
	}

	var appYaml struct {
		Runtime string `yaml:"runtime"`
	}
	if err := yaml.Unmarshal(blob, &appYaml); err != nil {
		return "", errors.Annotate(err, "file %q is not a valid YAML", path).Err()
	}

	return appYaml.Runtime, nil
}

// buildContext returns a build.Context targeting linux-amd64.
//
// If goMinorVer is not 0, sets ReleaseTags to pick the specific go release.
func buildContext(mainDir string, goMinorVer int) *build.Context {
	bc := build.Default
	bc.GOARCH = "amd64"
	bc.GOOS = "linux"
	bc.Dir = mainDir
	if goMinorVer != 0 {
		bc.ReleaseTags = nil
		for i := 1; i <= goMinorVer; i++ {
			bc.ReleaseTags = append(bc.ReleaseTags, fmt.Sprintf("go1.%d", i))
		}
	}
	return &bc
}

// loadPackageTree loads the main package with its dependencies.
func loadPackageTree(ctx context.Context, bc *build.Context) (*packages.Package, error) {
	logging.Infof(ctx, "Loading the package tree...")

	// Note: this can actually download files into the modules cache when running
	// in module mode and thus can be quite slow.
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedModule |
			packages.NeedEmbedFiles,
		Context: ctx,
		Logf:    func(format string, args ...interface{}) { logging.Debugf(ctx, format, args...) },
		Dir:     bc.Dir,
		Env:     append(os.Environ(), "GOOS="+bc.GOOS, "GOARCH="+bc.GOARCH),
	}, ".")
	if err != nil {
		return nil, errors.Annotate(err, "failed to load the main package").Err()
	}

	// `packages.Load` records some errors inside packages.Package.
	errs := 0
	visited := 0
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		visited++
		for _, err := range pkg.Errors {
			logging.Errorf(ctx, "Error loading package %q: %s", pkg.ID, err)
			errs++
		}
	})
	if errs != 0 {
		return nil, errors.Reason("failed to load the package tree").Err()
	}

	// We expect only one package to match our load query.
	if len(pkgs) != 1 {
		return nil, errors.Reason("expected to load 1 package, but got %d", len(pkgs)).Err()
	}

	// Make sure it is indeed `main` and log its path in the package tree.
	mainPkg := pkgs[0]
	if mainPkg.PkgPath == "" {
		return nil, errors.Reason("could not figure out import path of the main package").Err()
	}
	logging.Infof(ctx, "Import path is %q", mainPkg.PkgPath)
	if mainPkg.Name != "main" {
		return nil, errors.Annotate(err, "only \"main\" package can be bundled, got %q", mainPkg.Name).Err()
	}
	if mainPkg.Module != nil {
		logging.Infof(ctx, "Module is %q at %q", mainPkg.Module.Path, mainPkg.Module.Dir)
	}

	logging.Infof(ctx, "Transitively depends on %d packages (including stdlib)", visited-1)
	return mainPkg, nil
}

// relPath calls filepath.Rel and annotates the error.
func relPath(base, path string) (string, error) {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return "", errors.Annotate(err, "failed to calculate rel(%q, %q)", base, path).Err()
	}
	return rel, nil
}

// isGoSourceFile returns true if rel may be read by Go compiler.
//
// See https://golang.org/src/go/build/build.go.
func isGoSourceFile(rel string) bool {
	switch filepath.Ext(rel) {
	case ".go", ".c", ".cc", ".cxx", ".cpp", ".m", ".s", ".h", ".hh", ".hpp", ".hxx", ".f", ".F", ".f90", ".S", ".sx", ".swig", ".swigcxx":
		return true
	default:
		return false
	}
}

// isStdlib returns true if the package has its *.go files under GOROOT.
func isStdlib(bc *build.Context, pkg *packages.Package) bool {
	switch {
	case pkg.Name == "unsafe":
		return true // this package is a magical indicator and has no Go files
	case len(pkg.GoFiles) == 0:
		return false // assume other stdlib packages have Go files
	default:
		root := filepath.Clean(bc.GOROOT) + string(filepath.Separator)
		return strings.HasPrefix(pkg.GoFiles[0], root)
	}
}

// prepareModDeps loads godep.Deps based on the state in the output.
func prepareModDeps(main *packages.Module, out *fileset.Set) (*godep.Deps, error) {
	// Existing go.mod with all dependencies of the main module.
	mainModPath := filepath.Join(main.Dir, "go.mod")
	mainModBlob, err := os.ReadFile(mainModPath)
	if err != nil {
		return nil, errors.Annotate(err, "reading main module's go.mod").Err()
	}
	mainMod, err := modfile.Parse(mainModPath, mainModBlob, nil)
	if err != nil {
		return nil, errors.Annotate(err, "parsing main module's go.mod").Err()
	}

	deps := godep.NewDeps(mainMod)

	// Load the existing state in the bundle, if any, to append to it.
	if bundleMod, ok := out.File(bundledGoModPath); ok {
		state := godep.SerializedState{}
		if state.GoMod, err = bundleMod.ReadAll(); err != nil {
			return nil, errors.Annotate(err, "reading %q", bundledGoModPath).Err()
		}
		modulesTxt, ok := out.File(bundledModulesTxtPath)
		if !ok {
			return nil, errors.Reason("unexpectedly missing %q", bundledModulesTxtPath).Err()
		}
		if state.ModulesTxt, err = modulesTxt.ReadAll(); err != nil {
			return nil, errors.Annotate(err, "reading %q", bundledModulesTxtPath).Err()
		}
		if err := deps.Load(state); err != nil {
			return nil, errors.Annotate(err, "loading bundle deps").Err()
		}
	}

	return deps, nil
}

// currentMode returns the bundling mode of the output.
//
// It looks at existing files in the output to decide.
func currentMode(out *fileset.Set) bundleMode {
	if _, ok := out.File(goPathRoot); ok {
		return bundleGoPath
	}
	if _, ok := out.File(goModRoot); ok {
		return bundleModules
	}
	return bundleUnknown
}

// hackStaticFiles copies static non-go files from `src` to `dst`.
//
// `src` is an absolute path to a source directory on disk. `dst` is a
// destination path relative to the fileset root.
func hackStaticFiles(ctx context.Context, out *fileset.Set, src, dst string, exclude fileset.Excluder) error {
	conflict := false
	err := out.AddFromDisk(src, dst, func(absPath string, isDir bool) bool {
		// Respect .gcloudignore.
		if exclude(absPath, isDir) {
			return true
		}
		// A path relative to the `src` e.g. "templates/index.html".
		rel, err := relPath(src, absPath)
		if err != nil {
			panic(fmt.Sprintf("impossible: %s", err))
		}
		// Keep descending at least one level deeper.
		if rel == "." {
			return false
		}
		// Skip YAMLs directly in the app directory. Most likely they are appengine
		// YAMLs and not really static files. Exposing them may cause issues.
		if filepath.Dir(rel) == "." && filepath.Ext(rel) == ".yaml" {
			return true
		}
		// Skip source code files: they aren't "static files" and we don't want to
		// confuse the go compiler with unexpected code files in weird places.
		if isGoSourceFile(rel) {
			return true
		}
		// The matching path in the output should not exist yet.
		setPath := filepath.Join(dst, rel)
		if _, ok := out.File(setPath); ok {
			logging.Errorf(ctx, "Conflict when copying static file into the module root (see b/323980048): %s", rel)
			conflict = true
			return true // skip descending this tree
		}
		if !isDir {
			logging.Infof(ctx, "Copying static file into the module root (see b/323980048): %s", rel)
		}
		return false
	})
	switch {
	case err != nil:
		return err
	case conflict:
		return errors.Reason("path conflicts, see the log").Err()
	default:
		return nil
	}
}

// envScriptGoPath is a script that modifies Go env vars to point to files
// in the tarball built for GOPATH mode. Can be used to manually test the
// tarball's soundness.
const envScriptGoPath = `#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"

echo "export GOARCH=amd64"
echo "export GOOS=linux"
echo "export GO111MODULE=off"
echo "export GOPATH=$(pwd)"
`

// envScriptModules is a script that modifies Go env vars to point to files
// in the tarball built for modules mode. Can be used to manually test the
// tarball's soundness.
const envScriptModules = `#!/usr/bin/env bash
echo "export GOARCH=amd64"
echo "export GOOS=linux"
echo "export GO111MODULE=on"
echo "unset GOPATH"
`
