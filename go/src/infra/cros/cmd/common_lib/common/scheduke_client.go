// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	schedukeapi "go.chromium.org/chromiumos/config/go/test/scheduling"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const schedukeDevPool = "schedukeTest"

var (
	schedukeDevURL               = "https://front-door-2q7tjgq5za-wl.a.run.app"
	schedukeProdURL              = "https://front-door-4vl5zcgwzq-wl.a.run.app"
	schedukeExecutionEndpoint    = "tasks/add"
	schedukeGetExecutionEndpoint = "tasks"
	maxHTTPRetries               = 5
)

type SchedukeClient struct {
	baseURL string
	client  *http.Client
	ctx     context.Context
	local   bool
}

func NewSchedukeClient(ctx context.Context, pool string, local bool) (*SchedukeClient, error) {
	baseURL := schedukeProdURL
	if pool == schedukeDevPool {
		baseURL = schedukeDevURL
	}

	client := SchedukeClient{ctx: ctx, local: local, baseURL: baseURL}
	err := client.setUpHTTPClient()
	return &client, err
}

// httpClient returns an HTTP client with authentication set up.
func (s *SchedukeClient) setUpHTTPClient() error {
	if s.local {
		s.client = &http.Client{}
		return nil
	}

	a := auth.NewAuthenticator(s.ctx, auth.SilentLogin, chromeinfra.SetDefaultAuthOptions(auth.Options{
		UseIDTokens: true,
		Audience:    s.baseURL,
	}))
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

	result := &schedukeapi.CreateTaskStatesResponse{}
	if err := proto.Unmarshal(body, result); err != nil {
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
	result := &schedukeapi.ReadTaskStatesResponse{}
	if err := proto.Unmarshal(body, result); err != nil {
		return nil, errors.Annotate(err, "unmarshal response").Err()
	}
	return result, nil
}

// ScheduleExecution will schedule TR executions via scheduke.
func (s *SchedukeClient) ScheduleExecution(req *schedukeapi.KeyedTaskRequestEvents) (*schedukeapi.CreateTaskStatesResponse, error) {
	endpoint, err := url.JoinPath(s.baseURL, schedukeExecutionEndpoint)
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
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	r, err := sendRequestWithRetries(s.client, req)
	if err != nil {
		return nil, errors.Annotate(err, "executing HTTP request").Err()
	}
	return r, nil

}

type clientThatSendsRequests interface {
	Do(*http.Request) (resp *http.Response, err error)
}

// sendRequestWithRetries sends the given request with the given HTTP client,
// retrying if any HTTP errors are returned. Retry count is controlled by
// maxHTTPRetries.
func sendRequestWithRetries(c clientThatSendsRequests, req *http.Request) (*http.Response, error) {
	var (
		retries int
		resp    *http.Response
		err     error
	)
	for retries < maxHTTPRetries {
		resp, err = c.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			return resp, err
		}
		retries += 1
	}
	return resp, err
}

// GetBBIDs will call scheduke to attempt to get BBIDs for the given tasks.
func (s *SchedukeClient) GetBBIDs(ids []int64) (*schedukeapi.ReadTaskStatesResponse, error) {
	endpoint, err := url.JoinPath(s.baseURL, schedukeGetExecutionEndpoint)
	if err != nil {
		return nil, errors.Annotate(err, "url.joinpath").Err()
	}
	withIds := fmt.Sprintf("%s?%s", endpoint, idsParam(ids))

	r, err := s.makeRequest(http.MethodGet, withIds, nil)
	if err != nil {
		return nil, errors.Annotate(err, "executing HTTP request").Err()
	}

	return s.parseGetIdsResponse(r)
}

// idsParam converts a list of BBIDs to the "ids" param for a GetBBIDs request.
func idsParam(bbIDs []int64) string {
	s := make([]string, len(bbIDs))
	for i, num := range bbIDs {
		s[i] = strconv.FormatInt(num, 10)
	}
	return fmt.Sprintf("ids=%s", strings.Join(s, ","))
}
