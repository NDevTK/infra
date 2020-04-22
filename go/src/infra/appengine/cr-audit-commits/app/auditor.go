// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"time"

	"context"

	"github.com/golang/protobuf/ptypes"
	ds "go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/proto/git"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/server/router"
)

// Tests can put mock clients here, prod code will ignore this global.
var auditorTestClients *Clients

// Auditor is the main entry point for scanning the commits on a given ref and
// auditing those that are relevant to the configuration.
//
// It scans the ref in the repo and creates entries for relevant commits,
// executes the audit functions on such commits, and calls notification
// functions when appropriate.
//
// This is expected to run every 5 minutes and for that reason, it has designed
// to stop 4 minutes 30 seconds and save any partial progress.
func Auditor(rc *router.Context) {
	outerCtx, resp := rc.Context, rc.Writer

	// Create a derived context with a 4:30 timeout s.t. we have enough
	// time to save results for at least some of the audited commits,
	// considering that a run of this handler will be scheduled every 5
	// minutes.
	ctx, cancelInnerCtx := context.WithTimeout(outerCtx, time.Second*time.Duration(4*60+30))
	defer cancelInnerCtx()

	cfg, refState, err := loadConfigFromContext(rc)
	if err != nil {
		http.Error(resp, err.Error(), 400)
		return
	}

	var cs *Clients
	if auditorTestClients != nil {
		cs = auditorTestClients
	} else {
		httpClient, err := getAuthenticatedHTTPClient(ctx, gerritScope, emailScope)
		if err != nil {
			http.Error(resp, err.Error(), 500)
			return
		}

		cs, err = initializeClients(ctx, cfg, httpClient)
		if err != nil {
			http.Error(resp, err.Error(), 500)
			return
		}
	}

	fl, err := getCommitLog(ctx, cfg, refState, cs)
	if err != nil {
		http.Error(resp, err.Error(), 502)
		return
	}

	// Iterate over the log, creating relevantCommit entries as appropriate
	// and updating refState. If the context expires during this process,
	// save the refState and bail.
	err = scanCommits(ctx, fl, cfg, refState)
	if err != nil && err != context.DeadlineExceeded {
		logging.WithError(err).Errorf(ctx, "Could not save new relevant commit")
		http.Error(resp, err.Error(), 503)
		return
	}
	// Save progress with an unexpired context.
	if putErr := ds.Put(outerCtx, refState); putErr != nil {
		logging.WithError(putErr).Errorf(outerCtx, "Could not save last known/interesting commits")
		http.Error(resp, putErr.Error(), 503)
		return
	}
	if err == context.DeadlineExceeded {
		// If the context has expired do not proceed with auditing.
		return
	}

	// TODO(crbug.com/976447): Split the auditing onto its own cron schedule.
	//
	// Send the relevant commits to workers to be audited, note that this
	// doesn't persist the changes, because we want to persist them together
	// in a single transaction for efficiency.
	//
	// If the context expires while performing the audits, save the commits
	// that were audited and bail.
	auditedCommits, err := performScheduledAudits(ctx, cfg, refState, cs)
	if err != nil && err != context.DeadlineExceeded {
		http.Error(resp, err.Error(), 500)
		return
	}
	if putErr := saveAuditedCommits(outerCtx, auditedCommits, cfg, refState); putErr != nil {
		http.Error(resp, err.Error(), 503)
		return
	}
	if err == context.DeadlineExceeded {
		// If the context has expired do not proceed with notifications.
		return
	}

	err = notifyAboutViolations(ctx, cfg, refState, cs)
	if err != nil {
		http.Error(resp, err.Error(), 502)
		return
	}

}

// getCommitLog gets from gitiles a list from the last known commit to the tip
// of the ref in chronological (as per git parentage) order.
func getCommitLog(ctx context.Context, cfg *RefConfig, refState *RefState, cs *Clients) ([]*git.Commit, error) {

	host, project, err := gitiles.ParseRepoURL(cfg.BaseRepoURL)
	if err != nil {
		return []*git.Commit{}, err
	}
	logReq := gitilespb.LogRequest{
		Project:            project,
		ExcludeAncestorsOf: refState.LastKnownCommit,
		Committish:         cfg.BranchName,
	}

	gc, err := cs.NewGitilesClient(host)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Could not create new gitiles client")
		return []*git.Commit{}, err
	}
	fl, err := gitiles.PagingLog(ctx, gc, logReq, 6000)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Could not get gitiles log from revision %s", refState.LastKnownCommit)
		return []*git.Commit{}, err
	}
	// Reverse the log to get revisions after `rev` time-ascending order.
	for i, j := 0, len(fl)-1; i < j; i, j = i+1, j-1 {
		fl[i], fl[j] = fl[j], fl[i]
	}

	// Make sure the log reaches the last known commit.
	if len(fl) > 0 && refState.LastKnownCommit != "" && len(fl[0].Parents) > 0 && fl[0].Parents[0] != refState.LastKnownCommit {
		panic("There is a gap between the last known commit and the beginning of the forward log")
	}
	return fl, nil
}

