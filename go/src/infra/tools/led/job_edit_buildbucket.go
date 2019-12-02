// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	api "go.chromium.org/luci/swarming/proto/api"
)

// EditBBJobDefinition is a temporary type returned by
// JobDefinition.EditBuildbucket. It holds a mutable buildbucket-based
// JobDefinition and an error, allowing a series of Edit commands to be called
// while buffering the error (if any).  Obtain the modified JobDefinition (or
// error) by calling Finalize.
type EditBBJobDefinition struct {
	bb          *Buildbucket
	userPayload *api.CASTree

	err error
}

// EditBuildbucket returns a mutator wrapper which knows how to manipulate
// various aspects of a Buildbucket-based JobDefinition.
func (jd *JobDefinition) EditBuildbucket(fn func(*EditBBJobDefinition)) error {
	bb := jd.GetBuildbucket()
	if bb == nil {
		return errors.New("only supported for Buildbucket builds")
	}
	bb.EnsureBasics()

	ejd := &EditBBJobDefinition{bb, jd.UserPayload, nil}
	fn(ejd)
	jd.UserPayload = ejd.userPayload

	return ejd.err
}

func (ejd *EditBBJobDefinition) tweak(fn func() error) {
	if ejd.err == nil {
		ejd.err = fn()
	}
}

// Recipe modifies the recipe to run. This must be resolvable in the current
// recipe source.
func (ejd *EditBBJobDefinition) Recipe(recipe string) {
	if recipe == "" {
		return
	}
	ejd.tweak(func() error {
		ejd.bb.WriteProperties(map[string]interface{}{
			"recipe": recipe,
		})
		return nil
	})
}

// RecipeSource modifies the source for the recipes. This can either be an
// isolated hash (i.e. bundled recipes) or it can be a cipd pkg/version pair.
func (ejd *EditBBJobDefinition) RecipeSource(isolated, cipdPkg, cipdVer string) {
	if isolated == "" && cipdPkg == "" && cipdVer == "" {
		return
	}

	ejd.tweak(func() error {
		switch {
		case isolated != "":
			if cipdPkg != "" {
				return errors.New("specify either isolated or cipdPkg, but not both")
			}
			ejd.userPayload.Digest = isolated
			ejd.bb.BbagentArgs.Build.Exe = nil
			recipe := ejd.bb.BbagentArgs.Build.Infra.Recipe
			ejd.bb.BbagentArgs.Build.Infra.Recipe = nil

			if recipe != nil {
				ejd.bb.WriteProperties(map[string]interface{}{
					"recipe": recipe.Name,
				})
			}

		default:
			ejd.bb.BbagentArgs.Build.Exe = &bbpb.Executable{
				CipdPackage: cipdPkg,
				CipdVersion: cipdVer,
			}
			ejd.userPayload = nil
		}

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
func (ejd *EditBBJobDefinition) Dimensions(dims map[string]string) {
	if len(dims) == 0 {
		return
	}

	ejd.tweak(func() error {
		updateStringPairList(
			(*bbReqDims)(&ejd.bb.BbagentArgs.Build.Infra.Swarming.TaskDimensions),
			dims)
		return nil
	})
}

// Env edits the swarming environment variables (i.e. before kitchen).
func (ejd *EditBBJobDefinition) Env(env map[string]string) {
	if len(env) == 0 {
		return
	}

	ejd.tweak(func() error {
		updateStringPairList((*swarmEnvs)(&ejd.bb.EnvVars), env)
		return nil
	})
}

// Priority edits the swarming task priority.
func (ejd *EditBBJobDefinition) Priority(priority int32) {
	if priority < 0 {
		return
	}
	ejd.tweak(func() error {
		ejd.bb.BbagentArgs.Build.Infra.Swarming.Priority = priority
		return nil
	})
}

// Properties edits the recipe properties.
func (ejd *EditBBJobDefinition) Properties(props map[string]string, auto bool) {
	if len(props) == 0 {
		return
	}
	ejd.tweak(func() error {
		b := ejd.bb.BbagentArgs.Build

		toWrite := map[string]interface{}{}

		for k, v := range props {
			if v == "" {
				delete(b.Input.Properties.Fields, k)
			} else {
				var obj interface{}
				if err := json.Unmarshal([]byte(v), &obj); err != nil {
					if !auto {
						return err
					}
					obj = v
				}
				toWrite[k] = obj
			}
		}

		ejd.bb.WriteProperties(toWrite)
		return nil
	})
}

// CipdPkgs allows you to edit the cipd packages. The mapping is in the form of:
//    subdir:name/of/package -> version
// If version is empty, this package will be removed (if it's present).
func (ejd *EditBBJobDefinition) CipdPkgs(cipdPkgs map[string]string) {
	if len(cipdPkgs) == 0 {
		return
	}

	ejd.tweak(func() error {
		updateCipdPkgs(cipdPkgs, &ejd.bb.CipdPackages)
		return nil
	})
}

// SwarmingHostname allows you to modify the current SwarmingHostname used by this
// led pipeline. Note that the isolated server is derived from this, so
// if you're editing this value, do so before passing the JobDefinition through
// the `isolate` subcommand.
func (ejd *EditBBJobDefinition) SwarmingHostname(host string) {
	if host == "" {
		return
	}

	ejd.tweak(func() (err error) {
		if err = errors.Annotate(validateHost(host), "SwarmingHostname").Err(); err == nil {
			ejd.bb.BbagentArgs.Build.Infra.Swarming.Hostname = host
		}
		return
	})
}

// Experimental allows you to conveniently modify the
// "$recipe_engine/runtime['is_experimental']" property.
func (ejd *EditBBJobDefinition) Experimental(trueOrFalse string) {
	if trueOrFalse == "" {
		return
	}

	ejd.tweak(func() error {
		ejd.bb.BbagentArgs.Build.Input.Experimental = (trueOrFalse == "true")
		return nil
	})
}

// PrefixPathEnv controls swarming's env_prefix mapping.
//
// Values prepended with '!' will remove them from the existing list of values
// (if present). Otherwise these values will be appended to the current list of
// path-prefix-envs.
func (ejd *EditBBJobDefinition) PrefixPathEnv(values []string) {
	if len(values) == 0 {
		return
	}

	ejd.tweak(func() error {
		updatePrefixPathEnv(values, &ejd.bb.EnvPrefixes)
		return nil
	})
}
