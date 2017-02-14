package vpython

import (
	"context"
	"os"

	"infra/tools/vpython/filesystem"
	"infra/tools/vpython/python"
	"infra/tools/vpython/spec"
	"infra/tools/vpython/venv"

	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/system/environ"
)

// Options is the set of options to use to construct and execute a VirtualEnv
// Python application.
type Options struct {
	// EnvConfig is the VirtualEnv configuration to run from.
	EnvConfig venv.Config

	// Args are the arguments to forward to the Python process.
	Args []string

	// WaitForEnv, if true, means that if another agent holds a lock on the target
	// environment, we will wait until it is available. If false, we will
	// immediately exit Setup with an error.
	WaitForEnv bool

	// WorkDir is the Python working directory. If empty, the current working
	// directory will be used.
	//
	// If EnvRoot is empty, WorkDir will be used as the base enviornment root.
	WorkDir string

	// Environ is the system environment. It can be used for some default values
	// if present.
	Environ environ.Env
}

func (o *Options) resolve(c context.Context) error {
	// Resolve our working directory to an absolute path.
	if o.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Annotate(err).Reason("failed to get working directory").Err()
		}
		o.WorkDir = wd
	}
	if err := filesystem.AbsPath(&o.WorkDir); err != nil {
		return errors.Annotate(err).Reason("failed to resolve absolute path of WorkDir").Err()
	}

	// Resolve our target python script.
	if err := o.resolveEnvSpec(c); err != nil {
		return errors.Annotate(err).Reason("failed to resolve Python script").Err()
	}

	// If no enviornment base directory was supplied, create one under the current
	// working directory.
	if o.EnvConfig.BaseDir == "" {
		// Is one specified in our environment?
		if v, ok := o.Environ.Get(EnvSpecPath); ok {
			var err error
			if o.EnvConfig.BaseDir, err = venv.EnvRootFromSpecPath(v); err != nil {
				return errors.Annotate(err).Reason("failed to get env root from environment: %(path)s").
					D("path", v).
					Err()
			}
			logging.Debugf(c, "Loaded environment root from environment variable: %s", o.EnvConfig.BaseDir)
		}
	}
	return nil
}

func (o *Options) resolveEnvSpec(c context.Context) error {
	cmd, err := python.ParseCommandLine(o.Args)
	if err != nil {
		return errors.Annotate(err).Reason("failed to parse Python command-line").Err()
	}

	// If we're running a Python script, assert that the target script exists.
	// Additionally, track whether its a file or a module (directory).
	isModule := false
	if cmd.Type == python.TargetScript {
		logging.Debugf(c, "Resolved Python target script: %s", cmd.Value)

		// Resolve to absolute script path.
		if err := filesystem.AbsPath(&cmd.Value); err != nil {
			return errors.Annotate(err).Reason("failed to get absolute path of: %(path)s").
				D("path", cmd.Value).
				Err()
		}

		// Confirm that the script path actually exists.
		st, err := os.Stat(cmd.Value)
		if err != nil {
			return errors.Annotate(err).Reason("failed to stat Python script: %(path)s").
				D("path", cmd.Value).
				Err()
		}

		// If the script is a directory, then we assume that we're doing a module
		// invocation (__main__.py).
		isModule = st.IsDir()
	}

	// If a spec is explicitly provided, we're done.
	if o.EnvConfig.Spec != nil {
		return nil
	}

	// If it's a script, try resolving from filesystem first.
	if cmd.Type == python.TargetScript {
		spec, err := spec.LoadForScript(c, cmd.Value, isModule)
		if err != nil {
			return errors.Annotate(err).Reason("failed to load spec for script: %(path)s").
				D("path", cmd.Value).
				D("isModule", isModule).
				Err()
		}
		if spec != nil {
			o.EnvConfig.Spec = spec
			return nil
		}
	}

	// Do we have a spec file in the enviornment?
	if v, ok := o.Environ.Get(EnvSpecPath); ok {
		if o.EnvConfig.Spec, err = spec.Load(v); err != nil {
			return errors.Annotate(err).Reason("failed to load enviornment-supplied spec from: %(path)s").
				D("path", v).
				Err()
		}
		logging.Infof(c, "Loaded spec from environment: %s", v)
		return nil
	}

	// Unable to load a spec.
	logging.Infof(c, "Unable to resolve specification path. Using empty specification.")
	return nil
}
