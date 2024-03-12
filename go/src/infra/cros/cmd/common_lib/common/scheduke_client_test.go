// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"net/http"
	"testing"
)

var testIDsParamData = []struct {
	bbIDs        []int64
	wantIDsParam string
}{
	{
		bbIDs:        []int64{4, 9, 2, 6, 0},
		wantIDsParam: "ids=4,9,2,6,0",
	},
	{
		bbIDs:        []int64{1},
		wantIDsParam: "ids=1",
	},
	{
		bbIDs:        []int64{},
		wantIDsParam: "ids=",
	},
	{
		bbIDs:        nil,
		wantIDsParam: "ids=",
	},
}

func TestIDsParam(t *testing.T) {
	t.Parallel()
	for _, tt := range testIDsParamData {
		tt := tt
		t.Run(fmt.Sprintf("(%v)", tt.bbIDs), func(t *testing.T) {
			t.Parallel()
			gotIDsParam := idsParam(tt.bbIDs)
			if gotIDsParam != tt.wantIDsParam {
				t.Errorf("got %s, want %s", gotIDsParam, tt.wantIDsParam)
			}
		})
	}
}

// clientWithHTTPError returns numErrorsToReturn HTTP error codes before
// returning a 200 HTTP response.
type clientWithHTTPError struct {
	numErrorsToReturn int
}

func (c *clientWithHTTPError) Do(_ *http.Request) (resp *http.Response, err error) {
	if c.numErrorsToReturn > 0 {
		c.numErrorsToReturn -= 1
		return &http.Response{StatusCode: http.StatusInternalServerError}, nil
	}
	return &http.Response{StatusCode: http.StatusOK}, nil
}

var testSendRequestWithRetriesData = []struct {
	client              *clientWithHTTPError
	wantRemainingErrors int
	wantStatusCode      int
}{
	{
		&clientWithHTTPError{7},
		2,
		http.StatusInternalServerError,
	},
	{
		&clientWithHTTPError{5},
		0,
		http.StatusInternalServerError,
	},
	{
		&clientWithHTTPError{4},
		0,
		http.StatusOK,
	},
	{
		&clientWithHTTPError{0},
		0,
		http.StatusOK,
	},
}

func TestSendRequestWithRetries(t *testing.T) {
	t.Parallel()
	for _, tt := range testSendRequestWithRetriesData {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.client), func(t *testing.T) {
			t.Parallel()
			gotResp, err := sendRequestWithRetries(tt.client, nil)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			if gotResp.StatusCode != tt.wantStatusCode {
				t.Errorf("gotResp.StatusCode: got %v, wanted %v", gotResp.StatusCode, tt.wantStatusCode)
			}
			if tt.client.numErrorsToReturn != tt.wantRemainingErrors {
				t.Errorf("remaining errors: got %v, wanted %v", tt.client.numErrorsToReturn, tt.wantRemainingErrors)
			}
		})
	}
}

func TestBaseSchedukeURL(t *testing.T) {
	wantProdURL := "https://front-door-4vl5zcgwzq-wl.a.run.app"
	gotProdURL := baseSchedukeURL(false)
	if gotProdURL != wantProdURL {
		t.Errorf("got %v, want %v", gotProdURL, wantProdURL)
	}

	wantDevURL := "https://front-door-2q7tjgq5za-wl.a.run.app"
	gotDevURL := baseSchedukeURL(true)
	if gotDevURL != wantDevURL {
		t.Errorf("got %v, want %v", gotDevURL, wantDevURL)
	}
}
