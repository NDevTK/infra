// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"context"
	"infra/cros/recovery/logger"
)

// Record records an action and returns the action that was just recorded.
// Note that an action contains zero or more observations in it and that observations are not
// separate.
func (c *kClient) Record(ctx context.Context, action *logger.Action) (*logger.Action, error) {
	panic("not implemented")
}

// Search takes a query struct and produces a resultset.
func (c *kClient) Search(ctx context.Context, q *logger.Query) (*logger.QueryResult, error) {
	panic("not implemented")
}
