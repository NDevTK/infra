// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package manifest defines structure of YAML files with target definitions.
package manifest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
)

// Manifest is a definition of what to build, how and where.
//
// Comments here describe the structure of the manifest file on disk. In the
// loaded form all paths use filepath.Separator as a directory separator.
type Manifest struct {
	// Name is the name of this target, required.
	//
	// When building Docker images it is an image name (without registry or any
	// tags).
	Name string `yaml:"name"`

	// ManifestDir is a directory that contains this manifest file.
	//
	// Populated when it is loaded.
	ManifestDir string `yaml:"-"`

	// Extends is a unix-style path (relative to this YAML file) to a manifest
	// used as a base.
	//
	// Optional.
	//
	// Such base manifests usually contain definitions shared by many files, such
	// as "imagepins" and "infra".
	//
	// Dicts are merged (recursively), lists are joined (base entries first).
	Extends string `yaml:"extends,omitempty"`

	// Dockerfile is a unix-style path to the image's Dockerfile, relative to this
	// YAML file.
	//
	// If unset, but there exists `${contextdir}/Dockerfile`, it will be used
	// instead (similarly to how Docker discovers Dockerfile).
	//
	// All images referenced in a Dockerfile are resolved into concrete digests
	// via an external file. See ImagePins field for more information.
	Dockerfile string `yaml:"dockerfile,omitempty"`

	// ContextDir is a unix-style path to the directory to use as a basis for
	// the build. The path is relative to this YAML file.
	//
	// All non-gitignored files there end up available to the remote builder
	// (e.g. a docker daemon will see this directory as a context directory when
	// building the image).
	//
	// All symlinks there are resolved to their targets. Only +w and +x file mode
	// bits are preserved (all files have 0444 mode by default, +w adds additional
	// 0200 bit and +x adds additional 0111 bis). All other file metadata (owners,
	// setuid bits, modification times) are ignored.
	//
	// The default value depends on whether Dockerfile is set. If it is, then
	// ContextDir defaults to the directory with Dockerfile. Otherwise the context
	// directory is assumed to be empty.
	ContextDir string `yaml:"contextdir,omitempty"`

	// InputsDir is an optional directory that can be used to reference files
	// consumed by build steps (as "${inputsdir}/path").
	//
	// Unlike ContextDir, its full content does not automatically end up in the
	// output.
	//
	// If unset, defaults to ContextDir.
	InputsDir string `yaml:"inputsdir,omitempty"`

	// ImagePins is a unix-style path to the YAML file with pre-resolved mapping
	// from (docker image, tag) pair to the corresponding docker image digest.
	//
	// The path is relative to the manifest YAML file. It should point to a YAML
	// file with the following structure:
	//
	//    pins:
	//      - image: <img>
	//        tag: <tag>
	//        digest: sha256:<sha256>
	//      - image: <img>
	//        tag: <tag>
	//        digest: sha256:<sha256>
	//      ...
	//
	// See dockerfile.Pins struct for more details.
	//
	// This file will be used to rewrite the input Dockerfile to reference all
	// images (in "FROM ..." lines) only by their digests. This is useful for
	// reproducibility of builds.
	//
	// Only following forms of "FROM ..." statement are allowed:
	//  * FROM <image> [AS <name>] (assumes "latest" tag)
	//  * FROM <image>[:<tag>] [AS <name>] (resolves the given tag)
	//  * FROM <image>[@<digest>] [AS <name>] (passes the definition through)
	//
	// In particular ARGs in FROM line (e.g. "FROM base:${CODE_VERSION}") are
	// not supported.
	//
	// If not set, the Dockerfile must use only digests to begin with, i.e.
	// all FROM statements should have form "FROM <image>@<digest>".
	//
	// Ignored if Dockerfile field is not set.
	ImagePins string `yaml:"imagepins,omitempty"`

	// Deterministic is true if Dockerfile (with all "FROM" lines resolved) can be
	// understood as a pure function of inputs in ContextDir, i.e. it does not
	// depend on the state of the world.
	//
	// Examples of things that make Dockerfile NOT deterministic:
	//   * Using "apt-get" or any other remote calls to non-pinned resources.
	//   * Cloning repositories from "master" ref (or similar).
	//   * Fetching external resources using curl or wget.
	//
	// When building an image marked as deterministic, the builder will calculate
	// a hash of all inputs (including resolve Dockerfile itself) and check
	// whether there's already an image built from them. If there is, the build
	// will be skipped completely and the existing image reused.
	//
	// Images marked as non-deterministic are always rebuilt and reuploaded, even
	// if nothing in ContextDir has changed.
	Deterministic *bool `yaml:"deterministic,omitempty"`

	// Sources is unix-style paths to the directories that contain the source code
	// used to build the artifact, for the express purpose of getting its revision
	// and propagating as artifact's metadata.
	//
	// Values of this field are not involved in the build process itself at all.
	// If omitted, defaults to ${inputsdir} or ${contextdir} whichever is set.
	//
	// The paths are relative to this YAML file.
	Sources []string `yaml:"sources"`

	// Infra is configuration of the build infrastructure to use: Google Storage
	// bucket, Cloud Build project, etc.
	//
	// Keys are names of presets (like "dev", "prod"). What preset is used is
	// controlled via "-infra" command line flag (defaults to "dev").
	Infra map[string]Infra `yaml:"infra"`

	// CloudBuild specifies what Cloud Build configuration to use.
	//
	// All available configurations are defined in `cloudbuild` section of `infra`
	// section. This field allows to pick the active one of this target.
	CloudBuild CloudBuildConfig `yaml:"cloudbuild,omitempty"`

	// Build defines a series of local build steps.
	//
	// Each step may add more files to the context directory. The actual
	// `contextdir` directory on disk won't be modified. Files produced here are
	// stored in a temp directory and the final context directory is constructed
	// from the full recursive copy of `contextdir` and files emitted here.
	Build []*BuildStep `yaml:"build,omitempty"`
}

