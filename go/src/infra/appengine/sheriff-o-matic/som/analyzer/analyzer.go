// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analyzer

import (
	"context"
	"sync"
	"time"

	"go.chromium.org/luci/gae/service/info"

	"infra/appengine/sheriff-o-matic/som/client"
	"infra/monitoring/messages"
)

const (
	// StepCompletedRun is a synthetic step name used to indicate the build run is complete.
	StepCompletedRun = "completed run"

	prodAppID = "sheriff-o-matic"
)

// Analyzer runs the process of checking builder groups, builders, test results and so on,
// in order to produce alerts.
type Analyzer struct {
	// MaxRecentBuilds is the maximum number of recent builds to check, per builder.
	MaxRecentBuilds int

	// MinRecentBuilds is the minimum number of recent builds to check, per builder.
	MinRecentBuilds int

	// HungBuilerThresh is the maxumum length of time a builder may be in state "building"
	// before triggering a "hung builder" alert.
	HungBuilderThresh time.Duration

	// OfflineBuilderThresh is the maximum length of time a builder may be in state "offline"
	//  before triggering an "offline builder" alert.
	OfflineBuilderThresh time.Duration

	// IdleBuilderCountThresh is the maximum number of builds a builder may have in queue
	// while in the "idle" state before triggering an "idle builder" alert.
	IdleBuilderCountThresh int64

	// rslck protects revisionSummaries from concurrent access.
	rslck             *sync.Mutex
	revisionSummaries map[string]messages.RevisionSummary

	// Now is useful for mocking the system clock in testing and simulating time
	// during replay.
	Now func() time.Time

	// Mock these out in tests.
	CrBug    client.CrBug
	GoFindit client.GoFindit
}

// New returns a new Analyzer. If client is nil, it assigns a default implementation.
// maxBuilds is the maximum number of builds to check, per builder.
func New(minBuilds, maxBuilds int) *Analyzer {
	return &Analyzer{
		MaxRecentBuilds:        maxBuilds,
		MinRecentBuilds:        minBuilds,
		HungBuilderThresh:      3 * time.Hour,
		OfflineBuilderThresh:   90 * time.Minute,
		IdleBuilderCountThresh: 50,
		rslck:                  &sync.Mutex{},
		revisionSummaries:      map[string]messages.RevisionSummary{},
		Now: func() time.Time {
			return time.Now()
		},
	}
}

// GetRevisionSummaries returns a slice of RevisionSummaries for the list of hashes.
func (a *Analyzer) GetRevisionSummaries(hashes []string) ([]*messages.RevisionSummary, error) {
	ret := []*messages.RevisionSummary{}
	for _, h := range hashes {
		a.rslck.Lock()
		s, ok := a.revisionSummaries[h]
		a.rslck.Unlock()
		if !ok {
			continue
		}
		ret = append(ret, &s)
	}

	return ret, nil
}

// CreateAnalyzer creates a new analyzer and set its service clients.
func CreateAnalyzer(c context.Context) *Analyzer {
	a := New(5, 100)
	setServiceClients(c, a)
	return a
}

func setServiceClients(c context.Context, a *Analyzer) {
	if info.AppID(c) == prodAppID {
		crBug, _, goFindit := client.ProdClients(c)
		a.CrBug = crBug
		a.GoFindit = goFindit
	} else {
		crBug, _, goFindit := client.StagingClients(c)
		a.CrBug = crBug
		a.GoFindit = goFindit
	}
}
