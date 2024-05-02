// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	schedukeapi "go.chromium.org/chromiumos/config/go/test/scheduling"
	"go.chromium.org/luci/auth"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

const (
	dmExperiment      = "dm"
	poolConfigsDirURL = "https://chrome-internal.googlesource.com/chromeos/infra/config/+/refs/heads/main/testingconfig/"
	schedukeDevPool   = "schedukeTest"
)

var (
	schedukeDevURL                  = "https://front-door-2q7tjgq5za-wl.a.run.app"
	schedukeProdURL                 = "https://front-door-4vl5zcgwzq-wl.a.run.app"
	schedukeExecutionEndpoint       = "tasks/add"
	schedukeGetExecutionEndpoint    = "tasks"
	schedukeCancelExecutionEndpoint = "tasks/cancel"
	maxHTTPRetries                  = 5
	blockedPoolsURL                 = poolConfigsDirURL + "blocked_pools.txt?format=text"
	dmPoolsURL                      = poolConfigsDirURL + "dm_pools.txt?format=text"
	schedukePoolsURL                = poolConfigsDirURL + "ctp2_pools.txt?format=text"
	gerritAuthOptsOnBot             = chromeinfra.SetDefaultAuthOptions(auth.Options{
		Method: auth.AutoSelectMethod,
		Scopes: []string{auth.OAuthScopeEmail, gitiles.OAuthScope},
	})
)

type SchedukeClient struct {
	baseURL                          string
	gerritClient, schedukeHTTPClient *http.Client
	ctx                              context.Context
	local                            bool
}

func NewLocalSchedukeClient(ctx context.Context, dev bool, gerritAuthOpts auth.Options) (*SchedukeClient, error) {
	baseURL := schedukeProdURL
	if dev {
		baseURL = schedukeDevURL
	}
	client := SchedukeClient{ctx: ctx, local: true, baseURL: baseURL}
	err := client.setUpHTTPClients(gerritAuthOpts)
	return &client, err
}

func NewSchedukeClient(ctx context.Context, pool string, local bool) (*SchedukeClient, error) {
	baseURL := schedukeProdURL
	if pool == schedukeDevPool {
		baseURL = schedukeDevURL
	}

	client := SchedukeClient{ctx: ctx, local: local, baseURL: baseURL}
	err := client.setUpHTTPClients(gerritAuthOptsOnBot)
	return &client, err
}

// httpClient configures HTTP clients for Scheduke and Gerrit, with
// authentication set up.
func (s *SchedukeClient) setUpHTTPClients(gerritAuthOpts auth.Options) error {
	// Gerrit requires auth options whether running locally or on a bot.
	ga := auth.NewAuthenticator(s.ctx, auth.SilentLogin, gerritAuthOpts)
	gc, err := ga.Client()
	if err != nil {
		return errors.Annotate(err, "create Gerrit http client").Err()
	}
	s.gerritClient = gc

	// Scheduke only requires auth options when running on a bot.
	if s.local {
		s.schedukeHTTPClient = &http.Client{}
		return nil
	}
	sa := auth.NewAuthenticator(s.ctx, auth.SilentLogin, chromeinfra.SetDefaultAuthOptions(auth.Options{
		UseIDTokens: true,
		Audience:    s.baseURL,
	}))
	sc, err := sa.Client()
	if err != nil {
		return errors.Annotate(err, "create Scheduke http client").Err()
	}
	s.schedukeHTTPClient = sc
	return nil
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

func (s *SchedukeClient) parseReadResponse(response *http.Response) (*schedukeapi.ReadTaskStatesResponse, error) {
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
	var pools []string
	for _, e := range req.GetEvents() {
		pools = append(pools, e.Pool)
	}
	poolsBlocked, err := s.AnyStringInGerritList(pools, blockedPoolsURL)
	if err != nil {
		return nil, err
	}
	if poolsBlocked {
		return nil, fmt.Errorf("leasing is currently blocked for pools %s; try again later", pools)
	}

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
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("scheduke returned %d", response.StatusCode)
	}
	return s.parseSchedukeRequestResponse(response)
}

