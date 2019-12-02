// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"infra/tools/kitchen/cookflags"
	"io"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/luci/buildbucket/cmd/bbagent/bbinput"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/errors"
	swarmingpb "go.chromium.org/luci/swarming/proto/api"
)

// JobDefinitionFromNewTaskRequest generates a new JobDefinition by parsing the
// given SwarmingRpcsNewTaskRequest.
//
// If the task's first slice looks like either a bbagent or kitchen-based
// Buildbucket task, the returned JobDefinition will have the `buildbucket`
// field populated, otherwise the `swarming` field will be populated.
func JobDefinitionFromNewTaskRequest(r *swarming.SwarmingRpcsNewTaskRequest, name, swarmingHost string) (*JobDefinition, error) {
	arg0, ts := "", &swarming.SwarmingRpcsTaskSlice{}
	if len(r.TaskSlices) > 0 {
		ts = r.TaskSlices[0]
	} else {
		ts.ExpirationSecs = r.ExpirationSecs
		ts.Properties = r.Properties
	}
	if ts.Properties != nil {
		if len(ts.Properties.Command) > 0 {
			arg0 = ts.Properties.Command[0]
		}
	}

	ret := &JobDefinition{}
	name = "led: " + name

	switch arg0 {
	case "bbagent${EXECUTABLE_SUFFIX}":
		bb := &Buildbucket{Name: name}
		ret.JobType = &JobDefinition_Buildbucket{Buildbucket: bb}
		if err := jobDefinitionFromBuildbucket(bb, r, ts); err != nil {
			return nil, err
		}

	case "kitchen${EXECUTABLE_SUFFIX}":
		bb := &Buildbucket{LegacyKitchen: true, Name: name}
		ret.JobType = &JobDefinition_Buildbucket{Buildbucket: bb}
		if err := jobDefinitionFromBuildbucketLegacy(bb, r, ts); err != nil {
			return nil, err
		}

	default:
		// non-Buildbucket Swarming task
		sw := &Swarming{Hostname: swarmingHost}
		ret.JobType = &JobDefinition_Swarming{Swarming: sw}
		jobDefinitionFromSwarming(sw, r)
		sw.Task.Name = name
	}

	if bb := ret.GetBuildbucket(); bb != nil {
		// set all buildbucket type tasks to experimental by default.
		bb.BbagentArgs.Build.Input.Experimental = true
	}

	return ret, nil
}

func jobDefinitionFromBuildbucket(bb *Buildbucket, r *swarming.SwarmingRpcsNewTaskRequest, ts *swarming.SwarmingRpcsTaskSlice) (err error) {
	bb.BbagentArgs, err = bbinput.Parse(ts.Properties.Command[1])
	if err != nil {
		return
	}

	bb.ScrubIncomingData()

	bb.CipdPackages = cipdPins(ts.Properties.CipdInput)
	bb.EnvVars = strPairs(ts.Properties.Env)
	bb.EnvPrefixes = strListPairs(ts.Properties.EnvPrefixes)

	return
}

func jobDefinitionFromBuildbucketLegacy(bb *Buildbucket, r *swarming.SwarmingRpcsNewTaskRequest, ts *swarming.SwarmingRpcsTaskSlice) error {
	var kitchenArgs cookflags.CookFlags
	fs := flag.NewFlagSet("kitchen_cook", flag.ContinueOnError)
	kitchenArgs.Register(fs)
	if err := fs.Parse(ts.Properties.Command[2:]); err != nil {
		return errors.Annotate(err, "parsing kitchen cook args").Err()
	}

	bb.BbagentArgs = &bbpb.BBAgentArgs{
		CacheDir:               kitchenArgs.CacheDir,
		KnownPublicGerritHosts: ([]string)(kitchenArgs.KnownGerritHost),
		Build:                  &bbpb.Build{},
	}

	// kitchen builds are sorta inverted; the Build message is in the buildbucket
	// module property, but it doesn't contain the properties in input.
	const bbModPropKey = "$recipe_engine/buildbucket"
	bbModProps := kitchenArgs.Properties[bbModPropKey].(map[string]interface{})
	delete(kitchenArgs.Properties, bbModPropKey)

	pipeR, pipeW := io.Pipe()
	done := make(chan error)
	go func() {
		done <- jsonpb.Unmarshal(pipeR, bb.BbagentArgs.Build)
	}()
	if err := json.NewEncoder(pipeW).Encode(bbModProps["build"]); err != nil {
		return errors.Annotate(err, "%s['build'] -> json", bbModPropKey).Err()
	}
	if err := <-done; err != nil {
		return errors.Annotate(err, "%s['build'] -> jsonpb", bbModPropKey).Err()
	}
	bb.ScrubIncomingData()

	err := jsonpb.UnmarshalString(kitchenArgs.Properties.String(),
		bb.BbagentArgs.Build.Input.Properties)
	return errors.Annotate(err, "populating properties").Err()
}

// Private stuff

func cipdPins(ci *swarming.SwarmingRpcsCipdInput) (ret []*swarmingpb.CIPDPackage) {
	if ci == nil {
		return
	}
	ret = make([]*swarmingpb.CIPDPackage, 0, len(ci.Packages))
	for _, pkg := range ci.Packages {
		ret = append(ret, &swarmingpb.CIPDPackage{
			PackageName: pkg.PackageName,
			Version:     pkg.Version,
			DestPath:    pkg.Path,
		})
	}
	return
}

func strPairs(pairs []*swarming.SwarmingRpcsStringPair) []*swarmingpb.StringPair {
	ret := make([]*swarmingpb.StringPair, len(pairs))
	for i, p := range pairs {
		ret[i] = &swarmingpb.StringPair{Key: p.Key, Value: p.Value}
	}
	return ret
}

func strListPairs(pairs []*swarming.SwarmingRpcsStringListPair) []*swarmingpb.StringListPair {
	ret := make([]*swarmingpb.StringListPair, len(pairs))
	for i, p := range pairs {
		vals := make([]string, len(p.Value))
		copy(vals, p.Value)
		ret[i] = &swarmingpb.StringListPair{Key: p.Key, Values: vals}
	}
	return ret
}

// swarming has two separate structs to represent a task request.
//
// Convert from 'TaskRequest' to 'NewTaskRequest'.
func taskRequestToNewTaskRequest(req *swarming.SwarmingRpcsTaskRequest) *swarming.SwarmingRpcsNewTaskRequest {
	return &swarming.SwarmingRpcsNewTaskRequest{
		Name:           req.Name,
		ExpirationSecs: req.ExpirationSecs,
		Priority:       req.Priority,
		Properties:     req.Properties,
		TaskSlices:     req.TaskSlices,
		// don't wan't these or some random person/service will get notified :
		//PubsubTopic:    req.PubsubTopic,
		//PubsubUserdata: req.PubsubUserdata,
		Tags:           req.Tags,
		User:           req.User,
		ServiceAccount: req.ServiceAccount,
	}
}
