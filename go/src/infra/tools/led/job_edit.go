// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

type EditJobDefinition interface {
	CipdPkgs(cipdPkgs map[string]string)
	Dimensions(dims map[string]string)
	Env(env map[string]string)
	Experimental(trueOrFalse string)
	PrefixPathEnv(values []string)
	Priority(priority int32)
	Properties(props map[string]string, auto bool)
	Recipe(recipe string)
	RecipeSource(isolated, cipdPkg, cipdVer string)
	SwarmingHostname(host string)
	Tags(values []string)
}