// Infra contains configuration of build infrastructure to use: Google Storage
// bucket, Cloud Build project, etc.
//
// Note: when adding new fields here, check if they need to have matching
// restrictions in restrictions.go.
type Infra struct {
	// Storage specifies Google Storage location to store *.tar.gz tarballs
	// produced after executing all local build steps.
	//
	// Expected format is "gs://<bucket>/<prefix>". Tarballs will be stored as
	// "gs://<bucket>/<prefix>/<name>/<sha256>.tar.gz", where <name> comes from
	// the manifest and <sha256> is a hex sha256 digest of the tarball.
	//
	// The bucket should exist already. Its contents is trusted, i.e. if there's
	// an object with desired <sha256>.tar.gz there already, it won't be replaced.
	//
	// Required when using Cloud Build.
	Storage string `yaml:"storage"`

	// Registry is a Cloud Registry to push images to e.g. "gcr.io/something".
	//
	// Required when using Cloud Build.
	Registry string `yaml:"registry"`

	// CloudBuild contains configuration presets of Cloud Build infrastructure.
	//
	// Each entry defines a named Cloud Build configuration that can be referenced
	// in individual manifests via `builder` field of `cloudbuild` section.
	CloudBuild map[string]CloudBuildBuilder `yaml:"cloudbuild"`

	// Notify indicates what downstream services to notify once the image is
	// built.
	//
	// It is not interpreted by cloudbuildhelper itself (just validated), and
	// passed to the JSON output of `build` and `upload` commands. Callers
	// (usually the recipes) know the meaning of this field and implement the
	// actual notification logic.
	Notify []NotifyConfig `yaml:"notify"`
}

// NotifyConfig is a single item in `notify` list.
//
// It is read from the YAML manifest and ends up in the -json-output.
type NotifyConfig struct {
	// Kind indicates a kind of service to notify.
	//
	// The only supported value now is "git", meaning to checkout a git repo and
	// invoke a script there with results of the build. Note that the actual
	// logic to do that is implemented in recipes that call cloudbuildhelper.
	Kind string `yaml:"kind" json:"kind"`

	// Repo is a git repo to checkout (as "https://..." URL).
	//
	// Effective only for "git" notifiers.
	Repo string `yaml:"repo" json:"repo"`

	// Script is a path to the script inside the repo to invoke.
	//
	// Effective only for "git" notifiers.
	Script string `yaml:"script" json:"script"`
}

