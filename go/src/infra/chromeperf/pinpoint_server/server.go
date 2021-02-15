// Copyright 2020 The Chromium Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"infra/chromeperf/pinpoint"
	"infra/chromeperf/pinpoint_server/conversion"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type pinpointServer struct {
	pinpoint.UnimplementedPinpointServer
	// Provide an HTTP Client to be used by the server, to the Pinpoint Legacy API.
	LegacyClient *http.Client
}

const endpointsHeader = "x-endpoint-api-userinfo"

func getRequestingUserEmail(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "missing metadata from request context")
	}
	auth, ok := md[endpointsHeader]
	if !ok || len(auth) == 0 {
		return "", status.Errorf(codes.PermissionDenied, "missing required auth header '%s'", endpointsHeader)
	}
	// Decode the auto header from base64encoded json, into a map we can inspect.
	decoded, err := base64.RawURLEncoding.DecodeString(auth[0])
	if err != nil {
		grpclog.Errorf("Failed decoding auth = '%v'; error = %s", auth, err)
		return "", status.Errorf(codes.InvalidArgument, "malformed %s: %v", endpointsHeader, err)
	}
	userInfo := make(map[string]interface{})
	if json.Unmarshal(decoded, &userInfo) != nil {
		return "", status.Errorf(codes.InvalidArgument, "malformed %s: %v", endpointsHeader, err)
	}
	email, ok := userInfo["email"].(string)
	if !ok || len(email) == 0 {
		return "", status.Errorf(codes.PermissionDenied, "missing 'email' field from token")
	}
	return email, nil
}

func (s *pinpointServer) ScheduleJob(ctx context.Context, r *pinpoint.ScheduleJobRequest) (*pinpoint.Job, error) {
	// First, ensure we can set the user from the incoming request, based on their identity provided in the OAuth2
	// headers, that make it into the context of this request. Because we intend this service to be hosted behind an
	// Endpoint Service Proxy (ESP), we're going to look for the authentication details in the
	// X-Endpoint-API-Userinfo header, as part of the context. We'll fail if we aren't being served behind an ESP.
	//
	// See
	// https://cloud.google.com/endpoints/docs/grpc/authenticating-users#receiving_authentication_results_in_your_api
	// for details on the format and specifications for the contents of this header.
	if s.LegacyClient == nil {
		return nil, status.Error(codes.Internal, "misconfigured service, please try again later")
	}
	userEmail, err := getRequestingUserEmail(ctx)
	if err != nil {
		return nil, err
	}

	// Before we make this service the source of truth for the Pinpoint service, we first proxy requests to the
	// actual Pinpoint legacy API from the provided request.
	values, err := conversion.ConvertToValues(r.Job, userEmail)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%v", err))
	}

	// Then we make the request to the Pinpoint service.
	res, err := s.LegacyClient.PostForm(*legacyPinpointService+"/api/new", values)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed request to legacy API: %v", err)
	}
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, status.Errorf(codes.Internal, "Internal error, service not authorized")
	case http.StatusBadRequest:
		return nil, status.Errorf(codes.Internal, "Internal error, service sent an invalid request")
	case http.StatusOK:
		break
	default:
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	// The response of the legacy service has the following format:
	//
	// {
	//    'jobId': <string>,
	//    'jobUrl': <string>
	// }
	//
	// We ignore the 'jobUrl' field for now.
	var newResponse struct {
		JobID string
	}
	if err := json.NewDecoder(res.Body).Decode(&newResponse); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse response from legacy API: %v", err)
	}

	// Return with a minimal Job response.
	// TODO(dberris): Write this data out to Spanner when we're ready to replace the legacy API.
	return &pinpoint.Job{
		Name:      fmt.Sprintf("jobs/legacy-%s", newResponse.JobID),
		CreatedBy: userEmail,
		JobSpec:   r.Job,
	}, nil
}

var jobNameRe = regexp.MustCompile(`^jobs/legacy-(?P<id>[a-f0-9]+)$`)

// MicroTime is an alias to time.Time which allows us to parse microsecond-precision time.
type MicroTime time.Time

// UnmarshalJSON supports parsing nanosecond timestamps.
func (t *MicroTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `\"`)
	p, err := time.Parse("2006-01-02T15:04:05.999999", s)
	if err != nil {
		return err
	}
	*t = MicroTime(p)
	return nil
}

