// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package labpack

import (
	"infra/libs/skylab/buildbucket"
)

type Params = buildbucket.Params

type CIPDVersion = buildbucket.CIPDVersion

const (
	CIPDProd   CIPDVersion = buildbucket.CIPDProd
	CIPDLatest CIPDVersion = buildbucket.CIPDLatest
)

var ScheduleTask = buildbucket.ScheduleTask
