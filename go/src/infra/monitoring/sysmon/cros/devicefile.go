package cros

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"go.chromium.org/luci/common/clock"
	"golang.org/x/net/context"
)

const (
	maxStaleness = time.Second * 160
	fileGlob     = "*_cros_device_status.json"
)

// deviceStatusFile is the contents of ~/*cros_device_status.json file, but
// only the fields we care about
type deviceStatusFile struct {
	Devices   map[string]deviceStatus `json:"devices"`
	Timestamp float64                 `json:"timestamp"`
}

type deviceStatus struct {
	Battery battery              `json:"battery"`
	Temp    map[string][]float64 `json:"temperature"`
	Board   string               `json:"CHROMEOS_RELEASE_BOARD"`
}

type battery struct {
	Level   float64 `json:"battery_percent"`
	Current float64 `json:"battery_current"`
}

// GetCPUTemp returns the cpu temperature of the DUT
func (d *deviceStatus) GetCPUTemp() *float64 {
	if d.Board == "kevin" {
		return &(d.Temp["soc-thermal"][0])
	}
	return nil
}

func loadfile(c context.Context, path string) (df deviceStatusFile, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &df)
	if err != nil {
		return
	}
	ts := time.Unix(0, int64(df.Timestamp*float64(time.Second)))
	now := clock.Now(c)
	staleness := now.Sub(ts)
	if staleness >= maxStaleness {
		err = fmt.Errorf(
			"Device status file is stale. Last update %v ago",
			staleness)
		return
	}
	return
}
