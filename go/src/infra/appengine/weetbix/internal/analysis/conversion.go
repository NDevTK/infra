package analysis

import (
	pb "infra/appengine/weetbix/proto/v1"
	"strings"
)

// ToBQBuildStatus converts a weetbix.v1.BuildStatus to its BigQuery
// column representation. This trims the BUILD_STATUS_ prefix to avoid
// excessive verbosity in the table.
func ToBQBuildStatus(value pb.BuildStatus) string {
	return strings.TrimPrefix(value.String(), "BUILD_STATUS_")
}

// FromBQBuildStatus extracts weetbix.v1.BuildStatus from
// its BigQuery column representation.
func FromBQBuildStatus(value string) pb.BuildStatus {
	return pb.BuildStatus(pb.BuildStatus_value["BUILD_STATUS_"+value])
}

// ToBQPresubmitRunStatus converts a weetbix.v1.PresubmitRunStatus to its
// BigQuery column representation. This trims the PRESUBMIT_RUN_STATUS_ prefix
// to avoid excessive verbosity in the table.
func ToBQPresubmitRunStatus(value pb.PresubmitRunStatus) string {
	return strings.TrimPrefix(value.String(), "PRESUBMIT_RUN_STATUS_")
}

// FromBQPresubmitRunStatus extracts weetbix.v1.PresubmitRunStatus from
// its BigQuery column representation.
func FromBQPresubmitRunStatus(value string) pb.PresubmitRunStatus {
	return pb.PresubmitRunStatus(pb.PresubmitRunStatus_value["PRESUBMIT_RUN_STATUS_"+value])
}

// ToBQPresubmitRunMode converts a weetbix.v1.PresubmitRunMode to its
// BigQuery column representation.
func ToBQPresubmitRunMode(value pb.PresubmitRunMode) string {
	return value.String()
}

// FromBQPresubmitRunMode extracts weetbix.v1.PresubmitRunMode from
// its BigQuery column representation.
func FromBQPresubmitRunMode(value string) pb.PresubmitRunMode {
	return pb.PresubmitRunMode(pb.PresubmitRunMode_value[value])
}

// FromBQExonerationReason extracts weetbix.v1.ExonerationReason from
// its BigQuery column representation.
func FromBQExonerationReason(value string) pb.ExonerationReason {
	return pb.ExonerationReason(pb.ExonerationReason_value[value])
}
