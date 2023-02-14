// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// originalFormatTagKey is a key of the tag indicating the format of the
	// source data. Possible values: FormatJTR, FormatGTest.
	originalFormatTagKey = "orig_format"

	// formatGTest is Chromium's GTest format.
	formatGTest = "chromium_gtest"

	// formatJTR is Chromium's JSON Test Results format.
	formatJTR = "chromium_json_test_results"

	// Gitiles URL for chromium/src repo.
	chromiumSrcRepo = "https://chromium.googlesource.com/chromium/src"

	// Gitiles URL for webrtc/src repo.
	webrtcSrcRepo = "https://webrtc.googlesource.com/src/"

	// ResultSink limits the summary html message to 4096 bytes in UTF-8.
	maxSummaryHtmlBytes = 4096
)

// summaryTmpl is used to generate SummaryHTML in GTest and JTR-based test
// results.
var summaryTmpl = template.Must(template.New("summary").Parse(`
{{ define "gtest" -}}
{{- template "links" .links -}}
{{- template "text_artifacts" .text_artifacts -}}
{{- end}}

{{ define "jtr" -}}
{{- template "links" .links -}}
{{- end}}

{{ define "links" -}}
{{- if . -}}
<ul>
{{- range $name, $url := . -}}
  <li><a href="{{ $url }}">{{ $name }}</a></li>
{{- end -}}
</ul>
{{- end -}}
{{- end -}}

{{ define "text_artifacts" -}}
{{- range $aid := . -}}
  <p><text-artifact artifact-id="{{ $aid }}" /></p>
{{- end -}}
{{- end -}}
`))

// msToDuration converts a time in milliseconds to duration.Duration.
func msToDuration(t float64) *duration.Duration {
	return ptypes.DurationProto(time.Duration(t) * time.Millisecond)
}

// ensureLeadingDoubleSlash ensures that the path starts with "//".
func ensureLeadingDoubleSlash(path string) string {
	return "//" + strings.TrimLeft(path, "/")
}

// normalizePath converts the artifact path to the canonical form.
func normalizePath(p string) string {
	return path.Clean(strings.ReplaceAll(p, "\\", "/"))
}

// processArtifacts walks the files in artifactDir then returns a map from normalized relative path to full path.
func processArtifacts(artifactDir string) (normPathToFullPath map[string]string, err error) {
	normPathToFullPath = map[string]string{}
	err = filepath.Walk(artifactDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			// normPath is the normalized relative path to artifactDir.
			relPath, err := filepath.Rel(artifactDir, path)
			if err != nil {
				return err
			}
			normPath := normalizePath(relPath)
			normPathToFullPath[normPath] = path
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return normPathToFullPath, err
}

// AppendTags appends a new tag to the tag slice if both key and value exist.
func AppendTags(tags []*pb.StringPair, key string, value string) []*pb.StringPair {
	if key == "" || value == "" {
		return tags
	}

	return append(tags, pbutil.StringPair(key, value))
}

// SortTags sorts the tags slice lexicographically by key, then value.
func SortTags(tags []*pb.StringPair) []*pb.StringPair {
	if len(tags) == 0 {
		return tags
	}

	pbutil.SortStringPairs(tags)
	return tags
}

// ReadJSONFileToString reads the JSON file content into a string. Return an
// empty string if the file read fails.
func ReadJSONFileToString(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}

	return string(data)
}

// parseMetadata reads the CFT test metadata file and parses into a map keyed by the test name.
func parseMetadata(filePath string) (map[string]*api.TestCaseMetadata, error) {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Annotate(err, "read test metadata file").Err()
	}

	metadata := api.TestCaseMetadataList{}
	if err := protojson.Unmarshal(f, &metadata); err != nil {
		return nil, errors.Annotate(err, "parsing test metadata file contents").Err()
	}

	mp := make(map[string]*api.TestCaseMetadata, 0)
	for _, v := range metadata.Values {
		if v.TestCase != nil {
			mp[v.TestCase.Name] = v
		}
	}
	return mp, nil
}

// metadataToTags converts the following TestCaseMetadata to a list of key value
// string pairs. Repeated fields are joined with a "," and boolean fields are
// converted to "true" or "false" strings:
//   - owners (repeated), e.g. ["chromeos-platform-power@google.com"]
//   - requirements (repeated), e.g. ["boot-perf-0001-v01"]
//   - bug_component, e.g. "b:167191"
//   - criteria, e.g. "This test is a benchmark"
//   - hw_agnostic (boolean), e.g. true, false
func metadataToTags(metadata *api.TestCaseMetadata) []*pb.StringPair {
	if metadata == nil {
		return []*pb.StringPair{}
	}

	tags := make([]*pb.StringPair, 0)
	info := metadata.TestCaseInfo
	if info == nil {
		return []*pb.StringPair{}
	}

	if info.Owners != nil {
		owners := make([]string, 0)
		for _, o := range info.Owners {
			owners = append(owners, o.Email)
		}
		tags = AppendTags(tags, "owners", strings.Join(owners, ","))
	}

	if info.Requirements != nil {
		requirements := make([]string, 0)
		for _, r := range info.Requirements {
			requirements = append(requirements, r.Value)
		}
		tags = AppendTags(tags, "requirements", strings.Join(requirements, ","))
	}

	if info.BugComponent != nil {
		tags = AppendTags(tags, "bug_component", info.BugComponent.Value)
	}

	if info.Criteria != nil {
		tags = AppendTags(tags, "criteria", info.Criteria.Value)
	}

	if info.HwAgnostic != nil {
		tags = AppendTags(tags, "hw_agnostic", strconv.FormatBool(info.HwAgnostic.Value))
	}

	return tags
}
