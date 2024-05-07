// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

const (
	DMDevURL     = "https://device-lease-service-dev-thnumbwdvq-uc.a.run.app"
	DMLeasesPort = 443
)

type DeviceManagerClient struct {
	baseURL string
	client  *http.Client
	ctx     context.Context
}

func NewDeviceManagerClient(ctx context.Context) (*DeviceManagerClient, error) {
	authOpts := chromeinfra.SetDefaultAuthOptions(auth.Options{
		UseIDTokens: true,
		Audience:    DMDevURL,
	})
	a := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)
	c, err := a.Client()
	if err != nil {
		return nil, err
	}
	return &DeviceManagerClient{
		// TODO: add prod URL once DM is productionized.
		baseURL: fmt.Sprintf("%s:%d", DMDevURL, DMLeasesPort),
		client:  c,
		ctx:     ctx,
	}, nil
}

// Extend extends the lease with the given ID by the given duration, and returns
// the new deadline.
func (d *DeviceManagerClient) Extend(leaseID string, dur time.Duration) (time.Time, error) {
	extendURL, err := url.JoinPath(d.baseURL, "ExtendLease")
	if err != nil {
		return time.Time{}, errors.Annotate(err, "joining URL paths").Err()
	}
	req := &api.ExtendLeaseRequest{
		LeaseId:        leaseID,
		ExtendDuration: durationpb.New(dur),
		IdempotencyKey: "1",
	}
	data, err := protojson.Marshal(req)
	if err != nil {
		return time.Time{}, errors.Annotate(err, "marshalling request").Err()
	}
	resp, err := d.makeHTTPRequest(http.MethodPost, extendURL, bytes.NewReader(data))
	if err != nil {
		return time.Time{}, errors.Annotate(err, "executing HTTP request").Err()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, errors.Annotate(err, "parsing response").Err()
	}
	if resp.StatusCode != 200 {
		return time.Time{}, fmt.Errorf("device manager returned %d: %s", resp.StatusCode, body)
	}
	res := &api.ExtendLeaseResponse{}
	if err := protojson.Unmarshal(body, res); err != nil {
		return time.Time{}, errors.Annotate(err, "unmarshalling response").Err()
	}
	return res.ExpirationTime.AsTime(), nil
}

// Release releases the lease with the given ID.
func (d *DeviceManagerClient) Release(leaseID string) error {
	releaseURL, err := url.JoinPath(d.baseURL, "ReleaseDevice")
	if err != nil {
		return errors.Annotate(err, "joining URL paths").Err()
	}
	req := &api.ReleaseDeviceRequest{LeaseId: leaseID}
	data, err := protojson.Marshal(req)
	if err != nil {
		return errors.Annotate(err, "marshalling request").Err()
	}
	resp, err := d.makeHTTPRequest(http.MethodPost, releaseURL, bytes.NewReader(data))
	if err != nil {
		return errors.Annotate(err, "executing HTTP request").Err()
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("device manager returned %d", resp.StatusCode)
	}
	return nil
}

// makeHTTPRequest makes an HTTP request with retries.
func (d *DeviceManagerClient) makeHTTPRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Annotate(err, "creating new HTTP request").Err()
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	r, err := sendHTTPRequestWithRetries(d.client, req)
	if err != nil {
		return nil, errors.Annotate(err, "executing HTTP request").Err()
	}
	return r, nil

}
