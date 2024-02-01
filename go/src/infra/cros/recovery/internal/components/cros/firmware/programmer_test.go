// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/config/go/api/test/xmlrpc"
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/mocks"
	"infra/cros/recovery/logger"
)

func TestNewProgrammer(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	logger := logger.NewLogger()
	Convey("Fail if servod fail to respond to servod", t, func() {
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "servo_type").Return(nil, errors.Reason("fail to get servo_type!").Err()).Times(1)
		run, runCounter := mockRunnerWithCheck(nil)
		p, err := NewProgrammer(ctx, run, servod, logger)
		So(p, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(runCounter(), ShouldEqual, 0)
	})
	Convey("Fail as servo_v2 is not supported", t, func() {
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v2"), nil).Times(1)
		run, runCounter := mockRunnerWithCheck(nil)
		p, err := NewProgrammer(ctx, run, servod, logger)
		So(p, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(runCounter(), ShouldEqual, 0)
	})
	Convey("Creates programmer for servo_v3", t, func() {
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v3"), nil).Times(1)
		run, runCounter := mockRunnerWithCheck(nil)
		p, err := NewProgrammer(ctx, run, servod, logger)
		So(p, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(runCounter(), ShouldEqual, 0)
	})
	Convey("Creates programmer for servo_v4", t, func() {
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4"), nil).Times(1)
		run, runCounter := mockRunnerWithCheck(nil)
		p, err := NewProgrammer(ctx, run, servod, logger)
		So(p, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(runCounter(), ShouldEqual, 0)
	})
}

func stringValue(v string) *xmlrpc.Value {
	return &xmlrpc.Value{
		ScalarOneof: &xmlrpc.Value_String_{
			String_: v,
		},
	}
}

func mockRunner(runResponses map[string]string) components.Runner {
	run, _ := mockRunnerWithCheck(runResponses)
	return run
}
func mockRunnerWithCheck(runResponses map[string]string) (components.Runner, func() int) {
	calls := make(map[string]bool)
	return func(ctx context.Context, timeout time.Duration, cmd string, args ...string) (string, error) {
			cmd = strings.Join(append([]string{cmd}, args...), " ")
			// Mark that call was done.
			calls[cmd] = true
			if v, ok := runResponses[cmd]; ok {
				return v, nil
			}
			return "", errors.Reason("Did not found response for %q!", cmd).Err()
		}, func() int {
			return len(calls)
		}
}
