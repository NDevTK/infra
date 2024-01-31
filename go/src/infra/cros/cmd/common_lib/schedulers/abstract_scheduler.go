// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package schedulers

import (
	"infra/cros/cmd/common_lib/interfaces"
)

// AbstractScheduler defines abstract scheduler that other schedulers can
// extend.
type AbstractScheduler struct {
	interfaces.SchedulerInterface

	schedulerType interfaces.SchedulerType
}

func NewAbstractScheduler(scType interfaces.SchedulerType) *AbstractScheduler {
	return &AbstractScheduler{schedulerType: scType}
}

func (sc *AbstractScheduler) GetSchedulerType() interfaces.SchedulerType {
	return sc.schedulerType
}
