// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package harness

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"go.chromium.org/luci/common/errors"

	"infra/cmd/skylab_swarming_worker/internal/swmbot"
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/botinfo"
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/dutinfo"
	h_hostinfo "infra/cmd/skylab_swarming_worker/internal/swmbot/harness/hostinfo"
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/labelupdater"
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/resultsdir"
	"infra/libs/skylab/autotest/hostinfo"
	"infra/libs/skylab/inventory"
)

// Close closes and flushes out the harness resources.  This is safe
// to call multiple times.
func (dh *DUTHarness) Close(ctx context.Context) error {
	log.Printf("Wrapping up harness for %s", dh.DUTName)
	var errs []error
	for n := len(dh.closers) - 1; n >= 0; n-- {
		if err := dh.closers[n].Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Annotate(errors.MultiError(errs), "close harness").Err()
	}
	return nil
}

// DUTHarness holds information about a DUT's harness
type DUTHarness struct {
	BotInfo       *swmbot.Info
	DUTID         string
	DUTName       string
	DUTResultsDir string
	LocalState    *swmbot.LocalState
	labelUpdater  labelupdater.LabelUpdater
	// err tracks errors during setup to simplify error handling logic.
	err     error
	closers []closer
}

func makeDUTHarness(b *swmbot.Info) *DUTHarness {
	return &DUTHarness{
		BotInfo: b,
		DUTID:   b.DUTID,
		labelUpdater: labelupdater.LabelUpdater{
			BotInfo: b,
		},
	}
}

func (dh *DUTHarness) loadLocalState(ctx context.Context) {
	if dh.err != nil {
		return
	}
	if dh.DUTName == "" {
		dh.err = fmt.Errorf("DUT Name cannot be blank")
		return
	}
	bi, err := botinfo.Open(ctx, dh.BotInfo, dh.DUTName)
	if err != nil {
		dh.err = err
		return
	}
	dh.closers = append(dh.closers, bi)
	dh.LocalState = &bi.LocalState
}

func (dh *DUTHarness) loadDUTInfo(ctx context.Context) (*inventory.DeviceUnderTest, map[string]string) {
	if dh.err != nil {
		return nil, nil
	}
	var s *dutinfo.Store
	s, dh.err = dutinfo.Load(ctx, dh.BotInfo, dh.labelUpdater.Update)
	if dh.err != nil {
		return nil, nil
	}
	dh.DUTName = s.DUT.GetCommon().GetHostname()
	dh.closers = append(dh.closers, s)
	return s.DUT, s.StableVersions
}

func (dh *DUTHarness) makeHostInfo(d *inventory.DeviceUnderTest, stableVersion map[string]string) *hostinfo.HostInfo {
	if dh.err != nil {
		return nil
	}
	hip := h_hostinfo.FromDUT(d, stableVersion)
	dh.closers = append(dh.closers, hip)
	return hip.HostInfo
}

func (dh *DUTHarness) addBotInfoToHostInfo(hi *hostinfo.HostInfo) {
	if dh.err != nil {
		return
	}
	hib := h_hostinfo.BorrowBotInfo(hi, dh.LocalState)
	dh.closers = append(dh.closers, hib)
}

func (dh *DUTHarness) makeDUTResultsDir() {
	if dh.err != nil {
		return
	}
	path := filepath.Join(dh.BotInfo.ResultsDir(), dh.DUTName)
	_, err := resultsdir.Open(path)
	if err != nil {
		dh.err = err
		return
	}
	log.Printf("Created results sub-directory %s", path)
	dh.DUTResultsDir = path
}

func (dh *DUTHarness) exposeHostInfo(hi *hostinfo.HostInfo) {
	if dh.err != nil {
		return
	}
	hif, err := h_hostinfo.Expose(hi, dh.DUTResultsDir, dh.DUTName)
	if err != nil {
		dh.err = err
		return
	}
	dh.closers = append(dh.closers, hif)
}
