//go:build linux
// +build linux

// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"infra/cros/recovery/internal/components/mocks"
	"infra/cros/recovery/logger"
)

func getBaseTestRequest(installThroughServo bool) *InstallFirmwareImageRequest {
	return &InstallFirmwareImageRequest{
		Board:             "my-board",
		Model:             "my-model",
		FlashThroughServo: installThroughServo,
	}
}

func TestExtractECImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logger.NewLogger()
	ctrl := gomock.NewController(t)
	tarballPath := "/some/folder/my_folder/tarbar.tr"
	Convey("Happy path", t, func() {
		req := getBaseTestRequest(true)
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "ec_board").Return(stringValue("s-Board"), nil).Times(1)
		req.Servod = servod
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/EC": "",
			"tar tf /some/folder/my_folder/tarbar.tr s-board/ec.bin my-model/ec.bin my-board/ec.bin ec.bin": `ec.bin
my-board/ec.bin`,
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC ec.bin": "",
		}
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractECImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/EC/ec.bin")
	})
	Convey("Happy path with board file", t, func() {
		req := getBaseTestRequest(true)
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "ec_board").Return(stringValue("s-Board"), nil).Times(1)
		req.Servod = servod
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/EC": "",
			"tar tf /some/folder/my_folder/tarbar.tr s-board/ec.bin my-model/ec.bin my-board/ec.bin ec.bin": `my-ec.bin
my-board/ec.bin`,
			"tar tf /some/folder/my_folder/tarbar.tr s-board/npcx_monitor.bin my-model/npcx_monitor.bin my-board/npcx_monitor.bin npcx_monitor.bin": ``,
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC my-board/ec.bin":                                                  "",
		}
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractECImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/EC/my-board/ec.bin")
	})
	Convey("Happy path with board file with monitor", t, func() {
		req := getBaseTestRequest(true)
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "ec_board").Return(stringValue("s-Board"), nil).Times(1)
		req.Servod = servod
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/EC": "",
			"tar tf /some/folder/my_folder/tarbar.tr s-board/ec.bin my-model/ec.bin my-board/ec.bin ec.bin": `my-ec.bin
my-board/ec.bin
npcx_monitor.bin`,
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC my-board/ec.bin":  "",
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC npcx_monitor.bin": "",
		}
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractECImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/EC/my-board/ec.bin")
	})
	Convey("Happy path without servod", t, func() {
		req := getBaseTestRequest(true)
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/EC": "",
			"tar tf /some/folder/my_folder/tarbar.tr my-model/ec.bin my-board/ec.bin ec.bin": `my-ec.bin
my-board/ec.bin
npcx_monitor.bin`,
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC my-board/ec.bin":  "",
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC npcx_monitor.bin": "",
		}
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractECImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/EC/my-board/ec.bin")
	})
	Convey("Happy path run from DUT", t, func() {
		req := getBaseTestRequest(false)
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/EC": "",
			"tar tf /some/folder/my_folder/tarbar.tr my-model/ec.bin my-board/ec.bin ec.bin": `my-ec.bin
my-board/ec.bin
npcx_monitor.bin`,
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC my-board/ec.bin":  "",
			"tar xf /some/folder/my_folder/tarbar.tr -C /some/folder/my_folder/EC npcx_monitor.bin": "",
		}
		req.DutRunner = mockRunner(runRequest)
		image, err := extractECImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/EC/my-board/ec.bin")
	})
}

func TestExtractAPImage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logger.NewLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tarballPath := "/some/folder/my_folder/tarbar2.tr"
	Convey("Happy path", t, func() {
		req := getBaseTestRequest(true)
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/AP": "",
			"tar tf /some/folder/my_folder/tarbar2.tr image-s-board.bin image-my-model.bin image-my-board.bin image.bin": `image.bin
image-my-model.bin`,
			"tar xf /some/folder/my_folder/tarbar2.tr -C /some/folder/my_folder/AP image.bin": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "ec_board").Return(stringValue("s-Board"), nil).Times(1)
		req.Servod = servod
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractAPImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/AP/image.bin")
	})
	Convey("Happy path with board file", t, func() {
		req := getBaseTestRequest(true)
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/AP": "",
			"tar tf /some/folder/my_folder/tarbar2.tr image-s-board.bin image-my-model.bin image-my-board.bin image.bin": `image-my.bin
image-my-model.bin`,
			"tar xf /some/folder/my_folder/tarbar2.tr -C /some/folder/my_folder/AP image-my-model.bin": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Get(ctx, "ec_board").Return(stringValue("S-board"), nil).Times(1)
		req.Servod = servod
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractAPImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/AP/image-my-model.bin")
	})
	Convey("Happy path without servod", t, func() {
		req := getBaseTestRequest(true)
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/AP": "",
			"tar tf /some/folder/my_folder/tarbar2.tr image-my-model.bin image-my-board.bin image.bin": `image-my.bin
image-my-model.bin`,
			"tar xf /some/folder/my_folder/tarbar2.tr -C /some/folder/my_folder/AP image-my-model.bin": "",
		}
		req.ServoHostRunner = mockRunner(runRequest)
		image, err := extractAPImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/AP/image-my-model.bin")
	})
	Convey("Happy path run from DUT", t, func() {
		req := getBaseTestRequest(false)
		runRequest := map[string]string{
			"mkdir -p /some/folder/my_folder/AP": "",
			"tar tf /some/folder/my_folder/tarbar2.tr image-my-model.bin image-my-board.bin image.bin": `image-my.bin
image-my-model.bin`,
			"tar xf /some/folder/my_folder/tarbar2.tr -C /some/folder/my_folder/AP image-my-model.bin": "",
		}
		req.DutRunner = mockRunner(runRequest)
		image, err := extractAPImage(ctx, req, tarballPath, logger)
		So(err, ShouldBeNil)
		So(image, ShouldEqual, "/some/folder/my_folder/AP/image-my-model.bin")
	})
}
