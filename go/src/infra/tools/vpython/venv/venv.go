// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package venv

import (
	"os"
	"path/filepath"
	"time"

	"infra/tools/vpython/api/env"
	"infra/tools/vpython/filesystem"
	"infra/tools/vpython/python"
	"infra/tools/vpython/spec"
	"infra/tools/vpython/wheel"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"

	"github.com/danjacques/gofslock/fslock"
	"golang.org/x/net/context"
)

const lockHeldDelay = 5 * time.Second

// blocker is an fslock.Blocker implementation that sleeps lockHeldDelay in
// between attempts.
func blocker(c context.Context) fslock.Blocker {
	return func() error {
		logging.Debugf(c, "Lock is currently held. Sleeping %v and retrying...", lockHeldDelay)
		clock.Sleep(c, lockHeldDelay)
		return nil
	}
}

// EnvRootFromSpecPath calculates the environment root from an exported
// environment specification file path.
//
// The specification path is: <EnvRoot>/<SpecHash>/<spec>, so our EnvRoot
// is two directories up.
//
// We export EnvSpecPath as an asbolute path. However, since someone else
// could have overridden it or exported their own, let's make sure.
func EnvRootFromSpecPath(path string) (string, error) {
	if err := filesystem.AbsPath(&path); err != nil {
		return "", errors.Annotate(err).
			Reason("failed to get absolute path for specification file path: %(path)s").
			Err()
	}
	return filepath.Dir(filepath.Dir(path)), nil
}

// Env is a fully set-up Python virtual enviornment. It is configured
// based on the contents of an env.Spec file by Setup.
//
// Env should not be instantiated directly; it must be created by calling
// Config.Env.
//
// All paths in Env are absolute.
type Env struct {
	// Config is this Env's Config, fully-resolved.
	Config *Config

	// Root is the Env container's root directory path.
	Root string

	// Python is the path to the Env Python interpreter.
	Python string

	// SepcPath is the path to the specification file that was used to construct
	// this enviornment. It will be in text protobuf format, and, therefore,
	// suitable for input to other "vpython" invocations.
	SpecPath string

	// name is the hash of the specification file for this Env.
	name string
	// lockPath is the path to this Env-specific lock file. It will be at:
	// "<baseDir>/.<name>.lock".
	lockPath string
	// completeFlagPath is the path to this Env's complete flag.
	// It will be at "<Root>/complete.flag".
	completeFlagPath string
}

// Setup creates a new Env.
//
// It will lock around the Env to ensure that multiple processes do not
// conflict with each other. If a Env for this specification already
// exists, it will be used directly without any additional setup.
//
// If another process holds the lock and blocking is true, we will wait for our
// turn at the lock. Otherwise, we will return immediately with a locking error.
func (e *Env) Setup(c context.Context, blocking bool) error {
	if err := e.setupImpl(c, blocking); err != nil {
		return errors.Annotate(err).Err()
	}

	// Perform a pruning round. Failure is non-fatal.
	if perr := prune(c, e.Config, e.name); perr != nil {
		logging.WithError(perr).Warningf(c, "Failed to perform pruning round after initialization.")
	}
	return nil
}

func (e *Env) setupImpl(c context.Context, blocking bool) error {
	// Repeatedly try and create our Env. We do this so that if we
	// encounter a lock, we will let the other process finish and try and leverage
	// its success.
	for {
		// Fast path: if our complete flag is present, assume that the environment
		// is setup and complete. No locking or additional work necessary.
		if _, err := os.Stat(e.completeFlagPath); err == nil {
			logging.Debugf(c, "Completion flag found! Environment is set-up: %s", e.completeFlagPath)

			// Update the complete flag so the timestamp reflects our usage of it.
			// This is non-fatal if it fails.
			if err := e.touchCompleteFlag(); err != nil {
				logging.WithError(err).Warningf(c, "Failed to update complete flag.")
			}

			return nil
		}

		// We will be creating the Env. We will to lock around a file for this
		// Env hash so that any other processes that may be trying to
		// simultaneously create a Env will be forced to wait.
		err := fslock.With(e.lockPath, func() error {
			// Mark that we hit some lock contention. If we did, we will try again
			// from scratch.
			if err := e.createLocked(c); err != nil {
				return errors.Annotate(err).Reason("failed to create new VirtualEnv").Err()
			}
			return nil
		})
		switch err {
		case nil:
			// Successfully created the environment! Mark this with a completion flag.
			if err := e.touchCompleteFlag(); err != nil {
				return errors.Annotate(err).Reason("failed to create complete flag").Err()
			}
			return nil

		case fslock.ErrLockHeld:
			if !blocking {
				return errors.Annotate(err).Reason("VirtualEnv lock is currently held (non-blocking)").Err()
			}

			// Some other process holds the lock. Sleep a little and retry.
			logging.Warningf(c, "VirtualEnv lock is currently held. Retrying after delay (%s)...",
				lockHeldDelay)
			if tr := clock.Sleep(c, lockHeldDelay); tr.Incomplete() {
				return tr.Err
			}
			continue

		default:
			return errors.Annotate(err).Reason("failed to create VirtualEnv").Err()
		}
	}

	return nil
}