func (s *SchedukeClient) makeRequest(method string, url string, body io.Reader) (*http.Response, error) {
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

	r, err := sendRequestWithRetries(s.schedukeHTTPClient, req)
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

// ScheduleBuildReqToSchedukeReq converts a Buildbucket ScheduleBuildRequest to
// a Scheduke request with the given event time.
func (s *SchedukeClient) ScheduleBuildReqToSchedukeReq(bbReq *buildbucketpb.ScheduleBuildRequest) (*schedukeapi.KeyedTaskRequestEvents, error) {
	bbReqBytes := []byte(protojson.Format(bbReq))
	compressedReqJSON, err := compressAndEncodeBBReq(bbReqBytes)
	if err != nil {
		return nil, fmt.Errorf("error compressing and encoding ScheduleBuildRequest %v: %w", bbReq, err)
	}
	deadlineStruct, err := getDeadlineStruct(bbReq)
	if err != nil {
		return nil, err
	}
	parentBBIDStr, err := getParentBBIDstr(bbReq)
	if err != nil {
		return nil, err
	}
	var parentBBID int64
	// Fail softly if parentBuildId field is not set on the request, as Scheduke
	// only uses this for metadata/logging.
	if parentBBIDStr != "" {
		parentBBID, err = strconv.ParseInt(parentBBIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid parent BBID found on ScheduleBuildRequest %v", bbReq)
		}
	}
	deadline, err := timeFromTimestampPBString(deadlineStruct.GetStringValue())
	if err != nil {
		return nil, fmt.Errorf("error parsing deadline for ScheduleBuildRequest %v: %w", bbReq, err)
	}
	tags := bbReq.GetTags()
	qsAccount := qsAccount(tags)
	periodic := periodic(tags)
	asap := asap(qsAccount, periodic)
	dims, deviceName, pool := dimensionsDeviceNameAndPool(bbReq.GetDimensions())

	var experiments []string
	useDM, err := s.shouldUseDM(pool)
	if useDM {
		experiments = append(experiments, dmExperiment)
	}

	schedukeTask := &schedukeapi.TaskRequestEvent{
		EventTime:                time.Now().UnixMicro(),
		Deadline:                 deadline.UnixMicro(),
		Periodic:                 periodic,
		Priority:                 priority(tags),
		RequestedDimensions:      dims,
		RealExecutionMinutes:     0, // Unneeded outside of shadow mode.
		MaxExecutionMinutes:      30,
		QsAccount:                qsAccount,
		Pool:                     pool,
		Bbid:                     parentBBID,
		Asap:                     asap,
		ScheduleBuildRequestJson: compressedReqJSON,
		DeviceName:               deviceName,
		Experiments:              experiments,
	}

	return &schedukeapi.KeyedTaskRequestEvents{
		Events: map[int64]*schedukeapi.TaskRequestEvent{
			SchedukeTaskRequestKey: schedukeTask,
		},
	}, nil
}

// LeaseRequest constructs a keyed TaskRequestEvent to request a lease from
// Scheduke with the given dimensions and lease length in minutes, for the given
// user, at the given time.
func (s *SchedukeClient) LeaseRequest(schedukeDims *schedukeapi.SwarmingDimensions, pool, deviceName, user string, mins int64, t time.Time) (*schedukeapi.KeyedTaskRequestEvents, error) {
	useDM, err := s.shouldUseDM(pool)
	if err != nil {
		return nil, err
	}
	var (
		scheduleBuildReqJSON string
		experiments          []string
	)
	if useDM {
		experiments = append(experiments, dmExperiment)
	} else {
		req, err := leaseBBReq(schedukeDims, mins)
		if err != nil {
			return nil, err
		}
		reqByes := []byte(protojson.Format(req))
		scheduleBuildReqJSON, err = compressAndEncodeBBReq(reqByes)
		if err != nil {
			return nil, err
		}
	}

	return &schedukeapi.KeyedTaskRequestEvents{
		Events: map[int64]*schedukeapi.TaskRequestEvent{
			schedukeTaskKey: {
				EventTime:                t.UnixMicro(),
				Deadline:                 t.Add(leaseSchedulingWindow).UnixMicro(),
				Periodic:                 false,
				Priority:                 leasePriority,
				RequestedDimensions:      schedukeDims,
				RealExecutionMinutes:     mins,
				MaxExecutionMinutes:      mins,
				ScheduleBuildRequestJson: scheduleBuildReqJSON,
				QsAccount:                leasesSchedulingAccount,
				Pool:                     pool,
				Bbid:                     0,
				Asap:                     false,
				TaskStateId:              0,
				DeviceName:               deviceName,
				User:                     user,
				Experiments:              experiments,
			},
		},
	}, nil
}

// shouldUseDM returns a bool indicating whether a task request with the given
// pool should enable the Device Manager experiment.
func (s *SchedukeClient) shouldUseDM(pool string) (bool, error) {
	return s.AnyStringInGerritList([]string{pool}, dmPoolsURL)
}

// AnyStringInGerritList checks for any overlap between the given list of
// strings, adn the list at the given Gerrit URL.
func (s *SchedukeClient) AnyStringInGerritList(list []string, listURL string) (bool, error) {
	fileText, err := s.fetchFileFromURL(listURL)
	if err != nil {
		return false, err
	}
	listFromURL := strings.Split(string(fileText), ",")
	mapFromURL := map[string]bool{}
	for _, str := range listFromURL {
		mapFromURL[str] = true
	}
	for _, str := range list {
		if mapFromURL[str] {
			return true, nil
		}
	}
	return false, nil
}

// fetchFileFromURL retrieves text from the given URL, using LUCI auth.
func (s *SchedukeClient) fetchFileFromURL(url string) ([]byte, error) {
	resp, err := s.gerritClient.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("error fetching file from %s: %w", url, err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("error reading file body from %s: %w", url, err)
	}
	bs, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return []byte{}, fmt.Errorf("error decoding data from %s: %w", url, err)
	}
	return bs, nil
}