// scanCommits iterates over the list of commits in the given log, decides if
// each is relevant to any of the configured rulesets and creates records for
// each that is. Also updates the record for the ref, but does not persist it,
// this is instead done by Auditor after this function is executed. This is left
// to the handler in case the context given to this function expires before
// reaching the end of the log.
func scanCommits(ctx context.Context, fl []*git.Commit, cfg *RefConfig, refState *RefState) error {
	for _, commit := range fl {
		relevant := false
		for _, ruleSet := range cfg.Rules {
			if ruleSet.MatchesCommit(commit) {
				n, err := saveNewRelevantCommit(ctx, refState, commit)
				if err != nil {
					return err
				}
				refState.LastRelevantCommit = n.CommitHash
				refState.LastRelevantCommitTime = n.CommitTime
				// If the commit matches one ruleSet that's
				// enough. Break to move on to the next commit.
				relevant = true
				break
			}
		}
		ScannedCommits.Add(ctx, 1, relevant, refState.ConfigName)
		refState.LastKnownCommit = commit.Id
		// Ignore possible error, this time is used for display purposes only.
		if commit.Committer != nil {
			ct, _ := ptypes.Timestamp(commit.Committer.Time)
			refState.LastKnownCommitTime = ct
		}

	}
	return nil
}

func saveNewRelevantCommit(ctx context.Context, state *RefState, commit *git.Commit) (*RelevantCommit, error) {
	rk := ds.KeyForObj(ctx, state)

	commitTime, err := ptypes.Timestamp(commit.GetCommitter().GetTime())
	if err != nil {
		logging.WithError(err).Errorf(ctx, "Invalid commit time: %s", commitTime)
		return nil, err
	}
	rc := &RelevantCommit{
		RefStateKey:            rk,
		CommitHash:             commit.Id,
		PreviousRelevantCommit: state.LastRelevantCommit,
		Status:                 auditScheduled,
		CommitTime:             commitTime,
		CommitterAccount:       commit.Committer.Email,
		AuthorAccount:          commit.Author.Email,
		CommitMessage:          commit.Message,
	}

	if err := ds.Put(ctx, rc, state); err != nil {
		logging.WithError(err).Errorf(ctx, "Could not save %s", rc.CommitHash)
		return nil, err
	}
	logging.Infof(ctx, "saved %s", rc)

	return rc, nil
}

// saveAuditedCommits transactionally saves the records for the commits that
// were audited.
func saveAuditedCommits(ctx context.Context, auditedCommits map[string]*RelevantCommit, cfg *RefConfig, refState *RefState) error {
	// We will read the relevant commits into this slice before modifying
	// them, to ensure that we don't overwrite changes that may have been
	// saved to the datastore between the time the query in performScheduled
	// audits ran and the beginning of the transaction below; as may have
	// happened if two runs of the Audit handler ran in parallel.
	originalCommits := []*RelevantCommit{}
	for _, auditedCommit := range auditedCommits {
		originalCommits = append(originalCommits, &RelevantCommit{
			CommitHash:  auditedCommit.CommitHash,
			RefStateKey: auditedCommit.RefStateKey,
		})
	}

	// We save all the results produced by the workers in a single
	// transaction. We do it this way because there is rate limit of 1 QPS
	// in a single entity group. (All relevant commits for a single repo
	// are contained in a single entity group)
	return ds.RunInTransaction(ctx, func(ctx context.Context) error {
		commitsToPut := make([]*RelevantCommit, 0, len(auditedCommits))
		if err := ds.Get(ctx, originalCommits); err != nil {
			return err
		}
		for _, currentCommit := range originalCommits {
			if auditedCommit, ok := auditedCommits[currentCommit.CommitHash]; ok {
				// Only save those that are still not in a decided
				// state in the datastore to avoid racing a possible
				// parallel run of this handler.
				if currentCommit.Status == auditScheduled || currentCommit.Status == auditPending {
					commitsToPut = append(commitsToPut, auditedCommit)
				}
			}
		}
		if err := ds.Put(ctx, commitsToPut); err != nil {
			return err
		}
		for _, c := range commitsToPut {
			if c.Status != auditScheduled {
				AuditedCommits.Add(ctx, 1, c.Status.ToShortString(), refState.ConfigName)
			}
		}
		return nil
	}, nil)
}