// Delete deletes this enviornment, if it exists.
func (e *Env) Delete(c context.Context) error {
	err := fslock.WithBlocking(e.lockPath, blocker(c), func() error {
		if err := e.deleteLocked(c); err != nil {
			return errors.Annotate(err).Err()
		}
		return nil
	})
	if err != nil {
		errors.Annotate(err).Reason("failed to delete enviornment").Err()
	}
	return nil
}

func (e *Env) createLocked(c context.Context) error {
	// If our root directory already exists, delete it.
	if _, err := os.Stat(e.Root); err == nil {
		logging.Warningf(c, "Deleting existing VirtualEnv: %s", e.Root)
		if err := filesystem.RemoveAll(e.Root); err != nil {
			return errors.Reason("failed to remove existing root").Err()
		}
	}

	// Make sure our environment's base directory exists.
	if err := filesystem.MakeDirs(e.Root); err != nil {
		return errors.Annotate(err).Reason("failed to create environment root").Err()
	}
	logging.Infof(c, "Using virtual environment root: %s", e.Root)

	// Build our package list. Always install our base VirtualEnv package.
	// For resolution purposes, our VirtualEnv package will be index 0.
	packages := make([]*env.Spec_Package, 1, 1+len(e.Config.Spec.Wheel))
	packages[0] = e.Config.Spec.Virtualenv
	packages = append(packages, e.Config.Spec.Wheel...)

	bootstrapDir := filepath.Join(e.Root, ".vpython_bootstrap")
	pkgDir := filepath.Join(bootstrapDir, "packages")
	if err := filesystem.MakeDirs(pkgDir); err != nil {
		return errors.Annotate(err).Reason("could not create bootstrap packages directory").Err()
	}

	if err := e.downloadPackages(c, pkgDir, packages); err != nil {
		return errors.Annotate(err).Reason("failed to download packages").Err()
	}

	// Installing base VirtualEnv.
	if err := e.installVirtualEnv(c, pkgDir); err != nil {
		return errors.Annotate(err).Reason("failed to install VirtualEnv").Err()
	}

	// Download our wheel files.
	if len(e.Config.Spec.Wheel) > 0 {
		// Install wheels into our VirtualEnv.
		if err := e.installWheels(c, bootstrapDir, pkgDir); err != nil {
			return errors.Annotate(err).Reason("failed to install wheels").Err()
		}
	}

	// Write our specification file.
	if err := spec.Write(e.Config.Spec, e.SpecPath); err != nil {
		return errors.Annotate(err).Reason("failed to write spec file to: %(path)s").
			D("path", e.SpecPath).
			Err()
	}
	logging.Debugf(c, "Wrote specification file to: %s", e.SpecPath)

	// Finalize our VirtualEnv for bootstrap execution.
	if err := e.finalize(c, bootstrapDir); err != nil {
		return errors.Annotate(err).Reason("failed to prepare VirtualEnv").Err()
	}

	return nil
}

func (e *Env) downloadPackages(c context.Context, dst string, packages []*env.Spec_Package) error {
	// Create a wheel sub-directory underneath of root.
	logging.Debugf(c, "Loading %d package(s) into: %s", len(packages), dst)
	if err := e.Config.Loader.Ensure(c, dst, packages); err != nil {
		return errors.Annotate(err).Reason("failed to download packages").Err()
	}
	return nil
}

