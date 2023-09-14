// Copyright 2021 The Chromium Authors.
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

package render

import (
	"fmt"
	"regexp"

	"go.chromium.org/luci/common/errors"

	"infra/chromeperf/pinpoint/proto"
)

var (
	legacyJobRe    = regexp.MustCompile(`^jobs/legacy-(?P<legacy_id>[a-fA-F1-9][a-fA-F0-9]*)$`)
	legacyJobIDIdx = legacyJobRe.SubexpIndex("legacy_id")
)

func JobURL(j *proto.Job) (string, error) {
	return legacyJobURL(j)
}

func JobID(j *proto.Job) (string, error) {
	if len(j.Name) == 0 {
		return "", errors.Reason("invalid job, the Name field is required").Err()
	}
	m := legacyJobRe.FindStringSubmatch(j.Name)
	if m == nil {
		return "", errors.Reason("unsupported job id format: %s", j.Name).Err()
	}
	return m[legacyJobIDIdx], nil
}

func legacyJobURL(j *proto.Job) (string, error) {
	id, err := JobID(j)
	if err != nil {
		return "", err
	}
	return legacyURL(id), nil
}

func legacyURL(jobName string) string {
	return fmt.Sprintf("https://pinpoint-dot-chromeperf.appspot.com/job/%s", jobName)
}
