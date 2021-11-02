// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package monorail

import (
	"fmt"
	"regexp"
	"strings"

	"infra/appengine/weetbix/internal/bugs"
	"infra/appengine/weetbix/internal/config"
	mpb "infra/monorailv2/api/v3/api_proto"

	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	DescriptionTemplate = `%s

This bug has been automatically filed by Weetbix in response to a cluster of test failures.`
)

const (
	manualPriorityLabel = "Weetbix-Manual-Priority"
	restrictViewLabel   = "Restrict-View-Google"
	managedLabel        = "Weetbix-Managed"
)

// whitespaceRE matches blocks of whitespace, including new lines tabs and
// spaces.
var whitespaceRE = regexp.MustCompile(`[ \t\n]+`)

// priorityRE matches chromium monorail priority values.
var priorityRE = regexp.MustCompile(`^Pri-([0123])$`)

// AutomationUsers are the identifiers of Weetbix automation users in monorail.
var AutomationUsers = []string{
	"users/4149141945", // chops-weetbix-dev@appspot.gserviceaccount.com
}

// VerifiedStatus is that status of bugs that have been fixed and verified.
const VerifiedStatus = "Verified"

// AssignedStatus is the status of bugs that are open and assigned to an owner.
const AssignedStatus = "Assigned"

// UntriagedStatus is the status of bugs that have just been opened.
const UntriagedStatus = "Untriaged"

// Generator provides access to a methods to generate a new bug and/or bug
// updates for a cluster.
type Generator struct {
	// The cluster to generate monorail changes for.
	cluster *bugs.Cluster
	// The monorail configuration to use.
	monorailCfg *config.MonorailProject
}

// NewGenerator initialises a new Generator.
func NewGenerator(cluster *bugs.Cluster, monorailCfg *config.MonorailProject) (*Generator, error) {
	if len(monorailCfg.Priorities) == 0 {
		return nil, fmt.Errorf("invalid configuration for monorail project %q; no monorail priorities configured", monorailCfg.Project)
	}
	return &Generator{
		cluster:     cluster,
		monorailCfg: monorailCfg,
	}, nil
}

// PrepareNew prepares a new bug from the given cluster.
func (g *Generator) PrepareNew() *mpb.MakeIssueRequest {
	title := g.cluster.DisplayName
	issue := &mpb.Issue{
		Summary: fmt.Sprintf("Tests are failing: %v", sanitiseTitle(title, 150)),
		State:   mpb.IssueContentState_ACTIVE,
		Status:  &mpb.Issue_StatusValue{Status: UntriagedStatus},
		FieldValues: []*mpb.FieldValue{
			{
				Field: g.priorityFieldName(),
				Value: g.clusterPriority(),
			},
		},
		Labels: []*mpb.Issue_LabelValue{{
			Label: restrictViewLabel,
		}, {
			Label: managedLabel,
		}},
	}
	for _, fv := range g.monorailCfg.DefaultFieldValues {
		issue.FieldValues = append(issue.FieldValues, &mpb.FieldValue{
			Field: fmt.Sprintf("projects/%s/fieldDefs/%v", g.monorailCfg.Project, fv.FieldId),
			Value: fv.Value,
		})
	}

	return &mpb.MakeIssueRequest{
		Parent:      fmt.Sprintf("projects/%s", g.monorailCfg.Project),
		Issue:       issue,
		Description: g.bugDescription(),
		NotifyType:  mpb.NotifyType_EMAIL,
	}
}

func (g *Generator) priorityFieldName() string {
	return fmt.Sprintf("projects/%s/fieldDefs/%v", g.monorailCfg.Project, g.monorailCfg.PriorityFieldId)
}

// NeedsUpdate determines if the bug for the given cluster needs to be updated.
func (g *Generator) NeedsUpdate(issue *mpb.Issue) bool {
	// Bugs must have restrict view label to be updated.
	if !hasLabel(issue, restrictViewLabel) {
		return false
	}
	// Cases that a bug may be updated follow.
	switch {
	case !g.isCompatibleWithVerified(issueVerified(issue)):
		return true
	case !hasLabel(issue, manualPriorityLabel) &&
		!issueVerified(issue) &&
		!g.isCompatibleWithPriority(g.IssuePriority(issue)):
		// The priority has changed on a cluster which is not verified as fixed
		// and the user isn't manually controlling the priority.
		return true
	default:
		return false
	}
}

