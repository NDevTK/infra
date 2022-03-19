// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

const (
	// LastRecordedCr50FwReflashTimeKeyword is the name/kind in the karte metrics
	// used for query or update the cr 50 fw reflash information.
	LastRecordedCr50FwReflashTimeKeyword = "cr50_flash"
	// eachResourceRecoveryKeywordTemplate is the template for the name/kind for
	// query the complete recovery record for each resource.
	EachResourceRecoveryKeywordTemplate = "run_recovery_%s"
)
