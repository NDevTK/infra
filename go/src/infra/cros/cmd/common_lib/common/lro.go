// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/chromiumos/config/go/longrunning"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/types/known/anypb"
)

// ProcessLro process a long running operation and parses the result to a proto
func ProcessLro(ctx context.Context, lro *longrunning.Operation) (*anypb.Any, error) {
	if lro == nil {
		return nil, fmt.Errorf("Provided lro is nil")
	}

	// Wait for the operation to be done
	for !lro.Done {
		time.Sleep(LroSleepTime)
	}

	// Check operation result
	switch x := lro.Result.(type) {
	case *longrunning.Operation_Error:
		logging.Infof(ctx, "LRO ERROR: %s", x.Error.Message)
		return nil, fmt.Errorf(x.Error.Message)
	case *longrunning.Operation_Response:
		logging.Infof(ctx, "LRO RESPONSE: %s", x.Response)
		return x.Response, nil
	default:
		return nil, fmt.Errorf("Unexpected lro result type")
	}
}
