// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"fmt"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/luciexe/build"
)

// AddLinksToStepSummaryMarkdown adds provided links to provided step summary.
func AddLinksToStepSummaryMarkdown(
	step *build.Step,
	testHausUrl string,
	stainlessUrl string,
	gcsLink string) {

	links := []string{}
	if testHausUrl != "" {
		links = append(links, fmt.Sprintf("* [Testhaus Link](%s)", testHausUrl))
	}
	if stainlessUrl != "" {
		links = append(links, fmt.Sprintf("* [Stainless Link](%s)", stainlessUrl))
	}
	if gcsLink != "" {
		links = append(links, fmt.Sprintf("* [Test Artifacts Gcs Link](%s)", gcsLink))
	}

	if len(links) > 0 {
		step.SetSummaryMarkdown(strings.Join(links, "\n"))
	}
}

// CreateStepWithStatus creates a new step and sets step status based on
// provided flags. If failBuild is true, the returned error will have build
// failure status attached to it for caller to bubble up appropriately.
func CreateStepWithStatus(
	ctx context.Context,
	stepName string,
	summary string,
	isFailure bool,
	failBuild bool) (err error) {

	if stepName == "" {
		return nil
	}

	var stepErr error
	step, ctx := build.StartStep(ctx, stepName)
	defer func() {
		step.End(build.AttachStatus(stepErr, bbpb.Status_FAILURE, nil))
	}()

	if isFailure {
		step.SetSummaryMarkdown(summary)
		stepErr = fmt.Errorf("%s: %s", stepName, summary)
	}

	if isFailure && failBuild {
		err = stepErr
	}

	return err
}
