// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cloudsql

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"

	kronpb "go.chromium.org/chromiumos/infra/proto/go/test_platform/kron"
)

// InsertBuildsTemplate is the constant string template for how we will insert
// builds into the Cloud SQL PSQL database. The table name will need to be
// provided. "ON CONFLICT DO NOTHING" is added so that in the case of a
// duplicated build being inserted neither an error is returned nor, rows
// updated.
const InsertBuildsTemplate = "INSERT INTO \"%s\" (build_uuid,run_uuid,create_time,bbid,build_target,milestone,version,image_path,board,variant) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING;"

// SelectBuildsTemplate will be used as the basis to fetch the required builds
// for timed Event configs. The table name and where clause items will need to
// be provided.
//
// NOTE: The rank function here creates groupings of ("build_target",
// "milestone") pairs and gives the newest a value of rank 1. Using this we can
// ignore all but the latest image for the given build_target and milestone.
const SelectBuildsTemplate = `WITH
RankedBuilds AS (
SELECT
  *,
  RANK() OVER (PARTITION BY "build_target", "milestone" ORDER BY "create_time" DESC) AS "build_rank"
FROM
  "public"."%s" )
SELECT
	build_uuid, run_uuid, create_time, bbid, build_target, milestone, version,image_path, board, variant
FROM
	RankedBuilds
WHERE
	"build_rank" = 1 AND (%s)
ORDER BY
	"create_time" DESC`

const SelectWhereClauseItem = "(\"build_target\" = '%s' AND \"milestone\" = %d)"

// PSQLBuildRow is a PSQL compliant version of the kronpb.Build type. The
// difference here is that for the PSQL adapter we need to use
// pgtype.Timestamptz for timestamps rather than timestamppb.Timestamp.
type PSQLBuildRow struct {
	BuildUUID   string
	RunUUID     string
	CreateTime  pgtype.Timestamptz
	Bbid        int64
	BuildTarget string
	Milestone   int64
	Version     string
	ImagePath   string
	Board       string
	Variant     string
}

// ConvertBuildToPSQLRow converts a kronpb.Build into a PSQLBuildRow type.
func ConvertBuildToPSQLRow(build *kronpb.Build) (*PSQLBuildRow, error) {
	psqlRow := &PSQLBuildRow{
		BuildUUID:   build.GetBuildUuid(),
		RunUUID:     build.GetRunUuid(),
		Bbid:        build.GetBbid(),
		BuildTarget: build.GetBuildTarget(),
		Milestone:   build.GetMilestone(),
		Version:     build.GetVersion(),
		ImagePath:   build.GetImagePath(),
		Board:       build.GetBoard(),
		Variant:     build.GetVariant(),
	}

	// Get the build time as a time.Time and truncate to the nearest second to
	// remove any nanoseconds.
	buildTime := build.CreateTime.AsTime().Truncate(time.Second)

	// Populate the create_time field with the build time.
	err := psqlRow.CreateTime.Scan(buildTime)
	if err != nil {
		return nil, err
	}

	return psqlRow, nil
}

// ConvertPSQLRowToBuild converts a PSQLBuildRow into a kronpb.Build type.
func ConvertPSQLRowToBuild(Row *PSQLBuildRow) *kronpb.Build {
	build := &kronpb.Build{
		BuildUuid:   Row.BuildUUID,
		RunUuid:     Row.RunUUID,
		Bbid:        Row.Bbid,
		BuildTarget: Row.BuildTarget,
		Milestone:   Row.Milestone,
		Version:     Row.Version,
		ImagePath:   Row.ImagePath,
		Board:       Row.Board,
		Variant:     Row.Variant,
	}

	build.CreateTime = timestamppb.New(Row.CreateTime.Time)
	build.CreateTime.Nanos = 0

	return build
}

// RowToSlice converts the row type to a slice of field pointers. This is
// required when passing to the generic variadic functions for sql insertion.
func RowToSlice(row *PSQLBuildRow) []any {
	return []any{
		&row.BuildUUID,
		&row.RunUUID,
		&row.CreateTime,
		&row.Bbid,
		&row.BuildTarget,
		&row.Milestone,
		&row.Version,
		&row.ImagePath,
		&row.Board,
		&row.Variant,
	}
}

// ScanBuildRows handles row scans when querying build rows.
func ScanBuildRows(rows pgx.Rows) (any, error) {
	builds := []*kronpb.Build{}
	for rows.Next() {
		row := &PSQLBuildRow{}

		err := rows.Scan(RowToSlice(row)...)
		if err != nil {
			return nil, err
		}

		builds = append(builds, ConvertPSQLRowToBuild(row))
	}

	return builds, nil
}
