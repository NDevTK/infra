// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
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
