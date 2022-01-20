// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"fmt"
	"infra/rts"
	"infra/rts/filegraph/git"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/maruel/subcommands"
	luciflag "go.chromium.org/luci/common/flag"
	"google.golang.org/api/iterator"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"

	"go.chromium.org/luci/common/cli"
)

func cmdGenTrainingData(authOpt *auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `gen-training-data -out <path> -from <date> -to <date> -builder <builder>`,
		ShortDesc: "generate features and labels for ML model",
		LongDesc: text.Doc(`
			Benerate features and labels for ML model

			Flags -from -to -out are required.
		`),
		CommandRun: func() subcommands.CommandRun {
			r := &genTraingData{}
			r.authOpt = authOpt
			r.Flags.StringVar(&r.modelDir, "model-dir", "", text.Doc(`
				Path to the directory with the model files.
				Normally it is coming from CIPD package "chromium/rts/model"
				and precomputed by "rts-chromium create-model" command.
			`))
			r.Flags.StringVar(&r.in, "in", "", text.Doc(`
				Filename to append csv training data to. Does not duplicate data
			`))
			r.Flags.StringVar(&r.out, "out", "", text.Doc(`
				Filename to write csv training data to.
			`))
			r.Flags.StringVar(&r.testSuite, "test-suite", "", text.Doc(`
				Test suite to get training data for.
			`))
			r.Flags.StringVar(&r.builder, "builder", "", text.Doc(`
				Builder to get training data for.
			`))
			r.Flags.IntVar(&r.rowCount, "row-count", 100, text.Doc(`
				Max number of rows to process. Default: 100
			`))
			r.Flags.Var(luciflag.Date(&r.startTime), "from", "Fetch results starting from this date; format: yyyy-mm-dd")
			r.Flags.Var(luciflag.Date(&r.endTime), "to", "Fetch results until this date; format: yyyy-mm-dd")
			return r
		},
	}
}

type genTraingData struct {
	baseCommandRun

	rowCount  int
	startTime time.Time
	endTime   time.Time
	testSuite string
	builder   string

	in       string
	out      string
	modelDir string
	strategy git.SelectionStrategy

	authOpt       *auth.Options
	authenticator *auth.Authenticator
	http          *http.Client
}

func (r *genTraingData) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := r.ValidateFlags(); err != nil {
		return r.done(err)
	}

	ctx := cli.GetContext(a, r, env)

	entryHashes, err := r.ReadEntryHashes()
	if err != nil && err != io.EOF {
		return r.done(errors.Annotate(err, "failed to create BigQuery client").Err())
	}

	bqClient, err := newBQClient(ctx, auth.NewAuthenticator(ctx, auth.InteractiveLogin, *r.authOpt))
	if err != nil {
		return r.done(errors.Annotate(err, "failed to create BigQuery client").Err())
	}

	rows, err := r.queryResults(ctx, bqClient, entryHashes)
	if err != nil {
		return r.done(errors.Annotate(err, "failed to retrieve query results").Err())
	}

	r.loadInput(ctx)

	for i, row := range rows {
		if (i+1)%100 == 0 {
			fmt.Printf("Calculating distance on test %d\n", i+1)
		}
		if row.AffectedFilesCount > 100 || row.FileName == "" {
			rows[i].UseDistance = false
		} else {
			dist := r.calcDistance(row.AffectedFiles, row.FileName)

			valid := !math.IsInf(dist, 0)

			rows[i].Distance = dist
			rows[i].UseDistance = valid
		}
	}

	err = r.writeCsv(rows)
	if err != nil {
		return r.done(errors.Annotate(err, "failed to retrieve query results").Err())
	}

	return 0
}

