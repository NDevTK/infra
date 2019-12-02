// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	swarmingpb "go.chromium.org/luci/swarming/proto/api"
)

func casTreeFromSwarming(iso *swarming.SwarmingRpcsFilesRef) *swarmingpb.CASTree {
	if iso == nil {
		return nil
	}
	return &swarmingpb.CASTree{
		Digest:    iso.Isolated,
		Namespace: iso.Namespace,
		Server:    iso.Isolatedserver,
	}
}

func cipdPkgsFromSwarming(pkgs *swarming.SwarmingRpcsCipdInput) []*swarmingpb.CIPDPackage {
	if pkgs == nil || len(pkgs.Packages) == 0 {
		return nil
	}
	// TODO(iannucci): In practice we thought that ClientPackage would be useful
	// for users, but it turns out that it isn't. Practically, the version of cipd
	// to use for the swarming bot is (and should be) entirely driven by the
	// implementation of the swarming bot.
	//
	// The task must supply its own copy of cipd ANYWAY (if the task wants to use
	// cipd).
	//
	// Thus, we ignore the ClientPackage field.
	ret := make([]*swarmingpb.CIPDPackage, len(pkgs.Packages))
	for i, in := range pkgs.Packages {
		ret[i] = &swarmingpb.CIPDPackage{
			DestPath:    in.Path,
			PackageName: in.PackageName,
			Version:     in.Version,
		}
	}

	return ret
}

func namedCachesFromSwarming(caches []*swarming.SwarmingRpcsCacheEntry) []*swarmingpb.NamedCacheEntry {
	if len(caches) == 0 {
		return nil
	}
	ret := make([]*swarmingpb.NamedCacheEntry, len(caches))
	for i, in := range caches {
		ret[i] = &swarmingpb.NamedCacheEntry{
			Name:     in.Name,
			DestPath: in.Path,
		}
	}
	return ret
}

func dimensionsFromSwarming(dims []*swarming.SwarmingRpcsStringPair) []*swarmingpb.StringListPair {
	if len(dims) == 0 {
		return nil
	}

	intermediate := map[string][]string{}
	for _, dim := range dims {
		intermediate[dim.Key] = append(intermediate[dim.Key], dim.Value)
	}

	ret := make([]*swarmingpb.StringListPair, 0, len(intermediate))
	for key, values := range intermediate {
		ret = append(ret, &swarmingpb.StringListPair{Key: key, Values: values})
	}
	return ret
}

func envFromSwarming(env []*swarming.SwarmingRpcsStringPair) []*swarmingpb.StringPair {
	if len(env) == 0 {
		return nil
	}
	ret := make([]*swarmingpb.StringPair, len(env))
	for i, in := range env {
		ret[i] = &swarmingpb.StringPair{Key: in.Key, Value: in.Value}
	}
	return ret
}

func envPrefixesFromSwarming(envPrefixes []*swarming.SwarmingRpcsStringListPair) []*swarmingpb.StringListPair {
	if len(envPrefixes) == 0 {
		return nil
	}
	ret := make([]*swarmingpb.StringListPair, len(envPrefixes))
	for i, in := range envPrefixes {
		ret[i] = &swarmingpb.StringListPair{Key: in.Key, Values: in.Value}
	}
	return ret
}

func containmentFromSwarming(con *swarming.SwarmingRpcsContainment) *swarmingpb.Containment {
	if con == nil {
		return nil
	}
	conType, ok := swarmingpb.Containment_ContainmentType_value[con.ContainmentType]
	if !ok {
		// TODO(iannucci): handle this more gracefully?
		//
		// This is a relatively unused field, and I don't expect any divergence
		// between the proto / endpoints definitions...  hopefully by the time we
		// touch this swarming has a real prpc api and then this entire file can go
		// away.
		panic(fmt.Sprintf("unknown containment type %q", con.ContainmentType))
	}
	return &swarmingpb.Containment{
		ContainmentType:           swarmingpb.Containment_ContainmentType(conType),
		LimitProcesses:            con.LimitProcesses,
		LimitTotalCommittedMemory: con.LimitTotalCommittedMemory,
		LowerPriority:             con.LowerPriority,
	}
}

func taskPropertiesFromSwarming(ts *swarming.SwarmingRpcsTaskProperties) *swarmingpb.TaskProperties {
	// TODO(iannucci): log that we're dropping SecretBytes?

	return &swarmingpb.TaskProperties{
		CasInputs:   casTreeFromSwarming(ts.InputsRef),
		CipdInputs:  cipdPkgsFromSwarming(ts.CipdInput),
		NamedCaches: namedCachesFromSwarming(ts.Caches),
		Command:     ts.Command,
		RelativeCwd: ts.RelativeCwd,
		ExtraArgs:   ts.ExtraArgs,
		// SecretBytes/HasSecretBytes are not provided by the swarming server.
		Dimensions:       dimensionsFromSwarming(ts.Dimensions),
		Env:              envFromSwarming(ts.Env),
		EnvPaths:         envPrefixesFromSwarming(ts.EnvPrefixes),
		Containment:      containmentFromSwarming(ts.Containment),
		ExecutionTimeout: ptypes.DurationProto(time.Duration(ts.ExecutionTimeoutSecs) * time.Second),
		IoTimeout:        ptypes.DurationProto(time.Duration(ts.IoTimeoutSecs) * time.Second),
		GracePeriod:      ptypes.DurationProto(time.Duration(ts.GracePeriodSecs) * time.Second),
		Idempotent:       ts.Idempotent,
		Outputs:          ts.Outputs,
	}
}

func jobDefinitionFromSwarming(sw *Swarming, r *swarming.SwarmingRpcsNewTaskRequest) {
	// we ignore r.Properties; TaskSlices are the only thing generated from modern
	// swarming tasks.
	sw.Task = &swarmingpb.TaskRequest{}
	t := sw.Task

	inslices := r.TaskSlices
	if len(inslices) == 0 {
		inslices = []*swarming.SwarmingRpcsTaskSlice{{
			ExpirationSecs: r.ExpirationSecs,
			Properties:     r.Properties,
		}}
	}
	t.TaskSlices = make([]*swarmingpb.TaskSlice, len(inslices))

	for i, inslice := range inslices {
		outslice := &swarmingpb.TaskSlice{
			Expiration:      ptypes.DurationProto(time.Duration(inslice.ExpirationSecs) * time.Second),
			Properties:      taskPropertiesFromSwarming(inslice.Properties),
			WaitForCapacity: inslice.WaitForCapacity,
		}
		t.TaskSlices[i] = outslice
	}

	t.Priority = int32(r.Priority)
	t.ServiceAccount = r.ServiceAccount
	// CreateTime is unused for new task requests
	// Name is overwritten
	t.Tags = r.Tags
	t.User = r.User
	// TaskId is unpopulated for new task requests
	// ParentTaskId is unpopulated for new task requests
	// ParentRunId is unpopulated for new task requests
	// PubsubNotification is intentionally not propagated.
}