// DestinationID identifies the destination of the notification for the purpose
// of checking it against -restrict-notifications flags in restrictions.go.
func (n *NotifyConfig) DestinationID() string {
	switch n.Kind {
	case "git":
		return fmt.Sprintf("git:%s/%s", n.Repo, n.Script)
	default:
		return "unknown"
	}
}

// rebaseOnTop implements "extends" logic.
func (i *Infra) rebaseOnTop(b Infra) {
	setIfEmpty(&i.Storage, b.Storage)
	setIfEmpty(&i.Registry, b.Registry)

	// We support only adding CloudBuild configs, not overriding existing ones.
	// Copy all entries in 'b' that are not in 'i'.
	for k, v := range b.CloudBuild {
		if _, ok := i.CloudBuild[k]; !ok {
			if i.CloudBuild == nil {
				i.CloudBuild = make(map[string]CloudBuildBuilder, 1)
			}
			i.CloudBuild[k] = v
		}
	}

	if len(b.Notify) != 0 {
		i.Notify = append([]NotifyConfig(nil), b.Notify...)
	}
}

// CloudBuildBuilder contains a configuration of Cloud Build infrastructure.
//
// It has a name. Individual targets can specify what configuration they want
// to use via `builder` field of `cloudbuild` section.
type CloudBuildBuilder struct {
	// Project is Google Cloud Project name that hosts the Cloud Build instance.
	Project string `yaml:"project"`

	// Pool is a private worker pool to run builds on (if any).
	Pool *WorkerPool `yaml:"pool,omitempty"`

	// Executable is a name of the container to execute as Cloud Build step.
	//
	// May include ":<tag>" suffix. Its working directory will be "/workspace" and
	// it will be populated with files from the context directory, as prepared by
	// the cloudbuildhelper.
	Executable string `yaml:"executable"`

	// Args is command line arguments to pass to the executable.
	//
	// Following tokens will be substituted by the cloudbuildhelper before passing
	// the command to Cloud Build:
	//   * ${CBH_DOCKER_IMAGE}: the full image name (including the registry prefix
	//     and ":<tag>" suffix) that the Cloud Build step should build.
	//   * ${CBH_DOCKER_LABELS}: list of ["--label", "k=v", "--label", "k=v", ...]
	//     pairs with all labels to pass to `docker build`. Can be passed directly
	//     to `docker build`.
	Args []string `yaml:"args"`

	// PushesExplicitly, if true, indicates the step pushes the final image to
	// the registry itself, instead of relying on Cloud Build to do the push.
	//
	// Note that PushesExplicitly == true somewhat reduces security properties of
	// the build, since the pushed image can theoretically be replaced with
	// something else between the moment it was pushed by the builder and
	// cloudbuildhelper resolves its tag into a SHA256 digest. When pushing via
	// Cloud Build (i.e. PushesExplicitly == false) Cloud Build reports the image
	// SHA256 digest directly (based on its local docker cache) and the tag is
	// not used at all.
	PushesExplicitly bool `yaml:"pushes_explicitly"`

	// Timeout is how long the Cloud Build build is allowed to run.
	//
	// Default timeout is ten minutes. Specified as a duration in seconds,
	// terminated by 's'. Example: "3.5s".
	Timeout string `yaml:"timeout"`
}

// WorkerPool is an ID of a worker pool within a project.
type WorkerPool struct {
	// Region is a Cloud Region that hosts this pool.
	Region string `yaml:"region"`
	// ID is the short pool ID.
	ID string `yaml:"id"`
}

// CloudBuildConfig contains target-specific Cloud Build configuration.
//
// It is used to select/tweak a named configuration preset specified in
// `cloudbuild` field of `infra` section.
type CloudBuildConfig struct {
	Builder string `yaml:"builder"` // what named configuration to use
}

// rebaseOnTop implements "extends" logic.
func (c *CloudBuildConfig) rebaseOnTop(b CloudBuildConfig) {
	setIfEmpty(&c.Builder, b.Builder)
}