func (r *genTraingData) writeCsv(rows []bqRow) error {
	var f *os.File
	var err error

	outfile := r.in
	if outfile == "" {
		//in was not delcared, we're creating the file
		outfile = r.out

		f, err = os.Create(r.out)
		if err != nil {
			return err
		}

		fmt.Fprintf(f, "ResultId,TestName,TestId,FileName,VariantHash,SixMonthRunCount,SixMonthFailRate,OneMonthRunCount,OneMonthFailRate,OneWeekRunCount,OneWeekFailRate,Distance,UseDistance,Expected,\n")
	} else {
		//in was delcared, we're appending the file
		f, err = os.OpenFile(outfile, os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
	}

	defer f.Close()
	for _, row := range rows {
		fmt.Fprintf(f, "%v,", row.ResultId)
		fmt.Fprintf(f, "%v,", row.TestName)
		fmt.Fprintf(f, "%v,", row.TestId)
		fmt.Fprintf(f, "%v,", row.FileName)
		fmt.Fprintf(f, "%v,", row.VariantHash)
		fmt.Fprintf(f, "%v,", row.SixMonthRunCount)
		fmt.Fprintf(f, "%f,", row.SixMonthFailRate)
		fmt.Fprintf(f, "%v,", row.OneMonthRunCount)
		fmt.Fprintf(f, "%f,", row.OneMonthFailRate)
		fmt.Fprintf(f, "%v,", row.OneWeekRunCount)
		fmt.Fprintf(f, "%f,", row.OneWeekFailRate)

		//Framework will handle missing values. Don't need to figure out average
		if row.UseDistance {
			fmt.Fprintf(f, "%f,", row.Distance)
		} else {
			fmt.Fprintf(f, ",")
		}

		fmt.Fprintf(f, "%v,", row.UseDistance)
		fmt.Fprintf(f, "%v,", row.Expected)
		fmt.Fprintf(f, "\n")
	}
	return f.Close()
}

func (r *genTraingData) ValidateFlags() error {
	switch {
	case r.out == "" && r.in == "":
		return errors.New("-out or -in is required")
	case r.startTime.IsZero():
		return errors.New("-from is required")
	case r.endTime.IsZero():
		return errors.New("-to is required")
	case r.endTime.Before(r.startTime):
		return errors.New("the -to date must not be before the -from date")
	default:
		return nil
	}
}

func (r *genTraingData) ReadEntryHashes() (map[string]interface{}, error) {
	hashset := make(map[string]interface{})
	if r.in == "" {
		return hashset, nil
	}

	f, err := os.Open(r.in)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	line, err := reader.ReadSlice('\n')
	for err == nil {
		hash := string(line[:strings.Index(string(line), ",")])
		hashset[hash] = nil
		line, err = reader.ReadSlice('\n')
	}
	return hashset, err
}

func (r *genTraingData) queryResults(ctx context.Context, bqClient *bigquery.Client, ignoreIds map[string]interface{}) ([]bqRow, error) {
	q := bqClient.Query(filtersQuery)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "rowCount", Value: r.rowCount},
		{Name: "startTime", Value: r.startTime},
		{Name: "endTime", Value: r.endTime},
		{Name: "testSuite", Value: r.testSuite},
		{Name: "builder", Value: r.builder},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	rows := []bqRow{}
	row := &bqRow{}
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Read the next row.
		switch err := it.Next(row); {
		case err == iterator.Done:
			return rows, ctx.Err()
		case err != nil:
			return nil, err
		}

		if _, exists := ignoreIds[row.ResultId]; !exists {
			for i, str := range row.AffectedFiles {
				row.AffectedFiles[i] = strings.Replace(str, "src/", "//", 1)
			}
			rows = append(rows, *row)

			if err == iterator.Done {
			}
		} else {
			fmt.Printf("Duplicate entry retrieved: %s", row.ResultId)
		}
	}
}

// loadInput loads all the input of the subcommand.
func (r *genTraingData) loadInput(ctx context.Context) error {
	gitGraphDir := filepath.Join(r.modelDir, "git-file-graph")
	return r.loadGraph(filepath.Join(gitGraphDir, "graph.fg"))
}

func (r *genTraingData) loadGraph(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// Note: it might be dangerous to sync with the current checkout.
	// There might have been such change in the repo that the chosen threshold,
	// the model or both are no longer good. Thus, do not sync.
	r.strategy.Graph = &git.Graph{}
	return r.strategy.Graph.Read(bufio.NewReader(f))
}

func (s *genTraingData) calcDistance(changedFiles []string, testFile string) float64 {
	found := false
	distance := 0.0
	s.strategy.RunQuery(changedFiles, func(name string, af rts.Affectedness) (keepGoing bool) {
		if name == testFile {
			found = true
			distance = af.Distance
			return false
		}
		return true
	})
	if found {
		return distance
	} else {
		return math.Inf(1)
	}
}

