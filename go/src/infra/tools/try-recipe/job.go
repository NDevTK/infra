// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"strings"

	"golang.org/x/net/context"

	"github.com/luci/luci-go/client/archiver"
	swarming "github.com/luci/luci-go/common/api/swarming/swarming/v1"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/isolated"
)

const recipePropertiesJSON = "$RECIPE_PROPERTIES_JSON"
const recipeName = "$RECIPE_NAME"
const recipeCheckoutDir = "recipe-checkout-dir"

// JobDefinition defines a 'try-recipe' job. It's like a normal Swarming
// NewTaskRequest, but with some recipe-specific extras.
//
// In particular, the RecipeIsolatedHash will be combined with the task's
// isolated (if any), by uploading a new isolated which 'includes' both.
//
// Additionally, RecipeProperties will replace any args in the swarming task's
// command which are the string $RECIPE_PROPERTIES_JSON.
type JobDefinition struct {
	RecipeIsolatedHash isolated.HexDigest `json:"recipe_isolated_hash"`

	RecipeName       string                 `json:"recipe_name"`
	RecipeProperties map[string]interface{} `json:"recipe_properties"`

	// TODO(iannucci):
	// this should really be a swarming.SwarmingRpcsNewTaskRequest, but the way
	// that buildbucket sends it is incompatible with the go endpoints generated
	// struct. Hooray...  *rollseyes*.
	SwarmingTask *swarming.SwarmingRpcsNewTaskRequest `json:"swarming_task"`
}

func JobDefinitionFromNewTaskRequest(r *swarming.SwarmingRpcsNewTaskRequest) (*JobDefinition, error) {
	ret := &JobDefinition{SwarmingTask: r}

	// This function removes a flag when used in the toProcess map below.
	deleteFn := func(value string) ([]string, error) { return nil, nil }

	// This is a map of flag to a processing function.
	//
	// The processing function recieves the flag's 'value', and is expected to
	// return the replacement for the flag+value. Returning a nil replacement will
	// delete the flag from the command line, entirely.
	toProcess := map[string]func(value string) (replace []string, err error){
		"-properties": func(value string) ([]string, error) {
			ret.RecipeProperties = map[string]interface{}{}
			if err := json.NewDecoder(strings.NewReader(value)).Decode(&ret.RecipeProperties); err != nil {
				return nil, errors.Annotate(err).Reason("decoding -properties JSON").Err()
			}
			return []string{"-properties", recipePropertiesJSON}, nil
		},

		"-recipe": func(value string) ([]string, error) {
			ret.RecipeName = value
			return []string{"-recipe", recipeName}, nil
		},

		// we need to remove this so that it doesn't conflict with the replacement
		// in -repository.
		"-checkout-dir": deleteFn,

		"-repository": func(value string) ([]string, error) {
			return []string{"-checkout-dir", recipeCheckoutDir}, nil
		},

		// this is meaningless with a bundled recipe.
		"-revision": deleteFn,
	}

	newCmd := make([]string, 0, len(r.Properties.Command))

	skip := false
	for i, arg := range r.Properties.Command {
		if skip {
			skip = false
			continue
		}
		if fn, ok := toProcess[arg]; !ok {
			newCmd = append(newCmd, arg)
		} else {
			if strings.Contains(arg, "=") {
				toks := strings.SplitN(arg, "=", 2)
				replace, err := fn(toks[1])
				if err != nil {
					return nil, err
				}
				newCmd = append(newCmd, replace...)
			} else {
				if i+1 >= len(r.Properties.Command) {
					return nil, errors.
						Reason("%(arg)s in task definition, but no flag value found").
						D("arg", arg).
						Err()
				}
				replace, err := fn(r.Properties.Command[i+1])
				if err != nil {
					return nil, err
				}
				skip = true
				newCmd = append(newCmd, replace...)
			}
		}
	}

	ret.SwarmingTask.Properties.Command = newCmd

	return ret, nil
}

func updateMap(updates map[string]string, slc *[]*swarming.SwarmingRpcsStringPair) {
	if len(updates) == 0 {
		return
	}

	newSlice := make([]*swarming.SwarmingRpcsStringPair, 0, len(*slc)+len(updates))
	for k, v := range updates {
		if v != "" {
			newSlice = append(newSlice, &swarming.SwarmingRpcsStringPair{
				Key: k, Value: v})
		}
	}
	for _, pair := range *slc {
		if _, ok := updates[pair.Key]; !ok {
			newSlice = append(newSlice, pair)
		}
	}

	*slc = newSlice
}

func (jd *JobDefinition) Edit(dims, props, env map[string]string, bundleIso isolated.HexDigest, recipe string) (*JobDefinition, error) {
	if len(dims) == 0 && len(props) == 0 && len(env) == 0 && bundleIso == "" && recipe == "" {
		return jd, nil
	}

	ret := *jd
	ret.SwarmingTask = &(*jd.SwarmingTask)

	if bundleIso != "" {
		ret.RecipeIsolatedHash = bundleIso
	}

	if recipeName != "" {
		ret.RecipeName = recipe
	}

	updateMap(dims, &ret.SwarmingTask.Properties.Dimensions)
	updateMap(env, &ret.SwarmingTask.Properties.Env)

	if len(props) > 0 {
		ret.RecipeProperties = make(map[string]interface{}, len(jd.RecipeProperties)+len(props))
		for k, v := range props {
			if v != "" {
				var obj interface{}
				if err := json.NewDecoder(strings.NewReader(v)).Decode(&obj); err != nil {
					return nil, err
				}
				ret.RecipeProperties[k] = obj
			}
		}
		for k, v := range jd.RecipeProperties {
			if new, ok := props[k]; ok && new == "" {
				continue
			}
			ret.RecipeProperties[k] = v
		}
	}

	return &ret, nil
}

// GetSwarmingNewTask builds a usable SwarmingRpcsNewTaskRequest from the
// JobDefinition, incorporating all of the extra bits of the JobDefinition.
func (jd *JobDefinition) GetSwarmingNewTask(ctx context.Context, arc *archiver.Archiver) (*swarming.SwarmingRpcsNewTaskRequest, error) {
	st := *jd.SwarmingTask
	st.Properties = &(*st.Properties)

	// Copy+modify the command
	st.Properties.Command = append([]string(nil), st.Properties.Command...)
	var properties string
	for i, arg := range st.Properties.Command {
		switch arg {
		case recipePropertiesJSON:
			if properties == "" {
				propertiesBytes, err := json.Marshal(jd.RecipeProperties)
				if err != nil {
					return nil, err
				}
				properties = string(propertiesBytes)
			}
			st.Properties.Command[i] = properties

		case recipeName:
			st.Properties.Command[i] = jd.RecipeName
		}
	}

	// Inject the recipe bundle, or combine it with the existing isolate, if
	// necessary.
	if st.Properties.InputsRef != nil {
		toCombine := isolated.HexDigests{
			jd.RecipeIsolatedHash,
			isolated.HexDigest(st.Properties.InputsRef.Isolated),
		}
		newHash, err := combineIsolates(ctx, arc, toCombine...)
		if err != nil {
			return nil, err
		}
		st.Properties.InputsRef = &swarming.SwarmingRpcsFilesRef{
			Isolated: string(newHash),
		}
	} else {
		st.Properties.InputsRef = &swarming.SwarmingRpcsFilesRef{
			Isolated: string(jd.RecipeIsolatedHash),
		}
	}

	return &st, nil
}