// ResolveCloudBuildConfig combines CloudBuildConfig specified in the target
// manifest with the corresponding CloudBuildBuilder in this infra section
// and returns the resolved validated configuration to use when calling
// Cloud Build.
func (i *Infra) ResolveCloudBuildConfig(cfg CloudBuildConfig) (*CloudBuildBuilder, error) {
	if cfg.Builder == "" {
		return nil, errors.Reason("cloudbuild.builder is required when using Cloud Build").Err()
	}
	builder, ok := i.CloudBuild[cfg.Builder]
	if !ok {
		return nil, errors.Reason("cloudbuild.builder references unknown Cloud Build builder %q", cfg.Builder).Err()
	}

	if builder.Project == "" {
		return nil, errors.Reason("infra[...].cloudbuild[...].project is required when using Cloud Build").Err()
	}

	// By default do regular "docker build" calls.
	if builder.Executable == "" {
		builder.Executable = "gcr.io/cloud-builders/docker:latest"
	}
	if len(builder.Args) == 0 {
		builder.Args = []string{
			"build",
			".",
			"--network", "cloudbuild", // this is what "gcloud build submit" uses, it is documented
			"--no-cache", // state of the cache on Cloud Build workers is not well defined
			"--tag", "${CBH_DOCKER_IMAGE}",
			"${CBH_DOCKER_LABELS}",
		}
	}

	if builder.Timeout == "" {
		builder.Timeout = "600s"
	}

	return &builder, nil
}

// BuildStep is one local build operation.
//
// It takes a local checkout and produces one or more output files put into
// the context directory.
//
// This struct is a "case class" with union of all supported build step kinds.
// The chosen "case" is returned by Concrete() method.
type BuildStep struct {
	// Fields common to two or more build step kinds.

	// Dest specifies a location to put the result into.
	//
	// Applies to `copy`, `go_build` and `go_gae_bundle` steps.
	//
	// Usually prefixed with "${contextdir}/" to indicate it is relative to
	// the context directory.
	//
	// Optional in the original YAML, always populated after Manifest is parsed.
	// See individual *BuildStep structs for defaults.
	Dest string `yaml:"dest,omitempty"`

	// Cwd is a working directory to run the command in.
	//
	// Applies to `run` and `go_build` steps.
	//
	// Default is ${inputsdir}, ${contextdir} or ${manifestdir}, whichever is set.
	Cwd string `yaml:"cwd,omitempty"`

	// Disjoint set of possible build kinds.
	//
	// To add a new step kind:
	//   1. Add a new embedded struct here with definition of the step.
	//   2. Add methods to implement ConcreteBuildStep.
	//   3. Add one more entry to a slice in wireStep(...) below.
	//   4. Add the actual step implementation to builder/step*.go.
	//   5. Add one more type switch to Builder.Build() in builder/builder.go.

	CopyBuildStep        `yaml:",inline"` // copy a file or directory into the output
	GoBuildStep          `yaml:",inline"` // build go binary using "go build"
	RunBuildStep         `yaml:",inline"` // run a command that modifies the checkout
	GoGAEBundleBuildStep `yaml:",inline"` // bundle Go source code for GAE

	manifest *Manifest         // the manifest that defined this step
	index    int               // zero-based index of the step in its parent manifest
	concrete ConcreteBuildStep // pointer to one of *BuildStep above
}

// ConcreteBuildStep is implemented by various *BuildStep structs.
type ConcreteBuildStep interface {
	Kind() string   // used by -restrict-build-steps
	String() string // used for human logs only, doesn't have to encode all details

	isEmpty() bool                                        // true if the struct is not populated
	initStep(bs *BuildStep, dirs map[string]string) error // populates 'bs' and self
}

// Concrete returns a pointer to some concrete populated *BuildStep.
func (bs *BuildStep) Concrete() ConcreteBuildStep { return bs.concrete }

