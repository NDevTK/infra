// The Test Orchestrator takes a request specifying criteria for tests to run,
// computes an optimal set of tests / HW to run, schedules the tests, and
// processes the results.
//
// See design doc at go/ctp2-dd.
//
// This program implements the luciexe protocol, and can be run locally or on
// Buildbucket. See https://pkg.go.dev/go.chromium.org/luci/luciexe.
package main

import (
	"context"
	"fmt"

	testpb "go.chromium.org/chromiumos/config/go/test/api"
	tpv2 "go.chromium.org/chromiumos/infra/proto/go/test_platform/v2"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
)

func main() {
	request := &tpv2.RequestBeta{}
	build.Main(request, nil, nil, func(ctx context.Context, userArgs []string, state *build.State) error {
		return RunOrch(ctx, request)
	})
}

// RunOrch runs tests based on request.
func RunOrch(ctx context.Context, request *tpv2.RequestBeta) error {
	testSpecs := request.GetHwTestRequest().GetTestSpecs()
	if len(testSpecs) == 0 {
		return fmt.Errorf("at least one TestSpec in request required")
	}

	for _, spec := range testSpecs {
		swarmingDims, err := GetRequestedDimensions(ctx, spec.GetRules().DutCriteria)
		if err != nil {
			return err
		}

		logging.Infof(ctx, "Computed RequestedDimensions: %s", swarmingDims)
	}

	return nil
}

// GetRequestedDimensions gets RequestedDimensions for Swarming based on
// dutCriteria.
func GetRequestedDimensions(
	ctx context.Context, dutCriteria []*testpb.DutCriterion,
) (dims []*bbpb.RequestedDimension, err error) {
	step, _ := build.StartStep(ctx, "get requested dimensions")
	defer func() { step.End(err) }()

	if len(dutCriteria) == 0 {
		return nil, fmt.Errorf("at least one DutCriterion required in each CoverageRule")
	}

	dims = []*bbpb.RequestedDimension{}

	for _, criterion := range dutCriteria {
		key := criterion.GetAttributeId().GetValue()
		if key == "" {
			return nil, fmt.Errorf("DutAttribute id must be set")
		}

		values := criterion.GetValues()
		if len(values) == 0 {
			return nil, fmt.Errorf("at least one value must be set on DutAttributes")
		}

		dims = append(dims, &bbpb.RequestedDimension{
			Key:   key,
			Value: values[0],
		})
	}

	step.SetSummaryMarkdown(fmt.Sprintf("Computed %d RequestedDimensions", len(dims)))

	return dims, nil
}
