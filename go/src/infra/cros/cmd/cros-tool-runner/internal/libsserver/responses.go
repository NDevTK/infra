// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Implements test_libs_service.proto (see proto for details)
package libsserver

import (
	pb "go.chromium.org/chromiumos/config/go/test/api"
)

func responseSuccess(id, port string) *pb.GetLibResponse {
	return &pb.GetLibResponse{
		Outcome: &pb.GetLibResponse_Success{
			Success: &pb.GetLibSuccess{
				Id:   id,
				Port: port,
			},
		},
	}
}
func responseFailure(reason pb.GetLibFailure_Reason) *pb.GetLibResponse {
	return &pb.GetLibResponse{
		Outcome: &pb.GetLibResponse_Failure{
			Failure: &pb.GetLibFailure{
				Reason: reason,
			},
		},
	}
}
