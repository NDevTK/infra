// Copyright 2017 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"

	"go.chromium.org/luci/common/bq"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	ds "go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/grpc/grpcutil"

	apibq "infra/tricium/api/bigquery"
	tricium "infra/tricium/api/v1"
	"infra/tricium/appengine/common"
	"infra/tricium/appengine/common/track"
)

// ReportNotUseful processes one report not useful request to Tricium.
func (r *TriciumServer) ReportNotUseful(c context.Context, req *tricium.ReportNotUsefulRequest) (res *tricium.ReportNotUsefulResponse, err error) {
	defer func() {
		err = grpcutil.GRPCifyAndLogErr(c, err)
	}()
	logging.Fields{
		"commentID": req.CommentId,
	}.Infof(c, "Request received.")
	if err = validateReportRequest(c, req); err != nil {
		return nil, err
	}
	response, err := reportNotUseful(c, req.CommentId)
	if err != nil {
		return nil, errors.Annotate(err, "report not useful request failed").
			Tag(grpcutil.InternalTag).Err()
	}
	return response, nil
}

func validateReportRequest(c context.Context, req *tricium.ReportNotUsefulRequest) error {
	if req.CommentId == "" {
		return errors.New("missing comment_id", grpcutil.InvalidArgumentTag)
	}
	return nil
}

func reportNotUseful(c context.Context, commentID string) (*tricium.ReportNotUsefulResponse, error) {
	comment, err := getCommentByID(c, commentID)
	if err != nil {
		return nil, err
	}

	if err = incrementCount(c, comment); err != nil {
		return nil, err
	}

	if err = streamToBigQuery(c, comment); err != nil {
		return nil, err
	}

	return responseForComment(comment), nil
}

// responseForComment returns bug-filing information for a comment.
//
// Looks up hard-coded bug owner and component information. This is intended to
// be a temporary hack until there is another way to specify bug filing
// information; see https://issues.chromium.org/40762296.
func responseForComment(comment *track.Comment) *tricium.ReportNotUsefulResponse {
	resp := &tricium.ReportNotUsefulResponse{
		Owner:             "yiwzhang@google.com",
		MonorailComponent: "Infra>LUCI>BuildService>PreSubmit>Tricium",
		// component for Tricium service:
		// https://issues.chromium.org/components/1456522
		IssueTrackerComponentId: 1456522,
	}
	switch name := comment.Analyzer; {
	case name == "Analyze" || strings.HasPrefix(name, "FuchsiaTricium"):
		resp.Owner = "olivernewman@google.com"
	case name == "ClangTidy" || name == "ChromiumTryLinuxClangTidyRel":
		resp.Owner = "gbiv@chromium.org"
	case name == "ChromiumosWrapper" || name == "ChromeosInfraTricium":
		resp.Owner = "bmgordon@google.com"
	case name == "Metrics" || name == "ChromiumTryTriciumMetricsAnalysis":
		resp.Owner = "isherman@chromium.org"
		resp.MonorailComponent = "Internals>Metrics>Tricium"
		resp.IssueTrackerComponentId = 1456158
	case name == "OilpanAnalyzer" || name == "ChromiumTryTriciumOilpanAnalysis":
		resp.Owner = "yukiy@chromium.org"
	case name == "PDFiumWrapper" || name == "PdfiumTryPdfiumAnalysis":
		resp.Owner = "nigi@chromium.org"
	}
	return resp
}

func getCommentByID(c context.Context, id string) (*track.Comment, error) {
	var comments []*track.Comment
	if err := ds.GetAll(c, ds.NewQuery("Comment").Eq("UUID", id), &comments); err != nil {
		return nil, errors.Annotate(err, "failed to get Comment").Err()
	}
	if len(comments) == 0 {
		return nil, errors.Reason("zero comments with UUID %q", id).Err()
	}
	if len(comments) > 1 {
		return nil, errors.Reason("multiple comments with UUID %q", id).Err()
	}
	return comments[0], nil
}

func incrementCount(c context.Context, comment *track.Comment) error {
	feedback := &track.CommentFeedback{ID: 1, Parent: ds.KeyForObj(c, comment)}
	if err := ds.Get(c, feedback); err != nil {
		return errors.Annotate(err, "failed to get CommentFeedback").Err()
	}
	feedback.NotUsefulReports++
	if err := ds.Put(c, feedback); err != nil {
		return errors.Annotate(err, "failed to store CommentFeedback").Err()
	}
	return nil
}

// streamToBigQuery adds an event row for the event of the not useful report.
func streamToBigQuery(c context.Context, comment *track.Comment) error {
	// The time used is the current time, but this time is not recorded in
	// datastore anywhere. Ideally the time used here should also be recorded
	// in datastore so that the data in BQ can be determined from datastore.
	// See crbug.com/943633.
	message := &tricium.Data_Comment{}
	if err := jsonpb.UnmarshalString(string(comment.Comment), message); err != nil {
		return errors.Annotate(err, "failed to unmarshal comment message").Err()
	}
	event := &apibq.FeedbackEvent{
		Type:     apibq.FeedbackEvent_NOT_USEFUL,
		Time:     ptypes.TimestampNow(),
		Comments: []*tricium.Data_Comment{message},
	}
	if err := common.EventsLog.Insert(c, &bq.Row{Message: event}); err != nil {
		return errors.Annotate(err, "failed in add row to bqlog.Log").Err()
	}
	return nil
}

func getFunctionRun(c context.Context, comment *track.Comment) (*track.FunctionRun, error) {
	commentKey := ds.KeyForObj(c, comment)
	functionKey := commentKey.Parent().Parent()
	function := &track.FunctionRun{
		ID:     functionKey.StringID(),
		Parent: functionKey.Parent(),
	}
	err := ds.Get(c, function)
	return function, err
}
