// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execute

import (
	"context"
	"fmt"
	"os"

	"infra/cros/internal/chromeos_version"

	vpb "go.chromium.org/chromiumos/infra/proto/go/chromiumos/version_bumper"
)

func validate(input *vpb.BumpVersionRequest) error {
	if input.ChromiumosOverlayRepo == "" {
		return fmt.Errorf("chomiumosOverlayRepo required")
	} else if _, err := os.Stat(input.ChromiumosOverlayRepo); err != nil {
		return fmt.Errorf("%s could not be found: %v", input.ChromiumosOverlayRepo, err)
	}
	if input.ComponentToBump == vpb.BumpVersionRequest_COMPONENT_TYPE_UNSPECIFIED {
		return fmt.Errorf("componentToBump was unspecified")
	}
	return nil
}

func componentToBump(input *vpb.BumpVersionRequest) (chromeos_version.VersionComponent, error) {
	switch component := input.ComponentToBump; component {
	case vpb.BumpVersionRequest_COMPONENT_TYPE_MILESTONE:
		return chromeos_version.ChromeBranch, nil
	case vpb.BumpVersionRequest_COMPONENT_TYPE_BUILD:
		return chromeos_version.Build, nil
	case vpb.BumpVersionRequest_COMPONENT_TYPE_BRANCH:
		return chromeos_version.Branch, nil
	case vpb.BumpVersionRequest_COMPONENT_TYPE_PATCH:
		return chromeos_version.Patch, nil
	default:
		return chromeos_version.Unspecified, fmt.Errorf("bad/unspecified version component")
	}
}

// Run executes the core logic for version_bumper.
func Run(ctx context.Context, input *vpb.BumpVersionRequest) error {
	if err := validate(input); err != nil {
		return err
	}

	vinfo, err := chromeos_version.GetVersionInfoFromRepo(input.ChromiumosOverlayRepo)
	if err != nil {
		return fmt.Errorf("error getting version info from repo: %v", err)
	}

	component, _ := componentToBump(input)
	vinfo.IncrementVersion(component)

	if err := vinfo.UpdateVersionFile(); err != nil {
		return fmt.Errorf("error updating version file: %v", err)
	}

	return nil
}