// MakeUpdate prepares an updated for the bug associated with a given cluster.
// Must ONLY be called if NeedsUpdate(...) returns true.
func (g *Generator) MakeUpdate(issue *mpb.Issue, comments []*mpb.Comment) *mpb.ModifyIssuesRequest {
	delta := &mpb.IssueDelta{
		Issue: &mpb.Issue{
			Name: issue.Name,
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{},
		},
	}

	var commentary []string
	notify := false
	issueVerified := issueVerified(issue)
	if !g.isCompatibleWithVerified(issueVerified) {
		// Verify or reopen the issue.
		comment := g.prepareBugVerifiedUpdate(issue, delta)
		commentary = append(commentary, comment)
		notify = true
		// After the update, whether the issue was verified will have changed.
		issueVerified = g.clusterResolved()
	}
	if !hasLabel(issue, manualPriorityLabel) &&
		!issueVerified &&
		!g.isCompatibleWithPriority(g.IssuePriority(issue)) {

		if hasManuallySetPriority(comments) {
			// We were not the last to update the priority of this issue.
			// Set the 'manually controlled priority' label to reflect
			// the state of this bug and avoid further attempts to update.
			comment := prepareManualPriorityUpdate(issue, delta)
			commentary = append(commentary, comment)
		} else {
			// We were the last to update the bug priority.
			// Apply the priority update.
			comment := g.preparePriorityUpdate(issue, delta)
			commentary = append(commentary, comment)
			// Notify if new priority is higher than existing priority.
			notify = notify || g.isHigherPriority(g.clusterPriority(), g.IssuePriority(issue))
		}
	}

	update := &mpb.ModifyIssuesRequest{
		Deltas: []*mpb.IssueDelta{
			delta,
		},
		NotifyType:     mpb.NotifyType_NO_NOTIFICATION,
		CommentContent: strings.Join(commentary, "\n\n"),
	}
	if notify {
		update.NotifyType = mpb.NotifyType_EMAIL
	}
	return update
}

func (g *Generator) prepareBugVerifiedUpdate(issue *mpb.Issue, update *mpb.IssueDelta) string {
	resolved := g.clusterResolved()
	var status string
	var comment string
	if resolved {
		status = VerifiedStatus
		comment = "No further occurances of the failure cluster have been identified. Weetbix is marking the issue verified."
	} else {
		if issue.GetOwner().GetUser() != "" {
			status = AssignedStatus
		} else {
			status = UntriagedStatus
		}
		comment = "Weetbix has identified new occurances of the failure cluster. The bug has been re-opened."
	}
	update.Issue.Status = &mpb.Issue_StatusValue{Status: status}
	update.UpdateMask.Paths = append(update.UpdateMask.Paths, "status")
	return comment
}

func prepareManualPriorityUpdate(issue *mpb.Issue, update *mpb.IssueDelta) string {
	update.Issue.Labels = []*mpb.Issue_LabelValue{{
		Label: manualPriorityLabel,
	}}
	update.UpdateMask.Paths = append(update.UpdateMask.Paths, "labels")
	return fmt.Sprintf("The bug priority has been manually set. To re-enable automatic priority updates by Weetbix, remove the %s label.", manualPriorityLabel)
}

func (g *Generator) preparePriorityUpdate(issue *mpb.Issue, update *mpb.IssueDelta) string {
	update.Issue.FieldValues = []*mpb.FieldValue{
		{
			Field: g.priorityFieldName(),
			Value: g.clusterPriority(),
		},
	}
	update.UpdateMask.Paths = append(update.UpdateMask.Paths, "field_values")
	return fmt.Sprintf("The impact of this bug's test failures has changed. "+
		"Weetbix has adjusted the bug priority from %v to %v.", g.IssuePriority(issue), g.clusterPriority())
}

// hasManuallySetPriority returns whether the the given issue has a manually
// controlled priority, based on its comments.
func hasManuallySetPriority(comments []*mpb.Comment) bool {
	// Example comment showing a user changing priority:
	// {
	// 	name: "projects/chromium/issues/915761/comments/1"
	// 	state: ACTIVE
	// 	type: COMMENT
	// 	commenter: "users/2627516260"
	// 	create_time: {
	// 	  seconds: 1632111572
	// 	}
	// 	amendments: {
	// 	  field_name: "Labels"
	// 	  new_or_delta_value: "Pri-1"
	// 	}
	// }
	for i := len(comments) - 1; i >= 0; i-- {
		c := comments[i]

		isManualPriorityUpdate := false
		isRevertToAutomaticPriority := false
		for _, a := range c.Amendments {
			if a.FieldName == "Labels" {
				deltaLabels := strings.Split(a.NewOrDeltaValue, " ")
				for _, lbl := range deltaLabels {
					if lbl == "-"+manualPriorityLabel {
						isRevertToAutomaticPriority = true
					}
					if priorityRE.MatchString(lbl) {
						if !isAutomationUser(c.Commenter) {
							isManualPriorityUpdate = true
						}
					}
				}
			}
		}
		if isRevertToAutomaticPriority {
			return false
		}
		if isManualPriorityUpdate {
			return true
		}
	}
	// No manual changes to priority indicates the bug is still under
	// automatic control.
	return false
}

func isAutomationUser(user string) bool {
	for _, u := range AutomationUsers {
		if u == user {
			return true
		}
	}
	return false
}

// hasLabel returns whether the bug the specified label.
func hasLabel(issue *mpb.Issue, label string) bool {
	for _, l := range issue.Labels {
		if l.Label == label {
			return true
		}
	}
	return false
}

