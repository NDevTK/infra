// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"strings"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	api "go.chromium.org/luci/swarming/proto/api"
)

// EditSWJobDefinition is a temporary type returned by
// JobDefinition.EditSwarming. It holds a mutable swarming-based
// JobDefinition and an error, allowing a series of Edit commands to be called
// while buffering the error (if any).  Obtain the modified JobDefinition (or
// error) by calling Finalize.
type EditSWJobDefinition struct {
	ctx context.Context

	sw          *Swarming
	userPayload *api.CASTree

	err error
}

// EditSwarming returns a mutator wrapper which knows how to manipulate
// various aspects of a Swarming-based JobDefinition.
func (jd *JobDefinition) EditSwarming(ctx context.Context, authOpts auth.Options, fn func(*EditSWJobDefinition)) error {
	if err := jd.FlattenToSwarming(ctx, authOpts); err != nil {
		return errors.Annotate(err, "flattening to a swarming JobDefinition").Err()
	}

	sw := jd.GetSwarming()
	if sw == nil {
		return errors.New("only supported for Swarming builds")
	}

	ejd := &EditSWJobDefinition{ctx, sw, jd.UserPayload, nil}
	fn(ejd)
	jd.UserPayload = ejd.userPayload

	return ejd.err
}

func (ejd *EditSWJobDefinition) tweak(fn func() error) {
	if ejd.err == nil {
		ejd.err = fn()
	}
}

func (ejd *EditSWJobDefinition) tweakSlices(fn func(*api.TaskSlice) error) {
	ejd.tweak(func() error {
		for _, slice := range ejd.sw.GetTask().TaskSlices {
			if err := fn(slice); err != nil {
				return err
			}
		}
		return nil
	})
}

// EditIsolated replaces the non-recipe isolate in the TaskSlice.
func (ejd *EditSWJobDefinition) EditIsolated(isolated string) {
	if isolated == "" {
		return
	}
	ejd.tweakSlices(func(slc *api.TaskSlice) error {
		props := slc.Properties
		if props == nil {
			slc.Properties = &api.TaskProperties{}
			props = slc.Properties
		}

		cas := props.CasInputs
		if cas == nil {
			props.CasInputs = &api.CASTree{}
			cas = props.CasInputs
		}
		cas.Digest = isolated
		return nil
	})
}

// Dimensions edits the swarming dimensions.
//
// TODO(iannucci): when named cache dimension support was added to swarming,
// request dimensions became plural; it's now legitimate for input dimensions to
// have multiple values for the same key. This function (and it's command line
// analog) should be updated to account for this. A possible solution would be
// to change the CLI syntax to an order of 'commands' which are evaluated in
// order:
//   * key-value : remove 'value' from the set of values associated with 'key'
//   * key=value : set 'key' to exactly the set {value}
//   * key=      : set 'key' to exactly the set {}
//   * key+value : add 'value' to the set of values associated with 'key'
//
// This would allow full manipulation of the dimension set from the command
// line. This syntax could also be used to improve the env prefix CLI... maybe.
func (ejd *EditSWJobDefinition) Dimensions(dims map[string]string) {
	if len(dims) == 0 {
		return
	}

	ejd.tweakSlices(func(slc *api.TaskSlice) error {
		updateStringPairList((*swarmDims)(&slc.Properties.Dimensions), dims)
		return nil
	})
}

// Env edits the swarming environment variables (i.e. before kitchen).
func (ejd *EditSWJobDefinition) Env(env map[string]string) {
	if len(env) == 0 {
		return
	}

	ejd.tweakSlices(func(slc *api.TaskSlice) error {
		updateStringPairList((*swarmEnvs)(&slc.Properties.Env), env)
		return nil
	})
}

// Priority edits the swarming task priority.
func (ejd *EditSWJobDefinition) Priority(priority int32) {
	if priority < 0 {
		return
	}
	ejd.tweak(func() error {
		ejd.sw.Task.Priority = priority
		return nil
	})
}