// CopyBuildStep indicates we want to copy a file or directory.
//
// Doesn't materialize copies on disk, just puts them directly into the output
// file set.
type CopyBuildStep struct {
	// Copy is a path to copy files from.
	//
	// Should start with either "${contextdir}/", "${inputsdir}/" or
	// "${manifestdir}/" to indicate the root path.
	//
	// Can either be a directory or a file. Whatever it is, it will be put into
	// the output as Dest. By default Dest is "${contextdir}/<basename of Copy>"
	// (i.e. we copy Copy into the root of the context dir).
	Copy string `yaml:"copy,omitempty"`

	// Ignore is a list of patterns of files and directories to *not* copy.
	//
	// The format of patterns is the same as used in .gitignore.
	Ignore []string `yaml:"ignore,omitempty"`
}

func (s *CopyBuildStep) Kind() string { return "copy" }

func (s *CopyBuildStep) String() string { return fmt.Sprintf("copy %q", s.Copy) }

func (s *CopyBuildStep) isEmpty() bool { return s.Copy == "" && len(s.Ignore) == 0 }

func (s *CopyBuildStep) initStep(bs *BuildStep, dirs map[string]string) (err error) {
	if s.Copy, err = renderPath("copy", s.Copy, dirs); err != nil {
		return err
	}
	if bs.Dest == "" {
		bs.Dest = "${contextdir}/" + filepath.Base(s.Copy)
	}
	if bs.Dest, err = renderPath("dest", bs.Dest, dirs); err != nil {
		return err
	}
	return
}

// GoBuildStep indicates we want to build a go command binary.
//
// Doesn't materialize the build output on disk, just puts it directly into the
// output file set.
type GoBuildStep struct {
	// GoBinary specifies a go command binary to build.
	//
	// This is a path (relative to GOPATH) to some 'main' package. It will be
	// built roughly as:
	//
	//  $ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build <go_binary> -o <dest>
	//
	// Where <dest> is taken from Dest and it must be under the context directory.
	// It is set to "${contextdir}/<go package name>" by default.
	GoBinary string `yaml:"go_binary,omitempty"`
}

func (s *GoBuildStep) Kind() string { return "go_binary" }

func (s *GoBuildStep) String() string { return fmt.Sprintf("go build %q", s.GoBinary) }

func (s *GoBuildStep) isEmpty() bool { return s.GoBinary == "" }

func (s *GoBuildStep) initStep(bs *BuildStep, dirs map[string]string) (err error) {
	if err = renderCwd(&bs.Cwd, dirs); err != nil {
		return
	}
	if bs.Dest == "" {
		bs.Dest = "${contextdir}/" + path.Base(s.GoBinary)
	}
	bs.Dest, err = renderPath("dest", bs.Dest, dirs)
	return
}

// RunBuildStep indicates we want to run some arbitrary command.
//
// The command may modify the checkout or populate the context dir.
type RunBuildStep struct {
	// Run indicates a command to run along with all its arguments.
	//
	// Strings that start with "${contextdir}/", "${inputsdir}/" or
	// "${manifestdir}/" will be rendered as absolute paths.
	Run []string `yaml:"run,omitempty"`

	// Outputs is a list of files or directories to put into the output.
	//
	// They are something that `run` should be generating.
	//
	// They are expected to be under "${contextdir}". A single output entry
	// "${contextdir}/generated/file" is equivalent to a copy step that "picks up"
	// the generated file:
	//   - copy: ${contextdir}/generated/file
	//     dest: ${contextdir}/generated/file
	//
	// If outputs are generated outside of the context directory, use `copy` steps
	// explicitly.
	Outputs []string
}

func (s *RunBuildStep) Kind() string { return "run" }

func (s *RunBuildStep) String() string { return fmt.Sprintf("run %q", s.Run) }

func (s *RunBuildStep) isEmpty() bool { return len(s.Run) == 0 && len(s.Outputs) == 0 }

func (s *RunBuildStep) initStep(bs *BuildStep, dirs map[string]string) (err error) {
	if len(s.Run) == 0 {
		return errors.Reason("bad `run` value: must not be empty").Err()
	}

	for i, val := range s.Run {
		if isTemplatedPath(val) {
			rel, err := renderPath(fmt.Sprintf("run[%d]", i), val, dirs)
			if err != nil {
				return err
			}
			// We are going to pass these arguments to a command with different cwd,
			// need to make sure they are absolute.
			if s.Run[i], err = filepath.Abs(rel); err != nil {
				return errors.Annotate(err, "bad `run[%d]` %q", i, rel).Err()
			}
		}
	}

	if err = renderCwd(&bs.Cwd, dirs); err != nil {
		return err
	}

	for i, out := range s.Outputs {
		if s.Outputs[i], err = renderPath(fmt.Sprintf("output[%d]", i), out, dirs); err != nil {
			return err
		}
	}

	return
}

