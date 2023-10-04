// Copyright 2022 The Chromium Authors
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
	"strings"
	"time"

	"infra/tools/vpython_ng/pkg/common"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/common/system/filesystem"
	"go.chromium.org/luci/vpython"
	vpythonAPI "go.chromium.org/luci/vpython/api/vpython"
	"go.chromium.org/luci/vpython/python"
	"go.chromium.org/luci/vpython/spec"
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

	// InterpreterENV is an environment variable that override the default
	// searching behaviour for the bundled interpreter. It should only be used
	// for testing and debugging purpose.
	InterpreterENV = "VPYTHON_INTERPRETER"

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

	// Loglevel is used to configure the default logger set in the context.
	LogLevel logging.Level

	// Help, if true, displays the usage from both vpython and python
	Help  bool
	Usage string

	// Path to environment specification file to load. Default probes for one.
	SpecPath string

	// Path to default specification file to load if no specification is found.
	DefaultSpecPath string

	// Pattern of default specification file. If empty, uses .vpython3.
	DefaultSpecPattern string

	// Path to virtual environment root directory.
	// If explicitly set to empty string, a temporary directory will be used and
	// cleaned up on completion.
	VpythonRoot string

	// Path to cipd cache directory.
	CIPDCacheDir string

	// Tool mode, if it's not empty, vpython will execute the tool instead of
	// python.
	ToolMode string

	// WorkDir is the Python working directory. If empty, the current working
	// directory will be used.
	WorkDir string

	// InterpreterPath is the path to the python interpreter cipd package. If
	// empty, uses the bundled python from paths relative to the vpython binary.
	InterpreterPath string

	Environments []string
	Arguments    []string

	VpythonSpec       *vpythonAPI.Spec
	PythonCommandLine *python.CommandLine
	PythonExecutable  string

	// Close() is usually unnecessary since resources will be released after
	// process exited. However we need to release them manually in the tests.
	Close func()
}

// Initialize logger first to make it available for all steps after.
func (a *Application) Initialize(ctx context.Context) context.Context {
	a.LogLevel = logging.Error
	if os.Getenv(LogTraceENV) != "" {
		a.LogLevel = logging.Debug
	}
	a.Close = func() {}

	ctx = gologger.StdConfig.Use(ctx)
	return logging.SetLevel(ctx, a.LogLevel)
}

func (a *Application) SetLogLevel(ctx context.Context) context.Context {
	return logging.SetLevel(ctx, a.LogLevel)
}

func (a *Application) ParseEnvs(ctx context.Context) (err error) {
	e := environ.New(a.Environments)

	// Determine our VirtualEnv base directory.
	if v, ok := e.Lookup(VirtualEnvRootENV); ok {
		a.VpythonRoot = v
	} else {
		cdir, err := os.UserCacheDir()
		if err != nil {
			return errors.Annotate(err, "failed to get user home directory").Err()
		}
		a.VpythonRoot = filepath.Join(cdir, ".vpython-root")
	}

	// Get default spec path
	a.DefaultSpecPath = e.Get(DefaultSpecENV)

	// Get interpreter path
	if p := e.Get(InterpreterENV); p != "" {
		p, err = filepath.Abs(p)
		if err != nil {
			return err
		}
		a.InterpreterPath = p
	}

	// Check if it's in bypass mode
	if e.Get(BypassENV) == BypassSentinel {
		a.Bypass = true
	}

	// Get CIPD cache directory
	a.CIPDCacheDir = e.Get(cipd.EnvCacheDir)

	return nil
}

func (a *Application) ParseArgs(ctx context.Context) (err error) {
	var fs flag.FlagSet
	fs.BoolVar(&a.Help, "help", a.Help,
		"Display help for 'vpython' top-level arguments.")
	fs.BoolVar(&a.Help, "h", a.Help,
		"Display help for 'vpython' top-level arguments (same as -help).")

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

	fs.Var(&a.LogLevel, "vpython-log-level",
		"The logging level. Valid options are: debug, info, warning, error.")

	vpythonArgs, pythonArgs, err := extractFlagsForSet("vpython-", a.Arguments, &fs)
	if err != nil {
		return errors.Annotate(err, "failed to extract flags").Err()
	}
	if err := fs.Parse(vpythonArgs); err != nil {
		return errors.Annotate(err, "failed to parse flags").Err()
	}

	// Set CIPD CacheDIR
	if a.CIPDCacheDir == "" {
		a.CIPDCacheDir = filepath.Join(a.VpythonRoot, "cipd")
	}

	if a.PythonCommandLine, err = python.ParseCommandLine(pythonArgs); err != nil {
		return errors.Annotate(err, "failed to parse python commandline").Err()
	}

	if a.Help {
		var usage strings.Builder
		fmt.Fprintln(&usage, "Usage of vpython:")
		fs.SetOutput(&usage)
		fs.PrintDefaults()
		a.Usage = usage.String()

		a.PythonCommandLine = &python.CommandLine{
			Target: python.NoTarget{},
		}
		a.PythonCommandLine.AddSingleFlag("h")
	}
	return nil
}

func (a *Application) LoadSpec(ctx context.Context) error {
	if a.SpecPath != "" {
		var sp vpythonAPI.Spec
		if err := spec.Load(a.SpecPath, &sp); err != nil {
			return err
		}
		a.VpythonSpec = sp.Clone()
		return nil
	}

	specPattern := a.DefaultSpecPattern
	if specPattern == "" {
		specPattern = ".vpython3"
	}

	opts := vpython.Options{
		SpecLoader: spec.Loader{
			CommonFilesystemBarriers: []string{
				".gclient",
			},
			CommonSpecNames: []string{
				specPattern,
			},
			PartnerSuffix: specPattern,
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

func (a *Application) BuildVENV(ctx context.Context, ap *actions.ActionProcessor, venv generators.Generator) error {
	root := a.VpythonRoot
	if root == "" {
		tmp, err := os.MkdirTemp("", "vpython")
		if err != nil {
			return errors.Annotate(err, "failed to create temporary vpython root").Err()
		}
		root = tmp
	}

	pm, err := workflow.NewLocalPackageManager(filepath.Join(root, "store"))
	if err != nil {
		return errors.Annotate(err, "failed to load storage").Err()
	}

	// Generate derivations
	curPlat := generators.CurrentPlatform()
	plats := generators.Platforms{
		Build:  curPlat,
		Host:   curPlat,
		Target: curPlat,
	}

	b := workflow.NewBuilder(plats, pm, ap)
	pkg, err := b.Build(ctx, "", venv)
	if err != nil {
		return errors.Annotate(err, "failed to generate venv derivation").Err()
	}
	workflow.MustIncRefRecursiveRuntime(pkg)
	a.Close = func() {
		workflow.MustDecRefRecursiveRuntime(pkg)
	}

	// Prune used packages
	if a.PruneThreshold > 0 {
		pm.Prune(ctx, a.PruneThreshold, a.MaxPrunesPerSweep)
	}

	a.PythonExecutable = common.PythonVENV(pkg.Handler.OutputDirectory(), a.PythonExecutable)
	return nil
}

func (a *Application) ExecutePython(ctx context.Context) error {
	if a.Bypass {
		var err error
		if a.PythonExecutable, err = exec.LookPath(a.PythonExecutable); err != nil {
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