// notifyAboutViolations is meant to notify about violations to audit
// policies by calling the notification functions registered for each ruleSet
// that matches a commit in the auditCompletedWithActionRequired status.
func notifyAboutViolations(ctx context.Context, cfg *RefConfig, refState *RefState, cs *Clients) error {

	cfgk := ds.KeyForObj(ctx, refState)

	cq := ds.NewQuery("RelevantCommit").Ancestor(cfgk).Eq("Status", auditCompletedWithActionRequired).Eq("NotifiedAll", false)
	err := ds.Run(ctx, cq, func(rc *RelevantCommit) error {
		errors := []error{}
		var err error
		for ruleSetName, ruleSet := range cfg.Rules {
			if ruleSet.MatchesRelevantCommit(rc) {
				state := rc.GetNotificationState(ruleSetName)
				state, err = ruleSet.NotificationFunction()(ctx, cfg, rc, cs, state)
				if err == context.DeadlineExceeded {
					return err
				} else if err != nil {
					errors = append(errors, err)
				}
				rc.SetNotificationState(ruleSetName, state)
			}
		}
		if len(errors) == 0 {
			rc.NotifiedAll = true
		}
		for _, e := range errors {
			logging.WithError(e).Errorf(ctx, "Failed notification for detected violation on %s.",
				cfg.LinkToCommit(rc.CommitHash))
			NotificationFailures.Add(ctx, 1, "Violation", refState.ConfigName)
		}
		err = ds.Put(ctx, rc)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	cq = ds.NewQuery("RelevantCommit").Ancestor(cfgk).Eq("Status", auditFailed).Eq("NotifiedAll", false)
	return ds.Run(ctx, cq, func(rc *RelevantCommit) error {
		err := reportAuditFailure(ctx, cfg, rc, cs)

		if err == nil {
			rc.NotifiedAll = true
			err = ds.Put(ctx, rc)
			if err != nil {
				logging.WithError(err).Errorf(ctx, "Failed to save notification state for failed audit on %s.",
					cfg.LinkToCommit(rc.CommitHash))
				NotificationFailures.Add(ctx, 1, "AuditFailure", refState.ConfigName)
			}
		} else {
			logging.WithError(err).Errorf(ctx, "Failed to file bug for audit failure on %s.", cfg.LinkToCommit(rc.CommitHash))
			NotificationFailures.Add(ctx, 1, "AuditFailure", refState.ConfigName)
		}
		return nil
	})
}

// reportAuditFailure is meant to file a bug about a revision that has
// repeatedly failed to be audited. i.e. one or more rules return errors on
// each run.
//
// This does not necessarily mean that a policy has been violated, but only
// that the audit app has not been able to determine whether one exists. One
// such failure could be due to a bug in one of the rules or an error in one of
// the services we depend on (monorail, gitiles, gerrit).
func reportAuditFailure(ctx context.Context, cfg *RefConfig, rc *RelevantCommit, cs *Clients) error {
	summary := fmt.Sprintf("Audit on %q failed over %d times", rc.CommitHash, rc.Retries)
	description := fmt.Sprintf("commit %s has caused the audit process to fail repeatedly, "+
		"please audit by hand and don't close this bug until the root cause of the failure has been "+
		"identified and resolved.", cfg.LinkToCommit(rc.CommitHash))

	var err error
	issueID := int32(0)
	issueID, err = postIssue(ctx, cfg, summary, description, cs, []string{"Infra>Audit"}, []string{"AuditFailure"})
	if err == nil {
		rc.SetNotificationState("AuditFailure", fmt.Sprintf("BUG=%d", issueID))
		// Do not sent further notifications for this commit. This needs
		// to be audited by hand.
		rc.NotifiedAll = true
	}
	return err
}
