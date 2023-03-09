// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (req *TriggerCronJobReq) Validate() error {
	if req.JobName == "" {
		return status.Errorf(codes.InvalidArgument, "Need cron job name to trigger")
	}
	return nil
}
