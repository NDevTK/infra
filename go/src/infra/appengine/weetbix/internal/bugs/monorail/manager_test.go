// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package monorail

import (
	"context"
	"infra/appengine/weetbix/internal/bugs"
	mpb "infra/monorailv2/api/v3/api_proto"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/genproto/protobuf/field_mask"
)

func NewCreateRequest() *bugs.CreateRequest {
	cluster := &bugs.CreateRequest{
		Title:       "ClusterID",
		Description: "Tests are failing with reason: Some failure reason.",
	}
	return cluster
}

func TestManager(t *testing.T) {
	t.Parallel()

	Convey("With Bug Manager", t, func() {
		ctx := context.Background()
		f := &FakeIssuesStore{
			NextID:            100,
			PriorityFieldName: "projects/chromium/fieldDefs/11",
		}
		user := AutomationUsers[0]
		cl, err := NewClient(UseFakeIssuesClient(ctx, f, user), "myhost")
		So(err, ShouldBeNil)
		monorailCfgs := ChromiumTestConfig()
		bm := NewBugManager(cl, monorailCfgs)

		Convey("Create", func() {
			c := NewCreateRequest()
			c.Impact = ChromiumHighP2Impact()
			Convey("With reason-based failure cluster", func() {
				reason := `Expected equality of these values:
					"Expected_Value"
					my_expr.evaluate(123)
						Which is: "Unexpected_Value"`
				c.Title = reason
				c.Description = "A cluster of failures has been found with reason: " + reason

				bug, err := bm.Create(ctx, c)
				So(err, ShouldBeNil)
				So(bug, ShouldEqual, "chromium/100")
				So(len(f.Issues), ShouldEqual, 1)
				issue := f.Issues[0]

				So(issue.Issue, ShouldResembleProto, &mpb.Issue{
					Name:     "projects/chromium/issues/100",
					Summary:  "Tests are failing: Expected equality of these values: \"Expected_Value\" my_expr.evaluate(123) Which is: \"Unexpected_Value\"",
					Reporter: AutomationUsers[0],
					State:    mpb.IssueContentState_ACTIVE,
					Status:   &mpb.Issue_StatusValue{Status: "Untriaged"},
					FieldValues: []*mpb.FieldValue{
						{
							// Type field.
							Field: "projects/chromium/fieldDefs/10",
							Value: "Bug",
						},
						{
							// Priority field.
							Field: "projects/chromium/fieldDefs/11",
							Value: "2",
						},
					},
					Labels: []*mpb.Issue_LabelValue{{
						Label: "Restrict-View-Google",
					}, {
						Label: "Weetbix-Managed",
					}},
				})
				So(len(issue.Comments), ShouldEqual, 1)
				So(issue.Comments[0].Content, ShouldContainSubstring, reason)
				So(issue.Comments[0].Content, ShouldNotContainSubstring, "ClusterIDShouldNotAppearInOutput")
				So(issue.NotifyCount, ShouldEqual, 1)
			})
			Convey("With test name failure cluster", func() {
				c.Title = "ninja://:blink_web_tests/media/my-suite/my-test.html"
				c.Description = "A test is failing " + c.Title

				bug, err := bm.Create(ctx, c)
				So(err, ShouldBeNil)
				So(bug, ShouldEqual, "chromium/100")
				So(len(f.Issues), ShouldEqual, 1)
				issue := f.Issues[0]

				So(issue.Issue, ShouldResembleProto, &mpb.Issue{
					Name:     "projects/chromium/issues/100",
					Summary:  "Tests are failing: ninja://:blink_web_tests/media/my-suite/my-test.html",
					Reporter: AutomationUsers[0],
					State:    mpb.IssueContentState_ACTIVE,
					Status:   &mpb.Issue_StatusValue{Status: "Untriaged"},
					FieldValues: []*mpb.FieldValue{
						{
							// Type field.
							Field: "projects/chromium/fieldDefs/10",
							Value: "Bug",
						},
						{
							// Priority field.
							Field: "projects/chromium/fieldDefs/11",
							Value: "2",
						},
					},
					Labels: []*mpb.Issue_LabelValue{{
						Label: "Restrict-View-Google",
					}, {
						Label: "Weetbix-Managed",
					}},
				})
				So(len(issue.Comments), ShouldEqual, 1)
				So(issue.Comments[0].Content, ShouldContainSubstring, "ninja://:blink_web_tests/media/my-suite/my-test.html")
				So(issue.NotifyCount, ShouldEqual, 1)
			})
			Convey("Does nothing if in simulation mode", func() {
				bm.Simulate = true
				_, err := bm.Create(ctx, c)
				So(err, ShouldEqual, bugs.ErrCreateSimulated)
				So(len(f.Issues), ShouldEqual, 0)
			})
		})
		Convey("Update", func() {
			c := NewCreateRequest()
			c.Impact = ChromiumP1Impact()
			bug, err := bm.Create(ctx, c)
			So(err, ShouldBeNil)
			So(bug, ShouldEqual, "chromium/100")
			So(len(f.Issues), ShouldEqual, 1)
			So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "1")

			bugToUpdate := &bugs.BugToUpdate{
				BugName: bug,
				Impact:  c.Impact,
			}
			bugsToUpdate := []*bugs.BugToUpdate{bugToUpdate}
			updateDoesNothing := func() {
				originalIssues := CopyIssuesStore(f)
				err := bm.Update(ctx, bugsToUpdate)
				So(err, ShouldBeNil)
				So(f, ShouldResembleIssuesStore, originalIssues)
			}
			// Create a monorail client that interacts with monorail
			// as an end-user. This is needed as we distinguish user
			// updates to the bug from system updates.
			user := "users/100"
			usercl, err := NewClient(UseFakeIssuesClient(ctx, f, user), "myhost")
			So(err, ShouldBeNil)

			Convey("If impact unchanged, does nothing", func() {
				updateDoesNothing()
			})
			Convey("If impact changed", func() {
				bugToUpdate.Impact = ChromiumP3Impact()
				Convey("Does not reduce priority if impact within hysteresis range", func() {
					c.Impact = ChromiumHighP2Impact()

					updateDoesNothing()
				})
				Convey("Reduces priority in response to reduced impact", func() {
					bugToUpdate.Impact = ChromiumP2Impact()
					originalNotifyCount := f.Issues[0].NotifyCount
					err := bm.Update(ctx, bugsToUpdate)
					So(err, ShouldBeNil)
					So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "2")

					// Does not notify.
					So(f.Issues[0].NotifyCount, ShouldEqual, originalNotifyCount)

					// Verify repeated update has no effect.
					updateDoesNothing()
				})
				Convey("Does not increase priority if impact within hysteresis range", func() {
					bugToUpdate.Impact = ChromiumLowP0Impact()

					updateDoesNothing()
				})
				Convey("Increases priority in response to increased impact", func() {
					bugToUpdate.Impact = ChromiumP0Impact()

					originalNotifyCount := f.Issues[0].NotifyCount
					err := bm.Update(ctx, bugsToUpdate)
					So(err, ShouldBeNil)
					So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "0")

					// Notified the increase.
					So(f.Issues[0].NotifyCount, ShouldEqual, originalNotifyCount+1)

					// Verify repeated update has no effect.
					updateDoesNothing()
				})
				Convey("Does not adjust priority if priority manually set", func() {
					updateReq := updateBugPriorityRequest(f.Issues[0].Issue.Name, "0")
					err = usercl.ModifyIssues(ctx, updateReq)
					So(err, ShouldBeNil)
					So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "0")

					// Check the update sets the label.
					expectedIssue := CopyIssue(f.Issues[0].Issue)
					expectedIssue.Labels = append(expectedIssue.Labels, &mpb.Issue_LabelValue{
						Label: manualPriorityLabel,
					})
					SortLabels(expectedIssue.Labels)

					So(f.Issues[0].NotifyCount, ShouldEqual, 1)
					err = bm.Update(ctx, bugsToUpdate)
					So(err, ShouldBeNil)
					So(f.Issues[0].Issue, ShouldResembleProto, expectedIssue)

					// Does not notify.
					So(f.Issues[0].NotifyCount, ShouldEqual, 1)

					// Check repeated update does nothing more.
					updateDoesNothing()

					Convey("Unless manual priority cleared", func() {
						updateReq := removeLabelRequest(f.Issues[0].Issue.Name, manualPriorityLabel)
						err = usercl.ModifyIssues(ctx, updateReq)
						So(err, ShouldBeNil)
						So(hasLabel(f.Issues[0].Issue, manualPriorityLabel), ShouldBeFalse)

						err := bm.Update(ctx, bugsToUpdate)
						So(err, ShouldBeNil)
						So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "3")

						// Verify repeated update has no effect.
						updateDoesNothing()
					})
				})
				Convey("Does nothing if in simulation mode", func() {
					bm.Simulate = true
					updateDoesNothing()
				})
				Convey("Does nothing if Restrict-View-Google is unset", func() {
					// This requirement comes from security review, see crbug.com/1245877.
					updateReq := removeLabelRequest(f.Issues[0].Issue.Name, restrictViewLabel)
					err = usercl.ModifyIssues(ctx, updateReq)
					So(err, ShouldBeNil)
					So(hasLabel(f.Issues[0].Issue, restrictViewLabel), ShouldBeFalse)

					updateDoesNothing()
				})
			})
			Convey("If impact falls below lowest priority threshold", func() {
				bugToUpdate.Impact = ChromiumClosureImpact()
				Convey("Update leaves bug open if impact within hysteresis range", func() {
					bugToUpdate.Impact = ChromiumClosureHighImpact()

					// Update may reduce the priority from P1 to P3, but the
					// issue should be left open. This is because hysteresis on
					// priority and issue verified state is applied separately.
					err := bm.Update(ctx, bugsToUpdate)
					So(err, ShouldBeNil)
					So(f.Issues[0].Issue.Status.Status, ShouldEqual, UntriagedStatus)
				})
				Convey("Update closes bug", func() {
					err := bm.Update(ctx, bugsToUpdate)
					So(err, ShouldBeNil)
					So(f.Issues[0].Issue.Status.Status, ShouldEqual, VerifiedStatus)

					// Verify repeated update has no effect.
					updateDoesNothing()

					Convey("Does not reopen bug if impact within hysteresis range", func() {
						c.Impact = ChromiumP3LowImpact()

						updateDoesNothing()
					})

					Convey("If impact increases, bug is re-opened with correct priority", func() {
						bugToUpdate.Impact = ChromiumP3Impact()
						Convey("Issue has owner", func() {
							// Update issue owner.
							updateReq := updateOwnerRequest(f.Issues[0].Issue.Name, "users/100")
							err = usercl.ModifyIssues(ctx, updateReq)
							So(err, ShouldBeNil)
							So(f.Issues[0].Issue.Owner.GetUser(), ShouldEqual, "users/100")

							// Issue should return to "Assigned" status.
							err := bm.Update(ctx, bugsToUpdate)
							So(err, ShouldBeNil)
							So(f.Issues[0].Issue.Status.Status, ShouldEqual, AssignedStatus)
							So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "3")

							// Verify repeated update has no effect.
							updateDoesNothing()
						})
						Convey("Issue has no owner", func() {
							// Issue should return to "Untriaged" status.
							err := bm.Update(ctx, bugsToUpdate)
							So(err, ShouldBeNil)
							So(f.Issues[0].Issue.Status.Status, ShouldEqual, UntriagedStatus)
							So(ChromiumTestIssuePriority(f.Issues[0].Issue), ShouldEqual, "3")

							// Verify repeated update has no effect.
							updateDoesNothing()
						})
					})
				})
			})
		})
	})
}

