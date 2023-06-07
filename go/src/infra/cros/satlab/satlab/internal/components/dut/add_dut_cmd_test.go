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

// TestValidateHostname tests hostname validation.
func TestValidateHostname(t *testing.T) {
	tests := []struct {
		testname string
		hostname string
		wantErr  bool
	}{
		{
			testname: "valid",
			hostname: "eli-123",
			wantErr:  false,
		},
		{
			testname: "uppercase",
			hostname: "ELI-123",
			wantErr:  true,
		},
		{
			testname: "nonalphanumeric",
			hostname: "eli-123!",
			wantErr:  true,
		},
		{
			testname: "too long",
			hostname: "eli-123-eli-123-eli-123-eli-123-eli-123-eli-123-",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.testname, func(t *testing.T) {
			t.Parallel()

			if err := validateHostname(tt.hostname); (err != nil) != tt.wantErr {
				t.Errorf("validateHostname() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
