// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dynamic_updates

import (
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/common_lib/dynamic_updates/finders"
	"infra/cros/cmd/common_lib/dynamic_updates/resolver"
	"infra/cros/cmd/common_lib/dynamic_updates/updaters"
)

// AddUserDefinedDynamicUpdates resolves placeholders and applies
// the dynamic updates to the trv2 dynamic request.
func AddUserDefinedDynamicUpdates(req *api.CrosTestRunnerDynamicRequest, dynamicUpdates []*api.UserDefinedDynamicUpdate, lookup map[string]string) error {
	for _, dynamicUpdate := range dynamicUpdates {
		dynamicUpdate, err := resolver.Resolve(dynamicUpdate, lookup)
		if err != nil {
			return errors.Annotate(err, "failed to resolve dynamic update placeholders").Err()
		}

		focalIndex, err := finders.GetFocalTaskFinder(req, dynamicUpdate.FocalTaskFinder)
		if err != nil {
			return errors.Annotate(err, "failed to get relative position for dynamic update, %v", dynamicUpdate).Err()
		}

		err = updaters.ProcessUpdateAction(req, dynamicUpdate.UpdateAction, focalIndex)
		if err != nil {
			return errors.Annotate(err, "failed to process update action for dynamic update, %v", dynamicUpdate).Err()
		}
	}

	return nil
}
