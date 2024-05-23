// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"math"
	"net/http"
	"time"
)

type clientThatSendsRequests interface {
	Do(*http.Request) (resp *http.Response, err error)
}

// sendHTTPRequestWithRetries sends the given request with the given HTTP
// client, retrying if any HTTP errors are returned with optional backoff. Retry
// count is controlled by maxHTTPRetries.
func sendHTTPRequestWithRetries(c clientThatSendsRequests, req *http.Request, backoff bool) (*http.Response, error) {
	var (
		retries int
		resp    *http.Response
		err     error
	)
	for retries < maxHTTPRetries {
		resp, err = c.Do(req)
		// Only retry if request was sent successfully and status was not 200.
		if err != nil || resp.StatusCode == http.StatusOK {
			break
		}
		retries += 1
		if backoff {
			time.Sleep(time.Duration(math.Pow(2, float64(retries))) * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}
