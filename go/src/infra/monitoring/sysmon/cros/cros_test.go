package cros

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/common/tsmon"
	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdate(t *testing.T) {
	now := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	c := context.Background()
	c, _ = testclock.UseTime(c, now)
	Convey("In a temporary directory", t, func() {
		tmpPath, err := ioutil.TempDir("", "cros-devicefile-test")
		So(err, ShouldBeNil)
		defer os.RemoveAll(tmpPath)
		fileNames := []string{
			strings.Replace(fileGlob, "*", "device1", 1),
			strings.Replace(fileGlob, "*", "device2", 1),
			strings.Replace(fileGlob, "*", "device3", 1),
		}
		Convey("Loads a number of empty files", func() {
			for _, fileName := range fileNames {
				err := ioutil.WriteFile(filepath.Join(tmpPath,
					fileName), []byte(`
						{
						   "timestamp": 946782245,
						   "device": {}
						}
					`), 0644)
				So(err, ShouldBeNil)
			}
			err = update(c, tmpPath)
			So(err, ShouldBeNil)
		})
		Convey("Loads a number of broken files", func() {
			for _, fileName := range fileNames {
				err := ioutil.WriteFile(filepath.Join(tmpPath,
					fileName), []byte(`not json`), 0644)
				So(err, ShouldBeNil)
			}
			err = update(c, tmpPath)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMetrics(t *testing.T) {
	c := context.Background()
	c, _ = tsmon.WithDummyInMemory(c)
	devID := "build1__build2"
	Convey("Device metrics", t, func() {
		file := deviceStatusFile{
			Devices: map[string]deviceStatus{
				devID: {
					Battery: battery{
						Level:   94.63,
						Current: 0.012,
					},
					Temp: map[string][]float64{
						"soc-thermal": {25.0},
					},
					Board: "kevin",
				},
			},
			Timestamp: 946782245.0,
		}

		updateFromFile(c, file)

		So(cpuTemp.Get(c, devID), ShouldEqual, 25.0)
		So(battLevel.Get(c, devID), ShouldEqual, 94.63)
	})
}
