// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHasServo(t *testing.T) {
	t.Parallel()
	satlabId := "satlab123"
	Convey("Register servo for labstation", t, func() {
		ad := &addDUT{
			shivasAddDUT: shivasAddDUT{
				servo:       "servo_1",
				servoSerial: "servo_serial",
			},
		}
		if yes := ad.setupServoArguments(satlabId); !yes {
			t.Errorf("Expected servo is not detected but expected!")
		}
		So(ad.qualifiedServo, ShouldEqual, "satlab-satlab123-servo_1")
		So(ad.servoDockerContainerName, ShouldEqual, "")
	})
	Convey("Register servo for container", t, func() {
		ad := &addDUT{
			shivasAddDUT: shivasAddDUT{
				servo:       "",
				servoSerial: "servo_serial",
			},
		}
		if yes := ad.setupServoArguments(satlabId); !yes {
			t.Errorf("Expected servo is not detected but expected!")
		}
		So(ad.qualifiedServo, ShouldEqual, "satlab-satlab123--docker_servod:9999")
		So(ad.servoDockerContainerName, ShouldEqual, "satlab-satlab123--docker_servod")
	})
	Convey("Servo-less setup", t, func() {
		ad := &addDUT{
			shivasAddDUT: shivasAddDUT{
				servo:       "",
				servoSerial: "",
			},
		}
		if yes := ad.setupServoArguments(satlabId); yes {
			t.Errorf("Expected servo is detected but not expected!")
		}
		So(ad.qualifiedServo, ShouldEqual, "")
		So(ad.servoDockerContainerName, ShouldEqual, "")
	})
}