func (e *Env) installVirtualEnv(c context.Context, pkgDir string) error {
	// Create our VirtualEnv package staging sub-directory underneath of root.
	bsDir := filepath.Join(e.Root, ".virtualenv")
	if err := filesystem.MakeDirs(bsDir); err != nil {
		return errors.Annotate(err).Reason("failed to create VirtualEnv bootstrap directory").
			D("path", bsDir).
			Err()
	}

	// Identify the virtualenv directory: will have "virtualenv-" prefix.
	matches, err := filepath.Glob(filepath.Join(pkgDir, "virtualenv-*"))
	if err != nil {
		return errors.Annotate(err).Reason("failed to glob for 'virtualenv-' directory").Err()
	}
	if len(matches) == 0 {
		return errors.Reason("no 'virtualenv-' directory provided by package").Err()
	}

	logging.Debugf(c, "Creating VirtualEnv at: %s", e.Root)
	i := e.Config.systemInterpreter()
	i.WorkDir = matches[0]
	err = i.Run(c,
		"virtualenv.py",
		"--no-download",
		e.Root)
	if err != nil {
		return errors.Annotate(err).Reason("failed to create VirtualEnv").Err()
	}

	return nil
}

func (e *Env) installWheels(c context.Context, bootstrapDir, pkgDir string) error {
	// Identify all downloaded wheels and parse them.
	wheels, err := wheel.GlobFrom(pkgDir)
	if err != nil {
		return errors.Annotate(err).Reason("failed to load wheels").Err()
	}

	// Build a "wheel" requirements file.
	reqPath := filepath.Join(bootstrapDir, "requirements.txt")
	logging.Debugf(c, "Rendering requirements file to: %s", reqPath)
	if err := wheel.WriteRequirementsFile(reqPath, wheels); err != nil {
		return errors.Annotate(err).Reason("failed to render requirements file").Err()
	}

	i := e.venvInterpreter()
	err = i.Run(c,
		"-m", "pip",
		"install",
		"--use-wheel",
		"--compile",
		"--no-index",
		"--find-links", pkgDir,
		"--requirement", reqPath)
	if err != nil {
		return errors.Annotate(err).Reason("failed to install wheels").Err()
	}
	return nil
}

func (e *Env) finalize(c context.Context, bootstrapDir string) error {
	// Uninstall "pip" and "wheel", preventing (easy) augmentation of the
	// enviornment.
	i := e.venvInterpreter()
	err := i.Run(c,
		"-m", "pip",
		"uninstall",
		"--quiet",
		"--yes",
		"pip", "wheel")
	if err != nil {
		return errors.Annotate(err).Reason("failed to install wheels").Err()
	}

	// Delete our bootstrap directory (non-fatal).
	if err := filesystem.RemoveAll(bootstrapDir); err != nil {
		logging.WithError(err).Warningf(c, "Failed to delete bootstrap directory: %s", bootstrapDir)
	}

	// Change all files to read-only, except:
	// - Our root directory, which must be writable in order to update our
	//   completion flag.
	// - Our completion flag, which must be trivially re-writable.
	err = filesystem.MakeReadOnly(e.Root, func(path string) bool {
		switch path {
		case e.Root, e.completeFlagPath:
			return false
		default:
			return true
		}
	})
	if err != nil {
		return errors.Annotate(err).Reason("failed to mark environment read-only").Err()
	}
	return nil
}

func (e *Env) venvInterpreter() *python.Interpreter {
	i := e.Interpreter()
	i.WorkDir = e.Root
	return i
}

// Interpreter returns a Python interpreter pointing to the VirtualEnv's Python
// installation.
func (e *Env) Interpreter() *python.Interpreter {
	return &python.Interpreter{
		Python:   e.Python,
		Isolated: true,
	}
}

// touchCompleteFlag touches the complete flag, creating it and/or updating its
// timestamp.
//
// This is safe to call without the lock held, since worst-case an update is
// overwritten if contested.
func (e *Env) touchCompleteFlag() error {
	if err := filesystem.Touch(e.completeFlagPath, 0644); err != nil {
		return errors.Annotate(err).Err()
	}
	return nil
}

func (e *Env) deleteLocked(c context.Context) error {
	// Delete our environment directory.
	if err := filesystem.RemoveAll(e.Root); err != nil {
		return errors.Annotate(err).Reason("failed to delete environment root").Err()
	}

	// Delete our lock path.
	if err := os.Remove(e.lockPath); err != nil {
		return errors.Annotate(err).Reason("failed to delete lock").Err()
	}
	return nil
}
