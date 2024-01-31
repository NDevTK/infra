// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	schedukeapi "go.chromium.org/chromiumos/config/go/test/scheduling"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

var SCHEDUKE_PROD = "https://front-door-4vl5zcgwzq-wl.a.run.app"
var SCHEDUKE_STAGING = "https://front-door-usoglgosrq-wl.a.run.app"

var PROD = "prod"
var STAGING = "staging"

var SCHEDUKE_EXECUTION_ENDPOINT = "/tasks/add"
var SCHEDUKE_GET_EXECUTION_ENDPOINT = "/tasks"

type SchedukeClient struct {
	client  *http.Client
	ctx     context.Context
	baseURL string
	local   bool
}

func NewSchedukeClient(ctx *context.Context, env string, local bool) (*SchedukeClient, error) {
	baseURL := ""
	if env == PROD {
		baseURL = SCHEDUKE_PROD
	} else if env == STAGING {
		baseURL = SCHEDUKE_STAGING
	} else {
		return nil, fmt.Errorf("env must be oneof '%s' '%s'", PROD, STAGING)
	}

	client := SchedukeClient{ctx: *ctx, baseURL: baseURL, local: local}
	err := client.setUpHTTPClient()
	return &client, err

}

// httpClient returns an HTTP client with authentication set up.
func (s *SchedukeClient) setUpHTTPClient() error {
	if s.local {
		s.client = &http.Client{}
		return nil
	}

	o := auth.Options{
		Method: auth.LUCIContextMethod,
		Scopes: []string{
			auth.OAuthScopeEmail,
		},
	}
	a := auth.NewAuthenticator(s.ctx, auth.SilentLogin, o)
	c, err := a.Client()
	if err == nil {
		s.client = c
		return nil
	}
	return errors.Annotate(err, "create http client").Err()
}

// Not currently used; but is useful in the case we need to do a token based auth.
func token() (string, error) {
	args := []string{"auth", "print-identity-token"}
	out, err := exec.Command("gcloud", args...).Output()
	if err != nil {
		return "", errors.Annotate(err, "gcloud auth issue").Err()
	}
	o := string(out)
	fmted := strings.ReplaceAll(o, "\n", "")

	return fmted, nil
}

func (s *SchedukeClient) parseSchedukeRequestResponse(response *http.Response) (*schedukeapi.CreateTaskStatesResponse, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Annotate(err, "parsing response").Err()
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.Reason("Scheduke server responsonse was not OK: %s", body).Err()
	}
	var result *schedukeapi.CreateTaskStatesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Annotate(err, "unmarshal response").Err()
	}
	return result, nil

}

func (s *SchedukeClient) parseGetIdsResponse(response *http.Response) (*schedukeapi.ReadTaskStatesResponse, error) {
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Annotate(err, "parsing response").Err()
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.Reason("Scheduke server responsonse was not OK: %s", body).Err()
	}
	var result *schedukeapi.ReadTaskStatesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Annotate(err, "unmarshal response").Err()
	}
	return result, nil
}

// ScheduleExecution will schedule TR executions via scheduke.
func (s *SchedukeClient) ScheduleExecution(req *schedukeapi.KeyedTaskRequestEvents) (*schedukeapi.CreateTaskStatesResponse, error) {
	endpoint, err := url.JoinPath(s.baseURL, SCHEDUKE_EXECUTION_ENDPOINT)
	if err != nil {
		return nil, errors.Annotate(err, "url.joinpath").Err()
	}

	data, err := protojson.Marshal(req)
	if err != nil {
		return nil, errors.Annotate(err, "marshal request").Err()
	}
	response, err := s.makeRequest(http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, errors.Annotate(err, "HttpPost").Err()
	}
	return s.parseSchedukeRequestResponse(response)
}

func (s *SchedukeClient) makeRequest(method string, url string, body io.Reader) (*http.Response, error) {
	fmt.Println(body)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Annotate(err, "creating new HTTP request").Err()
	}

	if s.local {
		t, err := token()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t))
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	r, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, "executing HTTP request").Err()
	}
	return r, nil

}

// GetBBIDs will call scheduke to attempt to get BBIDs for the given tasks.
func (s *SchedukeClient) GetBBIDs(ids []string) (*schedukeapi.ReadTaskStatesResponse, error) {
	base := fmt.Sprintf("%s=%s", "ids", strings.Join(ids, ","))

	endpoint, err := url.JoinPath(s.baseURL, SCHEDUKE_GET_EXECUTION_ENDPOINT)
	if err != nil {
		return nil, errors.Annotate(err, "url.joinpath").Err()
	}
	withIds := fmt.Sprintf("%s?%s", endpoint, base)

	r, err := s.makeRequest(http.MethodGet, withIds, nil)
	if err != nil {
		return nil, errors.Annotate(err, "executing HTTP request").Err()
	}

	return s.parseGetIdsResponse(r)

}
