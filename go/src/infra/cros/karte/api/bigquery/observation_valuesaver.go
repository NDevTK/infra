// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package kbqpb

import (
	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
)

// Save populates a row, which is a string that points to interface{}'s and produces an insertID, which is, in this case,
// a no-op.
func (o *Observation) Save() (map[string]bigquery.Value, string, error) {
	if o == nil {
		return nil, "", errors.Reason("action save: refusing to insert empty action").Err()
	}
	if o.GetName() == "" {
		return nil, "", errors.Reason("action save: refusing to insert action with empty name").Err()
	}

	// NOTE: The NoDedupeID feature is experimental.
	insertID := bigquery.NoDedupeID

	row := make(map[string]bigquery.Value)
	row["name"] = o.GetName()
	row["action_name"] = o.GetActionName()
	row["metric_kind"] = o.GetMetricKind()
	row["type"] = o.GetType()
	row["value_string"] = o.GetValueString()
	row["value_number"] = o.GetValueNumber()

	return row, insertID, nil
}