func updateOwnerRequest(name string, owner string) *mpb.ModifyIssuesRequest {
	return &mpb.ModifyIssuesRequest{
		Deltas: []*mpb.IssueDelta{
			{
				Issue: &mpb.Issue{
					Name: name,
					Owner: &mpb.Issue_UserValue{
						User: owner,
					},
				},
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"owner"},
				},
			},
		},
		CommentContent: "User comment.",
	}
}

func updateBugPriorityRequest(name string, priority string) *mpb.ModifyIssuesRequest {
	return &mpb.ModifyIssuesRequest{
		Deltas: []*mpb.IssueDelta{
			{
				Issue: &mpb.Issue{
					Name: name,
					FieldValues: []*mpb.FieldValue{
						{
							Field: "projects/chromium/fieldDefs/11",
							Value: priority,
						},
					},
				},
				UpdateMask: &field_mask.FieldMask{
					Paths: []string{"field_values"},
				},
			},
		},
		CommentContent: "User comment.",
	}
}

func removeLabelRequest(name string, label string) *mpb.ModifyIssuesRequest {
	return &mpb.ModifyIssuesRequest{
		Deltas: []*mpb.IssueDelta{
			{
				Issue: &mpb.Issue{
					Name: name,
				},
				UpdateMask:   &field_mask.FieldMask{},
				LabelsRemove: []string{label},
			},
		},
		CommentContent: "User comment.",
	}
}
