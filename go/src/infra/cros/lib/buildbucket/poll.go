package buildbucket

import (
	"context"
	"fmt"
	"time"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// PollForOutputProp polls until all of buildIds are completed or have set
// outputProp. Between each call to Buildbucket, the fn. sleeps for interval.
// This functionality is similar to the `bb collect` command, but allows polling
// for an output prop being set, instead of only build completion; this is
// useful for things like polling until a builder has published images.
//
// On completion, returns a map from build id -> Build for all of buildIds.
func PollForOutputProp(
	ctx context.Context,
	client bbpb.BuildsClient,
	buildIds []int64,
	outputProp string,
	interval time.Duration,
) (map[int64]*bbpb.Build, error) {
	// For simplicity, keep a single map of builds that have either completed or
	// set the requested output prop, and request all builds in each batch. This
	// means that if a build goes from running with output prop set ->
	// completed, the updated build will be put into the map. In the future, we
	// could keep a separate map of builds that are running with the output prop
	// set if needed.
	batchReq := &bbpb.BatchRequest{}
	for _, id := range buildIds {
		batchReq.Requests = append(batchReq.Requests, &bbpb.BatchRequest_Request{
			Request: &bbpb.BatchRequest_Request_GetBuild{
				GetBuild: &bbpb.GetBuildRequest{
					Id:     id,
					Fields: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
				},
			},
		})
	}

	completedBuilds := make(map[int64]*bbpb.Build)

	for {
		batchResp, err := client.Batch(ctx, batchReq)
		if err != nil {
			return nil, err
		}

		for _, resp := range batchResp.GetResponses() {
			var build *bbpb.Build
			switch resp.GetResponse().(type) {
			case *bbpb.BatchResponse_Response_GetBuild:
				build = resp.GetGetBuild()
			case *bbpb.BatchResponse_Response_Error:
				return nil, fmt.Errorf("got error in BatchResponse: %q", resp.GetError())
			default:
				return nil, fmt.Errorf("got unexpected response type: %q", resp)
			}

			// If the build has completed or set the output prop, add it to
			// completedBuilds.
			if (build.GetStatus() & bbpb.Status_ENDED_MASK) != 0 {
				logging.Infof(ctx, "build %d completed", build.GetId())
				completedBuilds[build.GetId()] = build
			} else if _, containsOutputProp := build.GetOutput().GetProperties().GetFields()[outputProp]; containsOutputProp {
				logging.Infof(ctx, "build %d has output prop (%q)", build.GetId(), outputProp)
				completedBuilds[build.GetId()] = build
			} else {
				logging.Infof(ctx, "still waiting for build %d", build.GetId())
			}
		}

		// Once all requested builds are completed or set the output prop,
		// break. Otherwise sleep for interval.
		if len(completedBuilds) == len(buildIds) {
			break
		}

		logging.Infof(ctx, "sleeping for %s", interval)
		time.Sleep(interval)
	}

	return completedBuilds, nil
}
