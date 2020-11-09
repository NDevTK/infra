// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dumper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"
	"golang.org/x/net/context/ctxhttp"
	"google.golang.org/api/googleapi"

	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	megamdmPb "infra/unifiedfleet/app/dumper/megamdm_proto"
)

var mdmHostname = "applemdm.corp.googleapis.com"
var checkinPath = "/v1/mdm/checkin?fleet=CHROME"
var serverPath = "/v1/mdm/connect?fleet=GMAC"

var testCheckInXML = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>BuildVersion</key>
	<string>UNKNOWN</string>
	<key>Challenge</key>
	<data>
	UNKNOWN
	</data>
	<key>DeviceName</key>
	<string>chops_mac_1</string>
	<key>MessageType</key>
	<string>Authenticate</string>
	<key>Model</key>
	<string>MacBookAir9,1</string>
	<key>ModelName</key>
	<string>MacBook Air</string>
	<key>OSVersion</key>
	<string>10.15.4</string>
	<key>ProductName</key>
	<string>MacBookAir9,1</string>
	<key>SerialNumber</key>
	<string>UNKNOWN</string>
	<key>Topic</key>
	<string>UNKNOWN</string>
	<key>UDID</key>
	<string>UNKNOWN</string>
</dict>
</plist>
`

func register(ctx context.Context) error {
	baseURL := &url.URL{Scheme: "https", Host: mdmHostname}
	serverURL, err := baseURL.Parse(serverPath)
	if err != nil {
		return err
	}
	fmt.Println(serverURL)

	checkinURL, err := baseURL.Parse(checkinPath)
	if err != nil {
		return err
	}
	fmt.Println(checkinURL)

	req := &megamdmPb.CheckinRequest{
		PlistData: &httpbody.HttpBody{
			Data: []byte(testCheckInXML),
		},
		Fleet: megamdmPb.Namespace_CHROME.String(),
	}
	resp := &httpbody.HttpBody{}
	t, err := auth.GetRPCTransport(ctx, auth.AsSelf)
	hc := &http.Client{Transport: t}

	// Way 1 to call mdm service
	logging.Debugf(ctx, "Trying way 1")
	var reader io.Reader
	blob, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reader = bytes.NewReader(blob)
	httpReq, err := http.NewRequest("POST", checkinURL.String(), reader)
	if err != nil {
		return err
	}
	if reader != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	logging.Debugf(ctx, "POST %s", checkinURL)
	res, err := ctxhttp.Do(ctx, hc, httpReq)
	if err != nil {
		return err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		logging.WithError(err).Errorf(ctx, "POST %s failed", checkinURL)
		return err
	}
	if err := json.NewDecoder(res.Body).Decode(resp); err != nil {
		logging.WithError(err).Errorf(ctx, "failed to parse res")
		return err
	}

	// Way 2 to call mdm service
	logging.Debugf(ctx, "Trying way 2")
	checkinClient := megamdmPb.NewCheckinServiceClient(&prpc.Client{
		C:    hc,
		Host: mdmHostname,
	})
	if _, err := checkinClient.Checkin(ctx, req); err != nil {
		return err
	}
	return nil
}
