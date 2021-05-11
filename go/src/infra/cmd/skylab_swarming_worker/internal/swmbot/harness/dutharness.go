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
	"infra/cmd/skylab_swarming_worker/internal/swmbot/harness/resultsdir"
	"infra/libs/skylab/autotest/hostinfo"
	"infra/libs/skylab/inventory"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
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
	labelUpdater  labelUpdater
	// err tracks errors during setup to simplify error handling logic.
	err     error
	closers []closer
}

func makeDUTHarness(b *swmbot.Info) *DUTHarness {
	return &DUTHarness{
		BotInfo: b,
		DUTID:   b.DUTID,
		labelUpdater: labelUpdater{
			botInfo: b,
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
	s, dh.err = dutinfo.Load(ctx, dh.BotInfo, dh.labelUpdater.update)
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

// labelUpdater implements an update method that is used as a dutinfo.UpdateFunc.
type labelUpdater struct {
	botInfo      *swmbot.Info
	taskName     string
	updateLabels bool
}

// update is a dutinfo.UpdateFunc for updating DUT inventory labels.
// If adminServiceURL is empty, this method does nothing.
func (u labelUpdater) update(ctx context.Context, dutID string, old *inventory.DeviceUnderTest, new *inventory.DeviceUnderTest) error {
	// WARNING: This is an indirect check of if the job is a repair job.
	// By design, only repair job is allowed to update labels and has updateLabels set.
	// https://chromium.git.corp.google.com/infra/infra/+/7ae58795dd4badcfe9eadf4e109e27a498bed04c/go/src/infra/cmd/skylab_swarming_worker/main.go#207
	// And only repair job sets its local task account.
	// We cannot move this check later as swmbot.WithTaskAccount will fail for non-repair job.
	if u.botInfo.AdminService == "" || !u.updateLabels {
		log.Printf("Skipping label update since no admin service was provided")
		return nil
	}

	ctx, err := swmbot.WithTaskAccount(ctx)
	if err != nil {
		return errors.Annotate(err, "update inventory labels").Err()
	}

	log.Printf("Calling UFS to update dutstate")
	if err := u.updateUFS(ctx, dutID, old, new); err != nil {
		return errors.Annotate(err, "fail to update to DutState in UFS").Err()
	}
	return nil
}

func (u labelUpdater) updateUFS(ctx context.Context, dutID string, old, new *inventory.DeviceUnderTest) error {
	// Updating dutmeta, labmeta and dutstate to UFS
	ufsDutMeta := getUFSDutMetaFromSpecs(dutID, new.GetCommon())
	ufsLabMeta := getUFSLabMetaFromSpecs(dutID, new.GetCommon())
	ufsDutComponentState := getUFSDutComponentStateFromSpecs(dutID, new.GetCommon())
	ufsClient, err := swmbot.UFSClient(ctx, u.botInfo)
	if err != nil {
		return errors.Annotate(err, "fail to create ufs client").Err()
	}
	osCtx := swmbot.SetupContext(ctx, ufsUtil.OSNamespace)
	ufsResp, err := ufsClient.UpdateDutState(osCtx, &ufsAPI.UpdateDutStateRequest{
		DutState: ufsDutComponentState,
		DutMeta:  ufsDutMeta,
		LabMeta:  ufsLabMeta,
	})
	log.Printf("resp for UFS update: %#v", ufsResp)
	if err != nil {
		return errors.Annotate(err, "fail to update UFS meta & component states").Err()
	}
	return nil
}
