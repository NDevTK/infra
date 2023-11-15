// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/mocks"
	"infra/cros/recovery/logger"
)

func TestProgrammerV3ProgramEC(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logger.NewLogger()
	imagePath := "ec_image.bin"
	Convey("Happy path for stm32 chip", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
			"flash_ec --chip=stm32 --image=ec_image.bin --port=95 --bitbang_rate=57600 --verify --verbose": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("stm32"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_servo_micro_and_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(95).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldBeNil)
	})
	Convey("Happy path for other chips", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
			"flash_ec --chip=some_chip --image=ec_image.bin --port=96 --verify --verbose": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("some_chip"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_servo_micro_and_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(96).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldBeNil)
	})
	Convey("block for ite chips (1)", t, func() {
		// TODO(b:270170790): remove when we can recover functionality fo flash EC for ite chips.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(0)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(0)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("it8XXXX"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(95).Times(0)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldNotBeNil)
	})
	Convey("block for ite chips (2)", t, func() {
		// TODO(b:270170790): remove when we can recover functionality fo flash EC for ite chips.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(0)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(0)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("it8yyyyy"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(95).Times(0)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldNotBeNil)
	})
	Convey("allowed for ite chips if uses servo_micro", t, func() {
		// TODO(b:270170790): remove when we can recover functionality fo flash EC for ite chips.
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
			"flash_ec --chip=it8yyyyy --image=ec_image.bin --port=96 --verify --verbose": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("it8yyyyy"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_servo_micro_and_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(96).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldBeNil)
	})
	Convey("use try_apshutdown is expected for ccd_cpu_fw_spi_depends_on_ec_fw:yes (1)", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
			"flash_ec --chip=just_chip --image=ec_image.bin --port=96 --verify --verbose --try_apshutdown": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("yes"), nil).Times(1)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("just_chip"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_servo_micro_and_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(96).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldBeNil)
	})
	Convey("use try_apshutdown is expected for cpu_fw_spi_depends_on_ec_fw:yes (2)", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
			"flash_ec --chip=just_chip --image=ec_image.bin --port=96 --verify --verbose --try_apshutdown": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(nil).Times(1)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("yes"), nil).Times(1)
		// not called as above already told yes.
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(nil).Times(0)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("just_chip"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_servo_micro_and_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(96).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldBeNil)
	})
	Convey("do not use try_apshutdown if controls are not present", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		runRequest := map[string]string{
			"which flash_ec": "",
			"flash_ec --chip=just_chip --image=ec_image.bin --port=96 --verify --verbose": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Has(ctx, "cpu_fw_spi").Return(errors.Reason("Not present").Err()).Times(1)
		servod.EXPECT().Has(ctx, "ccd_cpu_fw_spi").Return(errors.Reason("Not present").Err()).Times(1)
		servod.EXPECT().Get(ctx, "cpu_fw_spi_depends_on_ec_fw").Return(stringValue("yes"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ccd_cpu_fw_spi_depends_on_ec_fw").Return(stringValue("no"), nil).Times(0)
		servod.EXPECT().Get(ctx, "ec_chip").Return(stringValue("just_chip"), nil).Times(1)
		servod.EXPECT().Get(ctx, "servo_type").Return(stringValue("servo_v4_with_servo_micro_and_ccd_cr50"), nil).Times(1)
		servod.EXPECT().Port().Return(96).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programEC(ctx, imagePath)
		So(err, ShouldBeNil)
	})
}

func TestProgrammerV3ProgramAP(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()
	logger := logger.NewLogger()
	imagePath := "image-board.bin"
	Convey("Happy path", t, func() {
		runRequest := map[string]string{
			"which futility": "",
			"futility update -i image-board.bin --servo_port=97": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Port().Return(97).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programAP(ctx, imagePath, "", false)
		So(err, ShouldBeNil)
	})
	Convey("Happy path with GBB 0x18", t, func() {
		runRequest := map[string]string{
			"which futility": "",
			"futility update -i image-board.bin --servo_port=91 --gbb_flags=0x18": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Port().Return(91).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programAP(ctx, imagePath, "0x18", false)
		So(err, ShouldBeNil)
	})
	Convey("Happy path with force update", t, func() {
		runRequest := map[string]string{
			"which futility": "",
			"futility update -i image-board.bin --servo_port=97 --force": "",
		}
		servod := mocks.NewMockServod(ctrl)
		servod.EXPECT().Port().Return(97).Times(1)

		p := &v3Programmer{
			run:    mockRunner(runRequest),
			servod: servod,
			log:    logger,
		}

		err := p.programAP(ctx, imagePath, "", true)
		So(err, ShouldBeNil)
	})
}

var gbbToIntCases = []struct {
	name string
	in   string
	out  int
	fail bool
}{
	{"Empty value", "", 0, true},
	{"Incorrect value", "raw", 0, true},
	{"GBB 0", "0", 0, false},
	{"GBB 0x0", "0x0", 0, false},
	{"GBB 0x1", "0x1", 1, false},
	{"GBB 0x8", "0x8", 8, false},
	{"GBB 0x18", "0x18", 24, false},
	{"GBB 0x24", "0x24", 36, false},
	{"GBB 0x39", "0x39", 57, false},
	{"GBB 0x00000039", "0x00000039", 57, false},
}

func TestGbbToInt(t *testing.T) {
	t.Parallel()
	for _, c := range gbbToIntCases {
		cs := c
		t.Run(cs.name, func(t *testing.T) {
			got, err := gbbToInt(cs.in)
			if cs.fail {
				if err == nil {
					t.Errorf("%q -> expected to fail but passed", cs.name)
				}
			} else {
				if err != nil {
					t.Errorf("%q -> expected to pass by fail %s", cs.name, err)
				} else if got != cs.out {
					t.Errorf("%q -> wanted: %v => got: %v", cs.name, cs.out, got)
				}
			}
		})
	}
}