// IssuePriority returns the priority of the given issue.
func (g *Generator) IssuePriority(issue *mpb.Issue) string {
	priorityFieldName := g.priorityFieldName()
	for _, fv := range issue.FieldValues {
		if fv.Field == priorityFieldName {
			return fv.Value
		}
	}
	return ""
}

func issueVerified(issue *mpb.Issue) bool {
	return issue.Status.Status == VerifiedStatus
}

// isHigherPriority returns whether priority p1 is higher than priority p2.
// The passed strings are the priority field values as used in monorail. These
// must be matched against monorail project configuration in order to
// identify the ordering of the priorities.
func (g *Generator) isHigherPriority(p1 string, p2 string) bool {
	i1 := g.indexOfPriority(p1)
	i2 := g.indexOfPriority(p2)
	// Priorities are configured from highest to lowest, so higher priorities
	// have lower indexes.
	return i1 < i2
}

func (g *Generator) indexOfPriority(priority string) int {
	for i, p := range g.monorailCfg.Priorities {
		if p.Priority == priority {
			return i
		}
	}
	// If we can't find the priority, treat it as one lower than
	// the lowest priority we know about.
	return len(g.monorailCfg.Priorities)
}

// bugDescription returns the description that should be used when creating
// a new bug for the cluster.
func (g *Generator) bugDescription() string {
	return fmt.Sprintf(DescriptionTemplate, g.cluster.Description)
}

// isCompatibleWithVerified returns whether the impact of the current cluster
// is compatible with the issue having the given verified status, based on
// configured thresholds and hysteresis.
func (g *Generator) isCompatibleWithVerified(verified bool) bool {
	hysteresisPerc := g.monorailCfg.PriorityHysteresisPercent
	lowestPriority := g.monorailCfg.Priorities[len(g.monorailCfg.Priorities)-1]
	if verified {
		// The issue is verified. Only reopen if there is enough impact
		// to exceed the threshold with hysteresis.
		return !g.cluster.Impact.MeetsInflatedThreshold(lowestPriority.Threshold, hysteresisPerc)
	} else {
		// The issue is not verified. Only close if the impact falls
		// below the threshold with hysteresis.
		return g.cluster.Impact.MeetsInflatedThreshold(lowestPriority.Threshold, -hysteresisPerc)
	}
}

// isCompatibleWithPriority returns whether the impact of the current cluster
// is compatible with the issue having the given priority, based on
// configured thresholds and hysteresis.
func (g *Generator) isCompatibleWithPriority(issuePriority string) bool {
	index := g.indexOfPriority(issuePriority)
	if index >= len(g.monorailCfg.Priorities) {
		// Unknown priority in use. The priority should be updated to
		// one of the configured priorities.
		return false
	}

	p := g.monorailCfg.Priorities[index]
	var nextP *config.MonorailPriority
	if (index - 1) >= 0 {
		nextP = g.monorailCfg.Priorities[index-1]
	}
	hysteresisPerc := g.monorailCfg.PriorityHysteresisPercent
	// The cluster does not satisfy its current priority if it falls below
	// the current priority's thresholds, even after deflating them by
	// the hystersis margin.
	if !g.cluster.Impact.MeetsInflatedThreshold(p.Threshold, -hysteresisPerc) {
		return false
	}
	// It also does not satisfy its current priority if it meets the
	// the next priority's priority's thresholds, after inflating them by
	// the hystersis margin. (Assuming there exists a higher priority.)
	if nextP != nil && g.cluster.Impact.MeetsInflatedThreshold(nextP.Threshold, hysteresisPerc) {
		return false
	}
	return true
}

// clusterPriority returns the desired priority of the bug, if no hysteresis
// is applied.
func (g *Generator) clusterPriority() string {
	// Default to using the lowest priority.
	priority := g.monorailCfg.Priorities[len(g.monorailCfg.Priorities)-1]
	for i := len(g.monorailCfg.Priorities) - 2; i >= 0; i-- {
		p := g.monorailCfg.Priorities[i]
		if !g.cluster.Impact.MeetsThreshold(p.Threshold) {
			// A cluster cannot reach a higher priority unless it has
			// met the thresholds for all lower priorities.
			break
		}
		priority = p
	}
	return priority.Priority
}

// clusterResolved returns the desired state of whether the cluster has been
// verified, if no hysteresis has been applied.
func (g *Generator) clusterResolved() bool {
	lowestPriority := g.monorailCfg.Priorities[len(g.monorailCfg.Priorities)-1]
	return !g.cluster.Impact.MeetsThreshold(lowestPriority.Threshold)
}

// sanitiseTitle removes tabs and line breaks from input, replacing them with
// spaces, and truncates the output to the given number of runes.
func sanitiseTitle(input string, maxLength int) string {
	// Replace blocks of whitespace, including new lines and tabs, with just a
	// single space.
	strippedInput := whitespaceRE.ReplaceAllString(input, " ")

	// Truncate to desired length.
	runes := []rune(strippedInput)
	if len(runes) > maxLength {
		return string(runes[0:maxLength-3]) + "..."
	}
	return strippedInput
}
