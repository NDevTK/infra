// Copyright 2019 The Chromium Authors
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
	"time"

	authclient "go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"

	invlibs "infra/cros/lab_inventory/protos"
	"infra/cros/lab_inventory/utils"
)

var (
	hwidServerURL              = "https://chromeos-hwid.appspot.com/api/chromeoshwid/v1/%s/%s/?key=%s"
	hwidServerResponseErrorKey = "error"
	cacheMaxAge                = 10 * time.Minute

	// New HWID oneplatform API endpoint
	hwidEndpoint      = "chromeoshwid-pa.googleapis.com"
	hwidEndpointScope = "https://www.googleapis.com/auth/chromeoshwid"
)

// Data we interested from HWID server.
type Data struct {
	// The Sku string returned by hwid server. It's not the SKU (aka variant).
	Sku string
	// The variant string returned by hwid server. It's not the variant (aka
	// SKU).
	Variant string
}

type hwidEntity struct {
	_kind   string `gae:"$kind,HwidData"`
	ID      string `gae:"$id"`
	Data    Data   `gae:",noindex"`
	Updated time.Time
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
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	// HWID server response errors as a json stream with 200 code.
	if v, ok := result[hwidServerResponseErrorKey]; ok {
		return nil, errors.Reason(v.(string)).Err()
	}
	return body, nil
}

// GetHwidData gets the hwid data from hwid server.
func GetHwidData(ctx context.Context, hwid, secret string) (*Data, error) {
	now := time.Now().UTC()
	e := hwidEntity{ID: hwid}
	errFromDatastore := datastore.Get(ctx, &e)
	if errFromDatastore == nil {
		if now.Sub(e.Updated) < cacheMaxAge {
			logging.Debugf(ctx, "HWID HIT: %#v", hwid)
			return &e.Data, nil
		}
	}
	logging.Debugf(ctx, "HWID MISS or STALE: %#v", hwid)
	d, err := getDataFromHwidServer(ctx, hwid, secret)
	if err != nil {
		if errFromDatastore == nil {
			logging.Warningf(ctx, "Use stale data as HWID server failed: %s", err.Error())
			return &e.Data, nil
		}
		return nil, err
	}
	e.Data = *d
	e.Updated = now
	if err := datastore.Put(ctx, &e); err != nil {
		logging.Warningf(ctx, "failed to cache hwid: %#v: %s", hwid, err.Error())
	}
	return d, nil
}

func getDataFromHwidServer(ctx context.Context, hwid string, secret string) (*Data, error) {
	data := Data{}
	rspBody, err := callHwidServer("dutlabel", hwid, secret)
	if err != nil {
		return nil, err
	}
	var dutlabels map[string][]interface{}
	if err := json.Unmarshal(rspBody, &dutlabels); err != nil {
		return nil, err
	}
	for key, value := range dutlabels {
		if key != "labels" {
			continue
		}
		for _, labelData := range value {
			label := labelData.(map[string]interface{})
			switch label["name"].(string) {
			case "sku":
				data.Sku = label["value"].(string)
			case "variant":
				data.Variant = label["value"].(string)
			}
		}
	}

	return &data, nil
}

type HWIDClient struct {
	hc *http.Client
}

func Init(ctx context.Context) (*HWIDClient, error) {
	tr, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(authclient.OAuthScopeEmail, hwidEndpointScope))
	if err != nil {
		return nil, err
	}
	return &HWIDClient{
		hc: &http.Client{Transport: tr},
	}, nil

}

func (c *HWIDClient) QueryHWID(ctx context.Context, hwid string) (*Data, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   hwidEndpoint,
		Path:   fmt.Sprintf("v2/dutlabel/%s", hwid),
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	dutLabel := &invlibs.GetDutLabelResponse{}
	if err := utils.ExecuteRequest(ctx, c.hc, req, dutLabel); err != nil {
		return nil, err
	}

	return parseGetDutLabelResponse(dutLabel), nil
}

func parseGetDutLabelResponse(resp *invlibs.GetDutLabelResponse) *Data {
	data := Data{}
	for _, l := range resp.GetDutLabel().GetLabels() {
		switch l.GetName() {
		case "sku":
			data.Sku = l.GetValue()
		case "variant":
			data.Variant = l.GetValue()
		}
	}
	return &data
}