// GoGAEBundleBuildStep can be used to prepare a tarball with Go GAE app source.
//
// Given a path to a GAE module yaml (that should reside in a directory with
// some `main` go package), it:
//   - Copies all files in the modules directory (including all non-go files)
//     to `_gopath/src/<its import path>`
//   - Copies all *.go code with transitive dependencies to `_gopath/src/`.
//   - Makes `Dest` a symlink pointing to `_gopath/src/<import path>`.
//
// This ensures "gcloud app deploy" eventually can upload all *.go files needed
// to deploy a module.
type GoGAEBundleBuildStep struct {
	// GoGAEBundle is path to GAE module YAML.
	GoGAEBundle string `yaml:"go_gae_bundle,omitempty"`
}

func (s *GoGAEBundleBuildStep) Kind() string { return "go_gae_bundle" }

func (s *GoGAEBundleBuildStep) String() string { return fmt.Sprintf("go gae bundle %q", s.GoGAEBundle) }

func (s *GoGAEBundleBuildStep) isEmpty() bool { return s.GoGAEBundle == "" }

func (s *GoGAEBundleBuildStep) initStep(bs *BuildStep, dirs map[string]string) (err error) {
	if s.GoGAEBundle, err = renderPath("go_gae_bundle", s.GoGAEBundle, dirs); err != nil {
		return
	}
	bs.Dest, err = renderPath("dest", bs.Dest, dirs)
	return
}

// Load loads the manifest from the given path, traversing all "extends" links.
//
// After the manifest is loaded, its fields (like ContextDir) can be manipulated
// (e.g. to set defaults), after which all "${dir}/" references in build steps
// must be resolved by a call to Finalize.
func Load(path string) (*Manifest, error) {
	return loadRecursive(path, 0)
}

// parse reads the manifest and populates paths there.
//
// If cwd is not empty, rebases all relative paths in it on top of it.
//
// Does not traverse "extends" links.
func parse(r io.Reader, cwd string) (*Manifest, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Annotate(err, "failed to read the manifest body").Err()
	}
	out := Manifest{}
	if err = yaml.Unmarshal(body, &out); err != nil {
		return nil, errors.Annotate(err, "failed to parse the manifest").Err()
	}
	if err := out.initBase(cwd); err != nil {
		return nil, err
	}
	return &out, nil
}

// loadRecursive implements Load by tracking how deep we go as a simple
// protection against recursive "extends" links.
func loadRecursive(path string, fileCount int) (*Manifest, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, errors.Annotate(err, "when opening manifest file").Err()
	}
	defer r.Close()

	m, err := parse(r, filepath.Dir(path))
	switch {
	case err != nil:
		return nil, errors.Annotate(err, "when parsing %q", path).Err()
	case m.Extends == "":
		return m, nil
	case fileCount > 10:
		return nil, errors.Reason("too much nesting").Err()
	}

	base, err := loadRecursive(m.Extends, fileCount+1)
	if err != nil {
		return nil, errors.Annotate(err, "when loading %q", path).Err()
	}
	m.rebaseOnTop(base)
	return m, nil
}

