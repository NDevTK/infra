package cros

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/clock/testclock"
	"golang.org/x/net/context"
)

func TestLoadfile(t *testing.T) {
	now := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	c := context.Background()
	c, _ = testclock.UseTime(c, now)
	Convey("Using tmp directory", t, func() {
		// Use tmp dir to create a mock file
		path, err := ioutil.TempDir("", "cros-devicefile-test")
		So(err, ShouldBeNil)
		defer os.RemoveAll(path)

		fileName := filepath.Join(path, "file.json")
		Convey("loads a valid file", func() {
			err := ioutil.WriteFile(fileName, []byte(`{
			"devices": {
			  "variable_chromeos_device_hostname": {
			    "CHROMEOS_RELEASE_BOARD": "kevin",
			    "CHROMEOS_RELEASE_VERSION": "12248.0.0-rc1",
			    "battery": {
			      "battery_current": 0.17,
			      "battery_percent": 98.99,
			      "battery_present": 1,
			      "battery_status": "Charging",
			      "battery_voltage": 8.63,
			      "line_power_connected": 1,
			      "line_power_current": 0.0,
			      "line_power_type": "USB_PD"
			    },
			    "memory": {
			      "Buffers": 22936,
			      "Cached": 296228,
			      "MemAvailable": 3541856,
			      "MemFree": 3393624,
			      "MemTotal": 3902892
			    },
			    "temperature": {
			      "bigcpu-reg-thermal": [
			        24.903
			      ],
			      "gpu-thermal": [
			        25.0
			      ],
			      "litcpu-reg-thermal": [
			        24.649
			      ],
			      "sbs-9-000b": [
			        23.9
			      ],
			      "soc-thermal": [
			        25.0
			      ]
			    }
			  }
			},
			"timestamp": 1559855998.093489
			}`), 0644)
			So(err, ShouldBeNil)
			f, err := loadfile(c, fileName)
			So(err, ShouldBeNil)
			So(f, ShouldResemble, deviceStatusFile{
				Devices: map[string]deviceStatus{
					"variable_chromeos_device_hostname": {
						Battery: battery{
							Level:   98.99,
							Current: 0.17,
						},
						Mem: memory{
							Avail: 3541856,
							Total: 3902892,
							Free:  3393624,
						},
						Temp: map[string][]float64{
							"bigcpu-reg-thermal": {24.903},
							"gpu-thermal":        {25.0},
							"litcpu-reg-thermal": {24.649},
							"sbs-9-000b":         {23.9},
							"soc-thermal":        {25.0},
						},
						OSversion: "12248.0.0-rc1",
						Board:     "kevin",
					},
				},
				Timestamp: 1559855998.093489,
			})
		})

		Convey("loads a valid file, no memory", func() {
			err := ioutil.WriteFile(fileName, []byte(`
			{
			"devices": {
			  "variable_chromeos_device_hostname": {
			    "CHROMEOS_RELEASE_BOARD": "kevin",
			    "CHROMEOS_RELEASE_VERSION": "12248.0.0-rc1",
			    "battery": {
			      "battery_current": 0.17,
			      "battery_percent": 98.99,
			      "battery_present": 1,
			      "battery_status": "Charging",
			      "battery_voltage": 8.63,
			      "line_power_connected": 1,
			      "line_power_current": 0.0,
			      "line_power_type": "USB_PD"
			    },
			    "temperature": {
			      "bigcpu-reg-thermal": [
			        24.903
			      ],
			      "gpu-thermal": [
			        25.0
			      ],
			      "litcpu-reg-thermal": [
			        24.649
			      ],
			      "sbs-9-000b": [
			        23.9
			      ],
			      "soc-thermal": [
			        25.0
			      ]
			    }
			  }
			},
			"timestamp": 1559855998.093489
			}
			`), 0644)
			So(err, ShouldBeNil)
			f, err := loadfile(c, fileName)
			So(f, ShouldResemble, deviceStatusFile{
				Devices: map[string]deviceStatus{
					"variable_chromeos_device_hostname": {
						Battery: battery{
							Level:   98.99,
							Current: 0.17,
						},
						Temp: map[string][]float64{
							"bigcpu-reg-thermal": {24.903},
							"gpu-thermal":        {25.0},
							"litcpu-reg-thermal": {24.649},
							"sbs-9-000b":         {23.9},
							"soc-thermal":        {25.0},
						},
						OSversion: "12248.0.0-rc1",
						Board:     "kevin",
					},
				},
				Timestamp: 1559855998.093489,
			})
		})

		Convey("loads a valid file with missing fields ", func() {
			err := ioutil.WriteFile(fileName, []byte(`
			{
			"devices": {
			  "variable_chromeos_device_hostname": {
			    "CHROMEOS_RELEASE_BOARD": "kevin",
			    "CHROMEOS_RELEASE_VERSION": "12248.0.0-rc1",
			    "battery": {
			      "battery_current": 0.17,
			      "battery_percent": 98.99,
			      "battery_present": 1,
			      "battery_status": "Charging",
			      "battery_voltage": 8.63,
			      "line_power_connected": 1,
			      "line_power_current": 0.0,
			      "line_power_type": "USB_PD"
			    },
			    "memory": {
			      "Buffers": 22936,
			      "MemFree": 3393624,
			      "MemTotal": 3902892
			    },
			    "temperature": {
			      "bigcpu-reg-thermal": [],
			      "gpu-thermal": [],
			      "litcpu-reg-thermal": [
			        24.649
			      ],
			      "sbs-9-000b": [
			        23.9
			      ],
			      "soc-thermal": [
			        25.0
			      ]
			    }
			  }
			},
			"timestamp": 1559855998.093489
			}
			`), 0644)
			So(err, ShouldBeNil)
			f, err := loadfile(c, fileName)
			So(f, ShouldResemble, deviceStatusFile{
				Devices: map[string]deviceStatus{
					"variable_chromeos_device_hostname": {
						Battery: battery{
							Level:   98.99,
							Current: 0.17,
						},
						Mem: memory{
							Total: 3902892,
							Free:  3393624,
						},
						Temp: map[string][]float64{
							"bigcpu-reg-thermal": {},
							"gpu-thermal":        {},
							"litcpu-reg-thermal": {24.649},
							"sbs-9-000b":         {23.9},
							"soc-thermal":        {25.0},
						},
						OSversion: "12248.0.0-rc1",
						Board:     "kevin",
					},
				},
				Timestamp: 1559855998.093489,
			})
		})

		Convey("file not found", func() {
			_, err := loadfile(c, "/file/not/found")
			So(err, ShouldNotBeNil)
		})
		Convey("invalid json", func() {
			err := ioutil.WriteFile(fileName,
				[]byte(`not valid json`), 0644)
			So(err, ShouldBeNil)

			_, err = loadfile(c, fileName)
			So(err, ShouldNotBeNil)
		})
		Convey("stale json", func() {
			staleTime := float64(now.Unix()) - 161.0
			err := ioutil.WriteFile(fileName, []byte(fmt.Sprintf(`
			{
			  "timestamp": %9f
			}
			`, staleTime)), 0644)
			So(err, ShouldBeNil)

			_, err = loadfile(c, fileName)
			So(err, ShouldNotBeNil)
		})
	})
}