func updateCipdPkgs(cipdPkgs map[string]string, pinSets ...*[]*api.CIPDPackage) {
	updates := map[string]map[string]string{}
	for subdirPkg, vers := range cipdPkgs {
		subdir := "."
		pkg := subdirPkg
		if toks := strings.SplitN(subdirPkg, ":", 2); len(toks) > 1 {
			subdir, pkg = toks[0], toks[1]
		}
		if updates[subdir] == nil {
			updates[subdir] = map[string]string{}
		}
		updates[subdir][pkg] = vers
	}

	addToPins := func(pins *[]*api.CIPDPackage) {
		// subdir -> pkg -> version
		currentState := map[string]map[string]string{}
		for _, pin := range *pins {
			subdir := pin.DestPath
			if subdir == "" {
				subdir = "."
			}
			if currentState[subdir] == nil {
				currentState[subdir] = map[string]string{}
			}
			currentState[subdir][pin.PackageName] = pin.Version
		}

		count := 0
		for subdir, pkgsVers := range updates {
			destSubdir := currentState[subdir]
			if destSubdir == nil {
				destSubdir = map[string]string{}
				currentState[subdir] = destSubdir
			}
			count += len(destSubdir)

			for pkg, vers := range pkgsVers {
				if vers == "" {
					delete(destSubdir, pkg)
					count--
				} else {
					if destSubdir[pkg] != "" {
						count++
					}
					destSubdir[pkg] = vers
				}
			}
		}

		newPins := make([]*api.CIPDPackage, 0, count)
		for subdir, pkgsVers := range currentState {
			for pkg, vers := range pkgsVers {
				newPins = append(newPins, &api.CIPDPackage{
					DestPath: subdir, PackageName: pkg, Version: vers,
				})
			}
		}
		*pins = newPins
	}

	for _, pins := range pinSets {
		addToPins(pins)
	}
}

// CipdPkgs allows you to edit the cipd packages. The mapping is in the form of:
//    subdir:name/of/package -> version
// If version is empty, this package will be removed (if it's present).
func (ejd *EditSWJobDefinition) CipdPkgs(cipdPkgs map[string]string) {
	if len(cipdPkgs) == 0 {
		return
	}

	ejd.tweak(func() error {
		pinSets := []*[]*api.CIPDPackage{}
		ejd.tweakSlices(func(slc *api.TaskSlice) error {
			pinSets = append(pinSets, &slc.Properties.CipdInputs)
			return nil
		})
		updateCipdPkgs(cipdPkgs, pinSets...)
		return nil
	})
}

// SwarmingHostname allows you to modify the current SwarmingHostname used by this
// led pipeline. Note that the isolated server is derived from this, so
// if you're editing this value, do so before passing the JobDefinition through
// the `isolate` subcommand.
func (ejd *EditSWJobDefinition) SwarmingHostname(host string) {
	if host == "" {
		return
	}

	ejd.tweak(func() (err error) {
		if err = errors.Annotate(validateHost(host), "SwarmingHostname").Err(); err == nil {
			ejd.sw.Hostname = host
		}
		return
	})
}

func updatePrefixPathEnv(values []string, envLists ...*[]*api.StringListPair) {
	doUpdate := func(prefixes *[]*api.StringListPair) {
		var newPath []string
		for _, pair := range *prefixes {
			if pair.Key == "PATH" {
				newPath = pair.Values
				break
			}
		}

		for _, v := range values {
			if strings.HasPrefix(v, "!") {
				var toCut []int
				for i, cur := range newPath {
					if cur == v[1:] {
						toCut = append(toCut, i)
					}
				}
				for _, i := range toCut {
					newPath = append(newPath[:i], newPath[i+1:]...)
				}
			} else {
				newPath = append(newPath, v)
			}
		}

		for _, pair := range *prefixes {
			if pair.Key == "PATH" {
				pair.Values = newPath
				return
			}
		}

		*prefixes = append(
			*prefixes, &api.StringListPair{Key: "PATH", Values: newPath})
	}

	for _, lst := range envLists {
		doUpdate(lst)
	}
}

// PrefixPathEnv controls swarming's env_prefix mapping.
//
// Values prepended with '!' will remove them from the existing list of values
// (if present). Otherwise these values will be appended to the current list of
// path-prefix-envs.
func (ejd *EditSWJobDefinition) PrefixPathEnv(values []string) {
	if len(values) == 0 {
		return
	}

	ejd.tweak(func() error {
		envLists := []*[]*api.StringListPair{}
		ejd.tweakSlices(func(slc *api.TaskSlice) error {
			envLists = append(envLists, &slc.Properties.EnvPaths)
			return nil
		})
		updatePrefixPathEnv(values, envLists...)
		return nil
	})
}

// Tags controls swarming tags.
func (ejd *EditSWJobDefinition) Tags(values []string) {
	if len(values) == 0 {
		return
	}
	ejd.tweak(func() error {
		ejd.sw.Task.Tags = append(ejd.sw.Task.Tags, values...)
		return nil
	})
}
