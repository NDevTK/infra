// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package application

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/vpython_ng/pkg/common"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/common/system/filesystem"
	"go.chromium.org/luci/vpython"
	vpythonAPI "go.chromium.org/luci/vpython/api/vpython"
	"go.chromium.org/luci/vpython/python"
	"go.chromium.org/luci/vpython/spec"

	"github.com/mitchellh/go-homedir"
)

const (
	// VirtualEnvRootENV is an environment variable that, if set, will be used
	// as the default VirtualEnv root.
	//
	// This value overrides the default (~/.vpython-root), but can be overridden
	// by the "-vpython-root" flag.
	//
	// Like "-vpython-root", if this value is present but empty, a tempdir will be
	// used for the VirtualEnv root.
	VirtualEnvRootENV = "VPYTHON_VIRTUALENV_ROOT"

	// DefaultSpecENV is an environment variable that, if set, will be used as the
	// default VirtualEnv spec file if none is provided or found through probing.
	DefaultSpecENV = "VPYTHON_DEFAULT_SPEC"

	// LogTraceENV is an environment variable that, if set, will set the default
	// log level to Debug.
	//
	// This is useful when debugging scripts that invoke "vpython" internally,
	// where adding the "-vpython-log-level" flag is not straightforward. The
	// flag is preferred when possible.
	LogTraceENV = "VPYTHON_LOG_TRACE"

	// BypassENV is an environment variable that is used to detect if we shouldn't
	// do any vpython stuff at all, but should instead directly invoke the next
	// `python` on PATH.
	BypassENV = "VPYTHON_BYPASS"

	// BypassSentinel must be the BypassENV value (verbatim) in order to trigger
	// vpython bypass.
	BypassSentinel = "manually managed python not supported by chrome operations"
)

// Config is an application's default configuration.
type Application struct {
	// PruneThreshold, if > 0, is the maximum age of a VirtualEnv before it
	// becomes candidate for pruning. If <= 0, no pruning will be performed.
	PruneThreshold time.Duration

	// MaxPrunesPerSweep, if > 0, is the maximum number of VirtualEnv that should
	// be pruned passively. If <= 0, no limit will be applied.
	MaxPrunesPerSweep int

	// Bypass, if true, instructs vpython to completely bypass VirtualEnv
	// bootstrapping and execute with the local system interpreter.
	Bypass bool

	// Path to environment specification file to load. Default probes for one.
	SpecPath string

	// Path to default specification file to load if no specification is found.
	DefaultSpecPath string

	// Path to virtual environment root directory.
	// If explicitly set to empty string, a temporary directory will be used and
	// cleaned up on completion.
	VpythonRoot string

	// Tool mode, if it's not empty, vpython will execute the tool instead of
	// python.
	ToolMode string

	// WorkDir is the Python working directory. If empty, the current working
	// directory will be used.
	WorkDir string

	// Context used through runtime. By default it will be context.Background.
	Context context.Context

	Environments []string
	Arguments    []string

	VpythonSpec       *vpythonAPI.Spec
	PythonCommandLine *python.CommandLine
	PythonExecutable  string
	VENVPackage       cipkg.Package
}

func (a *Application) Must(err error) {
	if err == nil {
		return
	}
	a.Fatal(err)
}

func (a *Application) Fatal(err error) {
	logging.Errorf(a.Context, "fatal error: %v", err)
	os.Exit(1)
}

// Initialize logger first to make it available for all steps after.
func (a *Application) Initialize() {
	if a.Context == nil {
		a.Context = context.Background()
	}
	defaultLogLevel := logging.Error
	if os.Getenv(LogTraceENV) != "" {
		defaultLogLevel = logging.Debug
	}
	ctx := gologger.StdConfig.Use(a.Context)
	a.Context = logging.SetLevel(ctx, defaultLogLevel)
}

func (a *Application) ParseEnvs() (err error) {
	e := environ.New(a.Environments)

	// Determine our VirtualEnv base directory.
	if v, ok := e.Get(VirtualEnvRootENV); ok {
		a.VpythonRoot = v
	} else {
		hdir, err := homedir.Dir()
		if err != nil {
			return errors.Annotate(err, "failed to get user home directory").Err()
		}
		a.VpythonRoot = filepath.Join(hdir, ".vpython-ng-root")
	}

	// Get default spec path
	a.DefaultSpecPath = e.GetEmpty(DefaultSpecENV)

	// Check if it's in bypass mode
	if e.GetEmpty(BypassENV) == BypassSentinel {
		a.Bypass = true
	}
	return nil
}

func (a *Application) ParseArgs() (err error) {
	var fs flag.FlagSet
	fs.StringVar(&a.VpythonRoot, "vpython-root", a.VpythonRoot,
		"Path to virtual environment root directory. "+
			"If explicitly set to empty string, a temporary directory will be used and cleaned up "+
			"on completion.")
	fs.StringVar(&a.SpecPath, "vpython-spec", a.SpecPath,
		"Path to environment specification file to load. Default probes for one.")
	fs.StringVar(&a.ToolMode, "vpython-tool", a.ToolMode,
		"Tools for vpython command:\n"+
			"install: installs the configured virtual environment.\n"+
			"verify: verifies that a spec and its wheels are valid.")

	vpythonArgs, pythonArgs := extractFlagsForSet(a.Arguments, &fs)
	if err := fs.Parse(vpythonArgs); err != nil {
		return errors.Annotate(err, "failed to parse flags").Err()
	}

	if a.PythonCommandLine, err = python.ParseCommandLine(pythonArgs); err != nil {
		return errors.Annotate(err, "failed to parse python commandline").Err()
	}
	return nil
}

