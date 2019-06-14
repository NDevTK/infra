package cros

import (
	"fmt"
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
	c, _ = tsmon.WithDummyInMemory(c)
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
						   "devices": {}
						}
					`), 0644)
				So(err, ShouldBeNil)
			}
			err = update(c, tmpPath)
			So(err, ShouldBeNil)
		})
		Convey("Loads valid files", func() {
			for idx, fileName := range fileNames {
				err := ioutil.WriteFile(filepath.Join(tmpPath,
					fileName), []byte(fmt.Sprintf(`
						{
						   "devices": {
						      "device_%v": {
						       "CHROMEOS_RELEASE_BOARD": "kevin"
						      }
						    },
						   "timestamp": 946782246
						}
						`, idx)), 0644)
				So(err, ShouldBeNil)
			}
			err := update(c, tmpPath)
			So(err, ShouldBeNil)
			for idx := range fileNames {
				deviceName := fmt.Sprintf("device_%v", idx)
				So(dutStatus.Get(c, deviceName),
					ShouldEqual, "Online")
			}
		})
		Convey("Loads some stale files", func() {
			var err error
			for idx, fileName := range fileNames {
				jsonContents := `{
						   "devices": {
						      "device_%v": {
						       "CHROMEOS_RELEASE_BOARD": "kevin"
						      }
						    },
						   "timestamp": %s
						 }`
				if idx != 2 {
					// Writing Stale Files
					err = ioutil.WriteFile(filepath.Join(tmpPath,
						fileName), []byte(
						fmt.Sprintf(jsonContents,
							idx, "946782084")), 0644)
				} else {
					// Writing one new file
					err = ioutil.WriteFile(filepath.Join(tmpPath,
						fileName), []byte(
						fmt.Sprintf(jsonContents,
							idx, "946782246")), 0644)
				}
				So(err, ShouldBeNil)
			}
			err = update(c, tmpPath)
			So(err, ShouldNotBeNil)
			for idx := range fileNames {
				deviceName := fmt.Sprintf("device_%v", idx)
				if idx == 2 {
					So(dutStatus.Get(c, deviceName),
						ShouldEqual, "Online")
				} else {
					So(dutStatus.Get(c, deviceName),
						ShouldEqual, "Offline")
				}
			}
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