// initBase initializes pointers in steps and rebases paths.
//
// Doesn't yet touch actual bodies of steps, they will be initialized later
// when the whole manifest tree is loaded, see Finalize.
func (m *Manifest) initBase(cwd string) error {
	if err := validateName(m.Name); err != nil {
		return errors.Annotate(err, `bad "name" field`).Err()
	}
	m.ManifestDir = cwd
	normPath(&m.Extends, cwd)
	normPath(&m.Dockerfile, cwd)
	normPath(&m.ContextDir, cwd)
	normPath(&m.InputsDir, cwd)
	normPath(&m.ImagePins, cwd)
	if m.ContextDir == "" && m.Dockerfile != "" {
		m.ContextDir = filepath.Dir(m.Dockerfile)
	}
	for i := range m.Sources {
		normPath(&m.Sources[i], cwd)
	}
	for k, v := range m.Infra {
		if err := validateInfra(v); err != nil {
			return errors.Annotate(err, "in infra section %q", k).Err()
		}
	}
	for i, b := range m.Build {
		if err := wireStep(b, m, i); err != nil {
			return errors.Annotate(err, "bad build step #%d", i+1).Err()
		}
	}
	return nil
}

// Finalize replaces "${dir}/" in paths in steps with actual values and fills
// in defaults.
func (m *Manifest) Finalize() error {
	if m.InputsDir == "" {
		m.InputsDir = m.ContextDir
	}
	if len(m.Sources) == 0 {
		switch {
		case m.InputsDir != "":
			m.Sources = []string{m.InputsDir}
		case m.ContextDir != "":
			m.Sources = []string{m.ContextDir}
		}
	}
	for _, b := range m.Build {
		dirs := map[string]string{
			"contextdir":  m.ContextDir,
			"inputsdir":   m.InputsDir,
			"manifestdir": b.manifest.ManifestDir,
		}
		if err := b.concrete.initStep(b, dirs); err != nil {
			return errors.Annotate(err, "bad build step #%d in %q", b.index+1, b.manifest.ManifestDir).Err()
		}
	}
	return nil
}

// rebaseOnTop implements "extends" logic.
func (m *Manifest) rebaseOnTop(b *Manifest) {
	m.Extends = "" // resolved now

	setIfEmpty(&m.Dockerfile, b.Dockerfile)
	setIfEmpty(&m.ContextDir, b.ContextDir)
	setIfEmpty(&m.InputsDir, b.InputsDir)
	setIfEmpty(&m.ImagePins, b.ImagePins)
	if m.Deterministic == nil && b.Deterministic != nil {
		cpy := *b.Deterministic
		m.Deterministic = &cpy
	}

	// Sources are joined as sets, but preserving order.
	m.Sources = joinStringSets(m.Sources, b.Sources)

	// Rebase all entries already present in 'm' on top of entries in 'b'.
	for k, v := range m.Infra {
		if base, ok := b.Infra[k]; ok {
			v.rebaseOnTop(base)
			m.Infra[k] = v
		}
	}
	// Copy all entries in 'b' that are not in 'm'.
	for k, v := range b.Infra {
		if _, ok := m.Infra[k]; !ok {
			if m.Infra == nil {
				m.Infra = make(map[string]Infra, 1)
			}
			m.Infra[k] = v
		}
	}

	m.CloudBuild.rebaseOnTop(b.CloudBuild)

	// Steps are just joined (base ones first).
	m.Build = append(b.Build, m.Build...)
}

func setIfEmpty(a *string, b string) {
	if *a == "" {
		*a = b
	}
}

func joinStringSets(a, b []string) []string {
	out := append([]string(nil), a...)
	seen := stringset.NewFromSlice(a...)
	for _, s := range b {
		if seen.Add(s) {
			out = append(out, s)
		}
	}
	return out
}

// validateName validates "name" field in the manifest.
func validateName(t string) error {
	const forbidden = "\\:@"
	switch {
	case t == "":
		return errors.Reason("can't be empty, it's required").Err()
	case strings.ContainsAny(t, forbidden):
		return errors.Reason("%q contains forbidden symbols (any of %q)", t, forbidden).Err()
	default:
		return nil
	}
}

func validateInfra(i Infra) error {
	if i.Storage != "" {
		url, err := url.Parse(i.Storage)
		if err != nil {
			return errors.Annotate(err, "bad storage %q", i.Storage).Err()
		}
		switch {
		case url.Scheme != "gs":
			return errors.Reason("bad storage %q, only gs:// is supported currently", i.Storage).Err()
		case url.Host == "":
			return errors.Reason("bad storage %q, bucket name is missing", i.Storage).Err()
		}
	}

	for idx, notify := range i.Notify {
		if err := validateNotify(notify); err != nil {
			return errors.Annotate(err, "bad notify config #%d", idx+1).Err()
		}
	}

	return nil
}

