// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package hwid

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.chromium.org/chromiumos/infra/proto/go/device"
	"go.chromium.org/chromiumos/infra/proto/go/manufacturing"
	"go.chromium.org/luci/common/errors"
)

var (
	hwidServerURL = "https://chromeos-hwid.appspot.com/api/chromeoshwid/v1/%s/%s/?key=%s"
)

// Data we interested from HWID server.
type Data struct {
	Phase   manufacturing.Config_Phase
	Variant device.VariantId
}

func callHwidServer(rpc string, hwid string, secret string) ([]byte, error) {
	url := fmt.Sprintf(hwidServerURL, rpc, url.PathEscape(hwid), secret)
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, errors.Reason("HWID server responsonse was not OK: %s", body).Err()
	}
	return body, nil
}

func getPhase(ctx context.Context, hwid string, secret string) (string, error) {
	body, err := callHwidServer("dutlabel", hwid, secret)
	if err != nil {
		return "", err
	}
	var dutlabels map[string][]interface{}
	if err := json.Unmarshal(body, &dutlabels); err != nil {
		return "", err
	}
	for key, value := range dutlabels {
		if key != "labels" {
			continue
		}
		for _, data := range value {
			label := data.(map[string]interface{})
			if label["name"].(string) != "phase" {
				continue
			}
			return label["value"].(string), nil
		}
	}
	return "", nil
}

func getVariant(ctx context.Context, hwid string, secret string) (string, error) {
	body, err := callHwidServer("bom", hwid, secret)
	if err != nil {
		return "", err
	}
	var components map[string][]struct {
		Class string `json:"componentClass"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &components); err != nil {
		return "", err
	}
	for _, comp := range components["components"] {
		if comp.Class == "sku" { // Variant is aka sku.
			fmt.Println("foudn")
			return comp.Name, nil
		}
	}
	return "", nil
}

// GetHwidData gets the hwid data from hwid server.
func GetHwidData(ctx context.Context, hwid string, secret string) (*Data, error) {
	// TODO (guocb) cache the hwid data.
	data := Data{}
	phase, err := getPhase(ctx, hwid, secret)
	if err != nil {
		return nil, err
	}
	if phaseValue, ok := manufacturing.Config_Phase_value["PHASE_"+phase]; ok {
		data.Phase = manufacturing.Config_Phase(phaseValue)
	} else {
		return nil, errors.Reason("Unknown phase: %s", phase).Err()
	}

	variant, err := getVariant(ctx, hwid, secret)
	if err != nil {
		return nil, err
	}
	data.Variant = device.VariantId{Value: variant}
	return &data, nil
}
