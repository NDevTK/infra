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
	"github.com/luci/luci-go/common/logging"
)

const recipePropertiesJSON = "$RECIPE_PROPERTIES_JSON"
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

	RecipeProperties map[string]interface{} `json:"recipe_properties"`

	// TODO(iannucci):
	// this should really be a swarming.SwarmingRpcsNewTaskRequest, but the way
	// that buildbucket sends it is incompatible with the go endpoints generated
	// struct. Hooray...  *rollseyes*.
	SwarmingTask *swarming.SwarmingRpcsNewTaskRequest `json:"swarming_task"`
}

func JobDefinitionFromNewTaskRequest(r *swarming.SwarmingRpcsNewTaskRequest) (*JobDefinition, error) {
	ret := &JobDefinition{SwarmingTask: r}

	toProcess := map[string]func(nextTok string) (replace []string, err error){
		"-properties": func(nextTok string) ([]string, error) {
			ret.RecipeProperties = map[string]interface{}{}
			if err := json.NewDecoder(strings.NewReader(nextTok)).Decode(&ret.RecipeProperties); err != nil {
				return nil, errors.Annotate(err).Reason("decoding -properties JSON").Err()
			}
			return []string{"-properties", recipePropertiesJSON}, nil
		},

		"-repository": func(nextTok string) ([]string, error) {
			return []string{"-checkout-dir", recipeCheckoutDir}, nil
		},

		"-revision": func(nextTok string) ([]string, error) {
			return nil, nil
		},
	}

	newCmd := make([]string, 0, len(r.Properties.Command))

	skip := false
	for i, arg := range r.Properties.Command {
		if skip {
			skip = false
			continue
		}
		if fn, ok := toProcess[arg]; ok {
			if i+1 >= len(r.Properties.Command) {
				return nil, errors.
					Reason("%s in task definition, but no following json property data").
					D("arg", arg).
					Err()
			}
			replace, err := fn(r.Properties.Command[i+1])
			if err != nil {
				return nil, err
			}
			skip = true
			newCmd = append(newCmd, replace...)
		} else {
			newCmd = append(newCmd, arg)
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

func (jd *JobDefinition) Edit(dims, props, env map[string]string, bundleIso isolated.HexDigest) (*JobDefinition, error) {
	if len(dims) == 0 && len(props) == 0 && len(env) == 0 && bundleIso == "" {
		return jd, nil
	}

	ret := *jd
	ret.SwarmingTask = &(*jd.SwarmingTask)

	if bundleIso != "" {
		ret.RecipeIsolatedHash = bundleIso
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

func (jd *JobDefinition) GetSwarmingNewTask(ctx context.Context, arc *archiver.Archiver) (*swarming.SwarmingRpcsNewTaskRequest, error) {
	st := *jd.SwarmingTask
	st.Properties = &(*st.Properties)

	toCombine := isolated.HexDigests{
		jd.RecipeIsolatedHash,
	}

	if st.Properties.InputsRef != nil {
		toCombine = append(toCombine,
			isolated.HexDigest(st.Properties.InputsRef.Isolated))
	}

	cmd := make([]string, len(st.Properties.Command))
	copy(cmd, st.Properties.Command)

	var properties string
	for i, arg := range cmd {
		if arg == recipePropertiesJSON {
			if properties == "" {
				propertiesBytes, err := json.Marshal(jd.RecipeProperties)
				if err != nil {
					return nil, err
				}
				properties = string(propertiesBytes)
			}
			cmd[i] = properties
		}
	}

	logging.Infof(ctx, "combining: %v %v", cmd, toCombine)
	newHash, err := combineIsolates(ctx, arc, cmd, toCombine...)
	if err != nil {
		return nil, err
	}
	st.Properties.InputsRef = &swarming.SwarmingRpcsFilesRef{
		Isolated: string(newHash),
	}
	st.Properties.Command = nil

	return &st, nil
}