func validateNotify(n NotifyConfig) error {
	if n.Kind != "git" {
		return errors.Reason("unsupported notify kind %q", n.Kind).Err()
	}
	if !strings.HasPrefix(n.Repo, "https://") {
		return errors.Reason("`repo` should be an https:// URL, not %q", n.Repo).Err()
	}
	if path.Clean(filepath.ToSlash(n.Script)) != n.Script {
		return errors.Reason("bad `script` %q, should be a normalized slash-separate path", n.Script).Err()
	}
	if n.Script == "." || n.Script == ".." ||
		strings.HasPrefix(n.Script, "../") || strings.HasPrefix(n.Script, "/") {
		return errors.Reason("bad `script` %q, not a path inside the repo", n.Script).Err()
	}
	return nil
}

// wireStep initializes `concrete` and `manifest` pointers in the step.
//
// Doesn't touch any other fields.
func wireStep(bs *BuildStep, m *Manifest, index int) error {
	set := make([]ConcreteBuildStep, 0, 1)
	for _, s := range []ConcreteBuildStep{
		&bs.CopyBuildStep,
		&bs.GoBuildStep,
		&bs.RunBuildStep,
		&bs.GoGAEBundleBuildStep,
	} {
		if !s.isEmpty() {
			set = append(set, s)
		}
	}
	// One and only one substruct should be populated.
	switch {
	case len(set) == 0:
		return errors.Reason("unrecognized or empty").Err()
	case len(set) > 1:
		return errors.Reason("ambiguous").Err()
	default:
		bs.manifest = m
		bs.index = index
		bs.concrete = set[0]
		return nil
	}
}

func normPath(p *string, cwd string) {
	if *p != "" {
		*p = filepath.FromSlash(*p)
		if !filepath.IsAbs(*p) && cwd != "" {
			*p = filepath.Join(cwd, *p)
		}
	}
}

// isTemplatedPath is true if 'p' starts with "${<something>}[/]".
func isTemplatedPath(p string) bool {
	parts := strings.SplitN(p, "/", 2)
	return strings.HasPrefix(parts[0], "${") && strings.HasSuffix(parts[0], "}")
}

// renderPath verifies `p` starts with "${<something>}[/]", replaces it with
// dirs[<something>], and normalizes the result.
func renderPath(title, p string, dirs map[string]string) (string, error) {
	if p == "" {
		return "", errors.Reason("bad `%s`: must not be empty", title).Err()
	}

	// Helper for error messages.
	keys := func() string {
		ks := make([]string, 0, len(dirs))
		for k := range dirs {
			ks = append(ks, fmt.Sprintf("${%s}", k))
		}
		sort.Strings(ks)
		return strings.Join(ks, " or ")
	}

	parts := strings.SplitN(p, "/", 2)
	if !strings.HasPrefix(parts[0], "${") || !strings.HasSuffix(parts[0], "}") {
		return "", errors.Reason("bad `%s`: must start with %s", title, keys()).Err()
	}

	switch val, ok := dirs[strings.TrimSuffix(strings.TrimPrefix(parts[0], "${"), "}")]; {
	case !ok:
		return "", errors.Reason("bad `%s`: unknown dir variable %s, expecting %s", title, parts[0], keys()).Err()
	case val == "":
		return "", errors.Reason("bad `%s`: dir variable %s it not set", title, parts[0]).Err()
	case len(parts) == 1:
		return val, nil
	default:
		return filepath.Join(val, filepath.FromSlash(parts[1])), nil
	}
}

// renderCwd renders `cwd` to be an absolute path.
func renderCwd(cwd *string, dirs map[string]string) error {
	if *cwd == "" {
		switch {
		case dirs["inputsdir"] != "":
			*cwd = "${inputsdir}"
		case dirs["contextdir"] != "":
			*cwd = "${contextdir}"
		case dirs["manifestdir"] != "":
			*cwd = "${manifestdir}"
		}
	}
	var err error
	*cwd, err = renderPath("cwd", *cwd, dirs)
	return err
}