// ReadTaskStates calls Scheduke to read task states for the given task state
// IDs, users, and/or device names.
func (s *SchedukeClient) ReadTaskStates(taskStateIDs []int64, users, deviceNames []string) (*schedukeapi.ReadTaskStatesResponse, error) {
	readEndpoint, err := url.JoinPath(s.baseURL, schedukeGetExecutionEndpoint)
	if err != nil {
		return nil, errors.Annotate(err, "url.joinpath").Err()
	}

	fullReadURL := fmt.Sprintf("%s?%s", readEndpoint, schedukeParams(taskStateIDs, users, deviceNames))
	r, err := s.makeRequest(http.MethodGet, fullReadURL, nil)
	if err != nil {
		return nil, errors.Annotate(err, "executing HTTP request").Err()
	}
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("scheduke returned %d", r.StatusCode)
	}
	return s.parseReadResponse(r)
}

// CancelTasks calls Scheduke to cancel tasks for the given task state IDs,
// users, and/or device names.
func (s *SchedukeClient) CancelTasks(taskStateIDs []int64, users, deviceNames []string) error {
	cancelEndpoint, err := url.JoinPath(s.baseURL, schedukeCancelExecutionEndpoint)
	if err != nil {
		return errors.Annotate(err, "url.joinpath").Err()
	}

	fullCancelURL := fmt.Sprintf("%s?%s", cancelEndpoint, schedukeParams(taskStateIDs, users, deviceNames))
	r, err := s.makeRequest(http.MethodPost, fullCancelURL, nil)
	if err != nil {
		return errors.Annotate(err, "executing HTTP request").Err()
	}
	if r.StatusCode != 200 {
		return fmt.Errorf("scheduke returned %d", r.StatusCode)
	}
	return nil
}

// schedukeParams converts a list of task state IDs, users, and device names to
// params for a request to read task states or cancel tasks.
func schedukeParams(taskStateIDs []int64, users, deviceNames []string) string {
	var params []string
	if len(taskStateIDs) > 0 {
		stringIDs := make([]string, len(taskStateIDs))
		for i, num := range taskStateIDs {
			stringIDs[i] = strconv.FormatInt(num, 10)
		}
		params = append(params, fmt.Sprintf("ids=%s", strings.Join(stringIDs, ",")))
	}
	if len(users) > 0 {
		params = append(params, fmt.Sprintf("users=%s", strings.Join(users, ",")))
	}
	if len(deviceNames) > 0 {
		params = append(params, fmt.Sprintf("device_names=%s", strings.Join(deviceNames, ",")))
	}
	return strings.Join(params, "&")
}
