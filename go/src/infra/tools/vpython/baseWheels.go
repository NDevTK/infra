// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"go.chromium.org/luci/vpython/api/vpython"
)

// matchMac is a PEP425 match tag that will match all modern OSX platforms.
var matchMac = &vpython.PEP425Tag{Platform: "macosx_10_10_x86_64"}

// baseWheels is the collection of wheels that will always be installed in every
// user VirtualEnv.
//
// Match tags can be used to mark system-specific wheels.
var baseWheels = []*vpython.Spec_Package{
	{Name: "infra/python/wheels/psutil/${platform}_${py_python}_${py_abi}", Version: "version:5.2.2"},

	// On Mac OSX, we
	{Name: "infra/python/wheels/pyobjc/${platform}_${py_python}_${py_abi}", Version: "version:3.2.1",
		MatchTag: []*vpython.PEP425Tag{matchMac}},
}