type bqRow struct {
	ResultId           string   `bigquery:"result_id"`
	TestName           string   `bigquery:"test_name"`
	TestId             string   `bigquery:"test_id"`
	FileName           string   `bigquery:"file_name"`
	VariantHash        string   `bigquery:"variant_hash"`
	SixMonthFailCount  int64    `bigquery:"six_month_fail_count"`
	SixMonthRunCount   int64    `bigquery:"six_month_run_count"`
	SixMonthFailRate   float32  `bigquery:"six_month_fail_rate"`
	OneMonthFailCount  int64    `bigquery:"one_month_fail_count"`
	OneMonthRunCount   int64    `bigquery:"one_month_run_count"`
	OneMonthFailRate   float32  `bigquery:"one_month_fail_rate"`
	OneWeekFailCount   int64    `bigquery:"one_week_fail_count"`
	OneWeekRunCount    int64    `bigquery:"one_week_run_count"`
	OneWeekFailRate    float32  `bigquery:"one_week_fail_rate"`
	AffectedFilesCount int64    `bigquery:"affected_files_count"`
	AffectedFiles      []string `bigquery:"affected_files"`
	Expected           bool     `bigquery:"expected"`
	Distance           float64
	UseDistance        bool
}

const filtersQuery = `
CREATE TEMP FUNCTION JsonToItems(input STRING)
RETURNS ARRAY<STRING>
LANGUAGE js AS """
return JSON.parse(input);
""";

WITH fail_rate as (
    SELECT 
        ds.test_id test_id,
        ds.variant_hash variant_hash,
        SUM(ARRAY_LENGTH(ds.patchsets_with_failures)) six_month_fail_count,
        SUM(ds.run_count) six_month_run_count,
        SUM(IF(day > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY), ARRAY_LENGTH(ds.patchsets_with_failures), 0)) one_month_fail_count,
        SUM(IF(day > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY), ds.run_count, 0)) one_month_run_count,
        SUM(IF(day > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY), ARRAY_LENGTH(ds.patchsets_with_failures), 0)) one_week_fail_count,
        SUM(IF(day > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY), ds.run_count, 0)) one_week_run_count,
    FROM
        chrome-trooper-analytics.test_results.daily_summary ds
    WHERE 
        ds.day > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 180 DAY)
    GROUP BY ds.test_id, ds.variant_hash
)

SELECT
    rdb.name test_name,
    rdb.test_id, 
    rdb.variant_hash,
    rdb.expected,
    rdb.result_id,
	IF(rdb.test_metadata.location.file_name IS NULL, "", REGEXP_REPLACE(rdb.test_metadata.location.file_name, r'^src/', r'//'))  file_name,
    (SELECT v.value FROM rdb.variant v where v.key = 'test_suite') AS test_suite,
    fr.six_month_fail_count,
    fr.six_month_run_count,
    IF(fr.six_month_run_count > 0, fr.six_month_fail_count / fr.six_month_run_count, 0) as six_month_fail_rate,
    fr.one_month_fail_count,
    fr.one_month_run_count,
    IF(fr.one_month_run_count > 0, fr.one_month_fail_count / fr.one_month_run_count, 0) as one_month_fail_rate,
    fr.one_week_fail_count,
    fr.one_week_run_count,
    IF(fr.one_week_run_count > 0, fr.one_week_fail_count / fr.one_week_run_count, 0) as one_week_fail_rate,
    rdb.expected passed,
    IF(JSON_EXTRACT(b.output.properties, "$.affected_files.total_count") IS NULL, 0, CAST(CAST(JSON_EXTRACT(b.output.properties, "$.affected_files.total_count") AS FLOAT64) AS INT)) affected_files_count, 
    JsonToItems(JSON_EXTRACT(b.output.properties, "$.affected_files.first_100")) affected_files
FROM
    chrome-luci-data.chromium.try_test_results rdb 
    LEFT JOIN fail_rate fr 
    ON rdb.test_id = fr.test_id AND rdb.variant_hash = fr.variant_hash
    LEFT JOIN cr-buildbucket.chromium.builds b 
    ON rdb.exported.id = 'build-' || b.id
WHERE
    (@testSuite = "" OR EXISTS (SELECT v.value FROM rdb.variant v where v.key = 'test_suite' AND v.value = @testSuite))
    AND (@builder = "" OR EXISTS (SELECT v.value FROM rdb.variant v where v.key = 'builder' AND v.value = @builder))
    AND fr.six_month_run_count IS NOT NULL
    AND fr.six_month_run_count > 0
    AND rdb.partition_time between @startTime and @endTime
ORDER BY rand()
LIMIT @rowCount
`