func (s *pinpointServer) GetJob(ctx context.Context, r *pinpoint.GetJobRequest) (*pinpoint.Job, error) {
	// This API does not require that the user be signed in, so we'll not need to check the credentials.
	// TODO(dberris): In the future, support ACL-limiting Pinpoint job results.
	if s.LegacyClient == nil {
		return nil, status.Error(
			codes.Internal,
			"misconfigured service, please try again later")
	}

	// Make a request to the legacy API.
	// Ensure that r.Id is a hex number.
	if !jobNameRe.MatchString(r.Name) {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid id format, must match %s", jobNameRe.String())
	}
	matches := jobNameRe.FindStringSubmatch(r.Name)
	legacyID := string(matches[jobNameRe.SubexpIndex("id")])
	if len(legacyID) == 0 {
		return nil, status.Error(codes.Unimplemented, "future ids not supported yet")
	}

	u := fmt.Sprintf("%s/api/job/%s?o=STATE", *legacyPinpointService, legacyID)
	if _, err := url.Parse(u); err != nil {
		grpclog.Errorf("Invalid URL: %s", err)
		return nil, status.Errorf(codes.Internal, "failed to form a valid legacy request")
	}
	grpclog.Infof("GET %s", u)
	res, err := s.LegacyClient.Get(u)
	if err != nil {
		grpclog.Errorf("HTTP Request Error: %s", err)
		return nil, status.Errorf(codes.Internal, "failed retrieving job data from legacy service")
	}
	switch res.StatusCode {
	case http.StatusNotFound:
		return nil, status.Errorf(codes.NotFound, "job not found")
	case http.StatusOK:
		break
	default:
		return nil, status.Errorf(codes.Internal, "failed request: %s", res.Status)
	}

	var l struct {
		Arguments           map[string]string       `json:"arguments"`
		BugID               int64                   `json:"bug_id"`
		ComparisonMode      string                  `json:"comparison_mode,omitempty"`
		ComparisonMagnitude float64                 `json:"comparison_magnitude,omitempty"`
		Cfg                 string                  `json:"configuration,omitempty"`
		Created             MicroTime               `json:"created,omitempty"`
		Exception           *map[string]interface{} `json:"exception,omitempty"`
		JobID               string                  `json:"job_id,omitempty"`
		Metric              string                  `json:"metric,omitempty"`
		Name                string                  `json:"name,omitempty"`
		Project             *string                 `json:"project,omitempty"`
		StepLabels          []string                `json:"quests,omitempty"`
		ResultsURL          string                  `json:"results_url,omitempty"`
		StartedTime         MicroTime               `json:"started_time,omitempty"`
		State               []struct {
			Attempts []struct {
				Executions []struct {
					Completed bool `json:"completed"`
					Details   []struct {
						Value string `json:"value,omitempty"`
						Key   string `json:"key,omitempty"`
						URL   string `json:"url,omitempty"`
					} `json:"details"`
				} `json:"executions"`
			} `json:"attempts"`
			Change struct {
				Commits []struct {
					Author         string    `json:"author,omitempty"`
					ChangeID       string    `json:"change_id,omitempty"`
					CommitPosition int64     `json:"commit_position,omitempty"`
					Created        MicroTime `json:"created,omitempty"`
					GitHash        string    `json:"git_hash,omitempty"`
					Message        string    `json:"message,omitempty"`
					Repo           string    `json:"repository,omitempty"`
					ReviewURL      string    `json:"review_url,omitempty"`
					Subject        string    `json:"subject,omitempty"`
					URL            string    `json:"url,omitempty"`
				} `json:"commits"`
			} `json:"change"`
			Comparisons struct {
				Prev string `json:"prev,omitempty"`
				Next string `json:"next,omitempty"`
			} `json:"comparisons"`
		} `json:"state,omitempty"`
		Status  string    `json:"status,omitempty"`
		Updated MicroTime `json:"updated,omitempty"`
		User    string    `json:"user,omitempty"`
	}
	if err := json.NewDecoder(res.Body).Decode(&l); err != nil {
		grpclog.Errorf("failed parsing json: %q", err)
		return nil, status.Errorf(
			codes.Internal,
			"received ill-formed response from legacy service",
		)
	}

	transformState := func() pinpoint.Job_State {
		switch l.Status {
		case "Running":
			return pinpoint.Job_RUNNING
		case "Queued":
			return pinpoint.Job_PENDING
		case "Cancelled":
			return pinpoint.Job_CANCELLED
		case "Failed":
			return pinpoint.Job_FAILED
		case "Completed":
			return pinpoint.Job_SUCCEEDED
		}
		return pinpoint.Job_STATE_UNSPECIFIED
	}

	transformMode := func() pinpoint.JobSpec_ComparisonMode {
		switch l.ComparisonMode {
		case "functional":
			return pinpoint.JobSpec_FUNCTIONAL
		case "try", "performance":
			return pinpoint.JobSpec_PERFORMANCE
		}
		return pinpoint.JobSpec_COMPARISON_MODE_UNSPECIFIED
	}

	// Now attempt to translate the parsed JSON structure into a protobuf.
	// FIXME(dberris): Interpret the results better, differentiating experiments from bisections, etc.
	return &pinpoint.Job{
		Name:           fmt.Sprintf("jobs/legacy-%s", l.JobID),
		State:          transformState(),
		CreatedBy:      l.User,
		CreateTime:     timestamppb.New(time.Time(l.Created)),
		LastUpdateTime: timestamppb.New(time.Time(l.Updated)),
		JobSpec: &pinpoint.JobSpec{
			ComparisonMode:      transformMode(),
			ComparisonMagnitude: l.ComparisonMagnitude,
			Config:              l.Cfg,
			Target:              l.Arguments["target"],
			MonorailIssue: func() *pinpoint.MonorailIssue {
				if l.Project == nil || l.BugID == 0 {
					return nil
				}
				return &pinpoint.MonorailIssue{
					Project: *l.Project,
					IssueId: l.BugID,
				}
			}(),
		},
	}, nil
}

func (s *pinpointServer) ListJobs(ctx context.Context, r *pinpoint.ListJobsRequest) (*pinpoint.ListJobsResponse, error) {
	// TODO(dberris): Implement this!
	return nil, nil
}

func (s *pinpointServer) CancelJob(ctx context.Context, r *pinpoint.CancelJobRequest) (*pinpoint.Job, error) {
	// TODO(dberris): Implement this!
	return nil, nil
}

// Email address for the service account to use.
var serviceAccountEmail = flag.String("service_account", "", "service account email")

// Contents of the service account credentials PEM file.
var privateKey = flag.String("private_key", "", "service account PEM file contents")

// Flag to configure the legacy Pinpoint service URL base.
var legacyPinpointService = flag.String("legacy_pinpoint_service", "https://pinpoint-dot-chromeperf.appspot.com", "base URL for the legacy Pinpoint service")
