// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"context"
	"fmt"
	"regexp"

	"google.golang.org/grpc"

	"go.chromium.org/luci/gae/service/info"

	monorailv3 "infra/monorailv2/api/v3/api_proto"
)

// monorailPriorityFieldMap records the resource name of priority field
// in different projects and environment.
var monorailPriorityFieldMap = map[string]map[string]string{
	"sheriff-o-matic": {
		"chromium":     "projects/chromium/fieldDefs/11",
		"fuchsia":      "projects/fuchsia/fieldDefs/168",
		"angleproject": "projects/angleproject/fieldDefs/32",
	},
	"sheriff-o-matic-staging": {
		"chromium":     "projects/chromium/fieldDefs/11",
		"fuchsia":      "projects/fuchsia/fieldDefs/246",
		"angleproject": "projects/angleproject/fieldDefs/32",
	},
	"default": {
		"chromium":     "projects/chromium/fieldDefs/11",
		"fuchsia":      "projects/fuchsia/fieldDefs/246",
		"angleproject": "projects/angleproject/fieldDefs/32",
	},
}

var monorailTypeFieldMap = map[string]map[string]string{
	"sheriff-o-matic": {
		"chromium":     "projects/chromium/fieldDefs/10",
		"angleproject": "projects/angleproject/fieldDefs/55",
	},
	"sheriff-o-matic-staging": {
		"chromium":     "projects/chromium/fieldDefs/10",
		"angleproject": "projects/angleproject/fieldDefs/55",
	},
	"default": {
		"chromium":     "projects/chromium/fieldDefs/10",
		"angleproject": "projects/angleproject/fieldDefs/55",
	},
}

// GetMonorailPriorityField get the fieldName for priority.
// TODO (nqmtuan): Put this in admin config.
func GetMonorailPriorityField(c context.Context, projectID string) (string, error) {
	return getFieldValue(c, projectID, monorailPriorityFieldMap)
}

// GetMonorailTypeField get the fieldName for type (e.g. Bug, Feature...).
// TODO (nqmtuan): Put this in admin config.
func GetMonorailTypeField(c context.Context, projectID string) (string, error) {
	return getFieldValue(c, projectID, monorailTypeFieldMap)
}

func getFieldValue(c context.Context, projectID string, fieldMap map[string]map[string]string) (string, error) {
	appID := info.AppID(c)
	if appID != "sheriff-o-matic" && appID != "sheriff-o-matic-staging" {
		appID = "default"
	}
	val, ok := fieldMap[appID][projectID]
	if !ok {
		return "", fmt.Errorf("Invalid ProjectID %q", projectID)
	}
	return val, nil
}

// GetMonorailProjectResourceName generates Monorail project resource from projectID
func GetMonorailProjectResourceName(projectID string) string {
	return "projects/" + projectID
}

// GetMonorailIssueResourceName generates Monorail issue resource from projectID
// and bugID
func GetMonorailIssueResourceName(projectID string, bugID string) string {
	return fmt.Sprintf("projects/%s/issues/%s", projectID, bugID)
}

// ParseMonorailIssueName gets projectID, bugID from issue resource name
func ParseMonorailIssueName(issueName string) (string, string, error) {
	rgx := regexp.MustCompile("projects/(.+)/issues/(\\d+)")
	rs := rgx.FindStringSubmatch(issueName)
	if len(rs) != 3 {
		return "", "", fmt.Errorf("Invalid resource %q", issueName)
	}
	return rs[1], rs[2], nil
}

// FakeMonorailIssueClient is a fake client that is used when running locally
// because the calls to Monorail always fail.  This prevents waiting for the
// 60 second timeout.
//
// Any functionality using monorail must be tested when deployed to AppEngine.
type FakeMonorailIssueClient struct{}

func (ic FakeMonorailIssueClient) SearchIssues(c context.Context, req *monorailv3.SearchIssuesRequest, ops ...grpc.CallOption) (*monorailv3.SearchIssuesResponse, error) {
	return &monorailv3.SearchIssuesResponse{}, nil
}