func (a *Application) LoadSpec() error {
	ctx := a.Context

	if a.SpecPath != "" {
		var sp vpythonAPI.Spec
		if err := spec.Load(a.SpecPath, &sp); err != nil {
			return err
		}
		a.VpythonSpec = sp.Clone()
		return nil
	}

	opts := vpython.Options{
		SpecLoader: spec.Loader{
			CommonFilesystemBarriers: []string{
				".gclient",
			},
			CommonSpecNames: []string{
				".vpython3",
			},
			PartnerSuffix: ".vpython3",
		},
		CommandLine: a.PythonCommandLine,
		WorkDir:     a.WorkDir,
	}

	if a.DefaultSpecPath != "" {
		if err := spec.Load(a.DefaultSpecPath, &opts.DefaultSpec); err != nil {
			return errors.Annotate(err, "failed to load default spec: %#v", a.DefaultSpecPath).Err()
		}
	}

	if opts.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Annotate(err, "failed to get working directory").Err()
		}
		opts.WorkDir = wd
	}
	if err := filesystem.AbsPath(&opts.WorkDir); err != nil {
		return errors.Annotate(err, "failed to resolve absolute path of WorkDir").Err()
	}

	if err := opts.ResolveSpec(ctx); err != nil {
		return err
	}
	a.VpythonSpec = opts.EnvConfig.Spec.Clone()
	return nil
}

func (a *Application) BuildVENV(venv cipkg.Generator) error {
	ctx := a.Context

	root := a.VpythonRoot
	if root == "" {
		tmp, err := os.MkdirTemp("", "vpython")
		if err != nil {
			return errors.Annotate(err, "failed to create temporary vpython root").Err()
		}
		root = tmp
	}

	s, err := utilities.NewLocalStorage(root)
	if err != nil {
		return errors.Annotate(err, "failed to load storage").Err()
	}

	// Generate derivations
	bctx := &cipkg.BuildContext{
		Platform: cipkg.Platform{
			Build:  fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
			Host:   fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
			Target: fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
		},
		Storage: s,
		Context: ctx,
	}

	drv, err := venv.Generate(bctx)
	if err != nil {
		return errors.Annotate(err, "failed to generate venv derivation").Err()
	}
	pkg := s.Add(drv)

	// Build derivations
	b := utilities.NewBuilder(s)
	if err := b.Add(pkg); err != nil {
		return errors.Annotate(err, "failed to add venv to builder").Err()
	}

	if err := b.BuildAll(func(p cipkg.Package) error {
		var out strings.Builder
		// We don't expect to use a temporary directory. So command is executed
		// in the output directory.
		cmd := utilities.CommandFromPackage(p)
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Dir = p.Directory()
		if err := builtins.Execute(ctx, cmd); err != nil {
			logging.Errorf(ctx, "%#v", out.String())
			return err
		}
		return nil
	}); err != nil {
		return errors.Annotate(err, "failed to build venv").Err()
	}

	// This lock is expected to be held during the lifetime of the process.
	// TODO: cpython need to be locked here
	if err := pkg.RLock(); err != nil {
		return errors.Annotate(err, "failed to acquire read lock for venv").Err()
	}
	a.VENVPackage = pkg

	a.PythonExecutable = common.Python3VENV(pkg.Directory())

	// Prune used packages
	if a.PruneThreshold > 0 {
		s.Prune(ctx, a.PruneThreshold, a.MaxPrunesPerSweep)
	}
	return nil
}

func (a *Application) ExecutePython() error {
	ctx := a.Context

	if a.Bypass && a.PythonExecutable == "" {
		var err error
		if a.PythonExecutable, err = exec.LookPath("python3"); err != nil {
			return errors.Annotate(err, "failed to find python in path").Err()
		}
	}

	env := environ.New(a.Environments)
	python.IsolateEnvironment(&env, true)

	// TODO: Pass exec.Cmd instead. exec.Cmd should includes enough information
	// for execution.
	if err := vpython.Exec(ctx, &python.Interpreter{Python: a.PythonExecutable}, a.PythonCommandLine, env, a.WorkDir, nil); err != nil {
		return errors.Annotate(err, "failed to execute python").Err()
	}
	return nil
}

func (a *Application) GetExecCommand() *exec.Cmd {
	env := environ.New(a.Environments)
	python.IsolateEnvironment(&env, false)

	cl := a.PythonCommandLine.Clone()
	cl.AddSingleFlag("s")

	return &exec.Cmd{
		Path: a.PythonExecutable,
		Args: append([]string{a.PythonExecutable}, cl.BuildArgs()...),
		Env:  env.Sorted(),
		Dir:  a.WorkDir,
	}
}
