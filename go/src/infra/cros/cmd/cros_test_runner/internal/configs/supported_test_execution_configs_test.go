// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package configs

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateHwConfigs(t *testing.T) {
	t.Parallel()
	Convey("GenerateHwConfigs", t, func() {
		ctx := context.Background()
		hwConfigs := GenerateHwConfigs(ctx, nil)

		So(hwConfigs, ShouldNotBeNil)
		So(hwConfigs.MainConfigs, ShouldNotBeNil)
		So(hwConfigs.CleanupConfigs, ShouldNotBeNil)
		So(len(hwConfigs.MainConfigs), ShouldBeGreaterThan, 0)
		So(len(hwConfigs.CleanupConfigs), ShouldBeGreaterThan, 0)
	})
}
