// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package totmanager encapsulates all the required functions for ensuring tot
// mapping rules are followed.
package totmanager

import (
	"fmt"

	suschpb "go.chromium.org/chromiumos/infra/proto/go/testplans"

	"infra/cros/cmd/kron/common"
	"infra/cros/internal/chromeosversion"
)

var (
	tot = chromeosversion.VersionInfo{}
)

// InitTotManager fetch the chromeos version info from the overlays repo.
func InitTotManager() error {
	fileData, err := common.FetchFileFromURL(common.TotFileURL)
	if err != nil {
		return err
	}

	tot, err = chromeosversion.ParseVersionInfo(fileData)

	return err
}

// GetTot returns the calculated ToT version. If the value is 0 then that means the ToT
// info was never fetched and the nil int value is being returned.
func GetTot() int {
	return tot.ChromeBranch
}

func isCanary(milestone int) bool {
	return milestone == GetTot()
}

func isDev(milestone int) bool {
	return milestone == GetTot()-1
}

func isBeta(milestone int) bool {
	return milestone == GetTot()-2
}

func isStable(milestone int) bool {
	// TODO: remove `<` logic once LTS Suites have been introduced.
	return milestone > 0 && milestone <= GetTot()-3
}

func isLTS(milestone int) bool {
	return milestone > 0 && milestone < GetTot()-3
}

// IsTargetedBranch checks to see if the given milestone is targeted by the
// passed in branch target list.
func IsTargetedBranch(milestone int, branches []suschpb.Branch) (bool, suschpb.Branch, error) {
	if len(branches) == 0 {
		return false, suschpb.Branch_BRANCH_UNSPECIFIED, fmt.Errorf("empty branch target list passed in to TOTManager")
	}

	for _, branch := range branches {
		isTargeted := false
		targetBranch := suschpb.Branch_BRANCH_UNSPECIFIED

		switch branch {
		case suschpb.Branch_CANARY:
			isTargeted = isCanary(milestone)
			targetBranch = suschpb.Branch_CANARY
		case suschpb.Branch_DEV:
			isTargeted = isDev(milestone)
			targetBranch = suschpb.Branch_DEV
		case suschpb.Branch_BETA:
			isTargeted = isBeta(milestone)
			targetBranch = suschpb.Branch_BETA
		case suschpb.Branch_STABLE:
			isTargeted = isStable(milestone)
			targetBranch = suschpb.Branch_STABLE
		case suschpb.Branch_LTS:
			isTargeted = isLTS(milestone)
			targetBranch = suschpb.Branch_LTS
		case suschpb.Branch_BRANCH_UNSPECIFIED:
			return false, targetBranch, fmt.Errorf("branch unspecified not supported")
		default:
			return false, targetBranch, fmt.Errorf("unknown branch enum value received")
		}

		if isTargeted {
			return true, targetBranch, nil
		}
	}

	return false, suschpb.Branch_BRANCH_UNSPECIFIED, nil
}
