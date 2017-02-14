// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package application

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"infra/tools/vpython"
	"infra/tools/vpython/api/env"
	"infra/tools/vpython/cipd"
	"infra/tools/vpython/filesystem"
	"infra/tools/vpython/python"
	"infra/tools/vpython/spec"
	"infra/tools/vpython/venv"

	cipdClient "github.com/luci/luci-go/cipd/client/cipd"
	"github.com/luci/luci-go/common/cli"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/logging/gologger"
	"github.com/luci/luci-go/common/system/environ"

	"github.com/maruel/subcommands"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/net/context"
)

// An A is an application's default configuration.
type A struct {
	// VENVPackage is the CIPD VirtualEnv package to use for bootstrap generation.
	VENVPackage env.Spec_Package

	// PruneThreshold, if > 0, is the maximum age of a VirtualEnv before it
	// becomes candidate for pruning. If <= 0, no pruning will be performed.
	//
	// See venv.Config's PruneThreshold.
	PruneThreshold time.Duration
	// PruneLimit, if > 0, is the maximum number of VirtualEnv that should be
	// pruned passively. If <= 0, no limit will be applied.
	//
	// See venv.Config's PruneLimit.
	PruneLimit int

	// CIPDServiceURL is the CIPD service URL string. If empty, the default
	// service URL will be used.
	CIPDServiceURL string

	// Opts is the set of configured options.
	Opts vpython.Options
}

func (a *A) mainDev(c context.Context) error {
	app := cli.Application{
		Name:  "vpython",
		Title: "VirtualEnv Python Bootstrap (Development Mode)",
		Context: func(context.Context) context.Context {
			// Discard the entry Context and use the one passed to us.
			c := c

			// Install our A instance into the Context.
			c = withApplication(c, a)

			// Drop down to Info level debugging.
			if logging.GetLevel(c) > logging.Info {
				c = logging.SetLevel(c, logging.Info)
			}
			return c
		},
		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			subcommandInstall,
		},
	}

	return python.Error(subcommands.Run(&app, a.Opts.Args))
}

func (a *A) mainImpl(c context.Context, args []string) error {
	logConfig := logging.Config{
		Level: logging.Warning,
	}

	hdir, err := homedir.Dir()
	if err != nil {
		return errors.Annotate(err).Reason("failed to get user home directory").Err()
	}

	a.Opts = vpython.Options{
		EnvConfig: venv.Config{
			BaseDir:        filepath.Join(hdir, ".vpython"),
			MaxHashLen:     6,
			Package:        a.VENVPackage,
			PruneThreshold: a.PruneThreshold,
			PruneLimit:     a.PruneLimit,
			Loader: &cipd.PackageLoader{
				Options: cipdClient.ClientOptions{
					ServiceURL: a.CIPDServiceURL,
					UserAgent:  "vpython",
				},
			},
		},
		WaitForEnv: true,
		Environ:    environ.System(),
	}
	var specPath string
	var devMode bool

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.BoolVar(&devMode, "dev", devMode,
		"Enter devevelopment / subcommand mode (use 'help' for more options).")
	fs.StringVar(&a.Opts.EnvConfig.Python, "python", a.Opts.EnvConfig.Python,
		"Path to system Python interpreter to use. Default is found on PATH.")
	fs.StringVar(&a.Opts.WorkDir, "workdir", a.Opts.WorkDir,
		"Working directory to run the Python interpreter in. Default is current working directory.")
	fs.StringVar(&a.Opts.EnvConfig.BaseDir, "root", a.Opts.EnvConfig.BaseDir,
		"Path to virtual enviornment root directory. Default is the working directory. "+
			"If explicitly set to empty string, a temporary directory will be used and cleaned up "+
			"on completion.")
	fs.StringVar(&specPath, "spec", specPath,
		"Path to enviornment specification file to load. Default probes for one.")
	logConfig.AddFlags(fs)

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return errors.Annotate(err).Reason("failed to parse flags").Err()
	}
	a.Opts.Args = fs.Args()

	c = logConfig.Set(c)

	// If an spec path was manually specified, load and use it.
	if specPath != "" {
		var err error
		if a.Opts.EnvConfig.Spec, err = spec.Load(specPath); err != nil {
			return errors.Annotate(err).Reason("failed to load environment specification file (-env) from: %(path)s").
				D("path", specPath).
				Err()
		}
	}

	// If an empty BaseDir was specified, use a temporary directory and clean it
	// up on completion.
	if a.Opts.EnvConfig.BaseDir == "" {
		tdir, err := ioutil.TempDir("", "vpython")
		if err != nil {
			return errors.Annotate(err).Reason("failed to create temporary directory").Err()
		}
		defer func() {
			logging.Debugf(c, "Removing temporary directory: %s", tdir)
			if terr := filesystem.RemoveAll(tdir); terr != nil {
				logging.WithError(terr).Warningf(c, "Failed to clean up temporary directory; leaking: %s", tdir)
			}
		}()
		a.Opts.EnvConfig.BaseDir = tdir
	}

	// Development mode (subcommands).
	if devMode {
		return a.mainDev(c)
	}

	if err := vpython.Run(c, a.Opts); err != nil {
		return errors.Annotate(err).Err()
	}
	return nil
}

// Main is the main application entry point.
func (a *A) Main(c context.Context) int {
	c = gologger.StdConfig.Use(c)
	c = logging.SetLevel(c, logging.Warning)

	return run(c, func(c context.Context) error {
		return a.mainImpl(c, os.Args[1:])
	})
}
