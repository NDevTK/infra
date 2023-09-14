// Copyright 2021 The Chromium Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"context"
	"encoding/json"
	"flag"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/aclements/go-moremath/stats"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
	"gonum.org/v1/gonum/stat"
	"gopkg.in/yaml.v2"

	"infra/chromeperf/histograms"
	"infra/chromeperf/pinpoint"
	"infra/chromeperf/pinpoint/cli/render"
	"infra/chromeperf/pinpoint/proto"
)

type metricNameKey string
type expNameKey string

const (
	baseLabel expNameKey = "base"
	expLabel  expNameKey = "exp"
)

func loadManifestFromJob(baseDir string, j *proto.Job) (*telemetryExperimentArtifactsManifest, error) {
	jobId, err := render.JobID(j)
	if err != nil {
		return nil, err
	}
	jobDir := path.Join(baseDir, jobId)
	manifest, err := loadManifestFromPath(path.Join(jobDir, "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func loadManifestFromPath(m string) (*telemetryExperimentArtifactsManifest, error) {
	a := &telemetryExperimentArtifactsManifest{}
	d, err := os.ReadFile(m)
	if err != nil {
		return nil, errors.Annotate(err, "failed reading manifest").Err()
	}
	if err := yaml.Unmarshal(d, a); err != nil {
		return nil, errors.Annotate(err, "failed unmarshaling manifest").Err()
	}
	return a, nil
}

type measurementSummary struct {
	Min    float64   `yaml:"min"`
	Median float64   `yaml:"median"`
	Mean   float64   `yaml:"mean"`
	Stddev float64   `yaml:"stddev"`
	Pct90  float64   `yaml:"pct90"`
	Pct99  float64   `yaml:"pct99"`
	Max    float64   `yaml:"max"`
	Count  int       `yaml:"count"`
	Raw    []float64 `yaml:"raw"`
}

type statTestSummary struct {
	Mean   float64 `yaml:"mean"`
	Stddev float64 `yaml:"stddev"`
}

type statTestMap map[expNameKey]statTestSummary
type measurementMap map[expNameKey]measurementSummary

type measurementReport struct {
	StatTestSummary statTestMap    `yaml:"stat-test-summary"` // map[base or exp]
	PValue          *float64       `yaml:"p-value"`
	Measurements    measurementMap `yaml:"measurements"` // map[base or exp]
	ErrorMessage    string         `yaml:"error-message" json:",omitempty"`
}

type reportMap map[metricNameKey]measurementReport

type experimentReport struct {
	OverallPValue float64   `yaml:"overall-p-value"`
	Alpha         float64   `yaml:"alpha"`
	Reports       reportMap `yaml:"reports"` // map[metric_name]
}

type reportMapKV struct {
	Name   metricNameKey
	Report measurementReport
}

type statTestMapKV struct {
	Name    expNameKey
	Summary statTestSummary
}

type measurmentMapKV struct {
	Name    expNameKey
	Summary measurementSummary
}

func (sm statTestMap) MarshalJSON() ([]byte, error) {
	m := []statTestMapKV{}
	for k, v := range sm {
		m = append(m, statTestMapKV{k, v})
	}

	return json.Marshal(m)
}

func (mm measurementMap) MarshalJSON() ([]byte, error) {
	m := []measurmentMapKV{}
	for k, v := range mm {
		m = append(m, measurmentMapKV{k, v})
	}
	return json.Marshal(m)
}

func (rm reportMap) MarshalJSON() ([]byte, error) {
	m := []reportMapKV{}
	for k, r := range rm {
		m = append(m, reportMapKV{k, r})
	}
	return json.Marshal(m)
}

func loadAndMergeHistograms(config *changeConfig, rootDir string) ([]*histograms.Histogram, error) {
	hs := []*histograms.Histogram{}
	for _, a := range config.Artifacts {
		if a.Selector != "test" {
			continue
		}
		for _, f := range a.Files {
			// In the manifest, the path relates to the root of the output
			// directory. This means we need to look for specific files in that
			// directory, in this case we're looking for `perf_results.json` by
			// convention. Theoretically we could be using other formats, but
			// this is the current format supported/generated by Telemetry via
			// TBMv2.
			dir := filepath.Join(rootDir, filepath.FromSlash(f.Path))
			if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if d.Name() == "perf_results.json" {
					jf, err := os.Open(path)
					if err != nil {
						return errors.Annotate(err, "failed loading file: %q", path).Err()
					}
					defer jf.Close()
					h, err := histograms.NewFromJSON(jf)
					// TODO: make this concurrent?
					hs = append(hs, h...)
				}
				return nil
			}); err != nil {
				return nil, err
			}
		}
	}
	return hs, nil
}

func analyzeExperiment(manifest *telemetryExperimentArtifactsManifest, rootDir string) (*experimentReport, error) {
	r := &experimentReport{
		Reports: make(map[metricNameKey]measurementReport),
	}

	// First thing we do is load up the data from the files referred to by the
	// manifest. In the end we'll have a structure which contains the
	// measurements/samples from all the histograms associated with the manifest
	// grouped by whether they're from the base or the experiment commit.
	data := make(map[metricNameKey]map[expNameKey]*histograms.Histogram) // map[metric_name][baseOrExp]
	for _, c := range []struct {
		label  expNameKey
		config *changeConfig
	}{
		{baseLabel, &manifest.Base},
		{expLabel, &manifest.Experiment},
	} {
		// For each of the artifact files, we want to load the histograms.
		hs := []*histograms.Histogram{}
		hs, err := loadAndMergeHistograms(c.config, rootDir)
		if err != nil {
			return nil, err
		}

		// Now that we have all the histograms associated with the change, we'll
		// then merge the ones that have the same name.
		hm := make(map[string]*histograms.Histogram)
		for _, h := range hs {
			if orig, found := hm[h.Name]; !found {
				hm[h.Name] = h
			} else {
				// Append the values into the already found histogram in the list.
				orig.SampleValues = append(orig.SampleValues, h.SampleValues...)
			}
		}

		// At this point we should have the summary of the values for a given change.
		// We can then place this "inside-out" from the histogram name in the
		// outer map key, and the label in the inner map key.
		for n, h := range hm {
			n := metricNameKey(n)
			m, found := data[n]
			if !found {
				data[n] = make(map[expNameKey]*histograms.Histogram)
				m = data[n]
			}
			m[c.label] = h
		}
	}

	// Conceptually, the `data` map gives us the following table:
	//
	//  measurement  | base    | experiment
	//  -------------|---------|------------
	//  hist1        | [ ... ] | [ ... ]
	//  hist2        | [ ... ] | [ ... ]
	//  ...          | ...     | ...
	//  histN        | [ ... ] | [ ... ]
	//
	// Given the list of samples for the base and experiment, we can start
	// performing the side-by-side comparison and compute the p-value from a
	// Mann-Whitney U-test between the samples.
	//
	//  measurement  | base    | experiment | p-value
	//  -------------|---------|------------|---------
	//  hist1        | [ ... ] | [ ... ]    | 0.xxxx
	//  hist2        | [ ... ] | [ ... ]    | 0.xxxx
	//  ...          | ...     | ...        | ...
	//  histN        | [ ... ] | [ ... ]    | 0.xxxx
	//
	//
	pvs := []float64{}
	for m, v := range data {
		mr := measurementReport{
			StatTestSummary: make(map[expNameKey]statTestSummary),
			Measurements:    make(map[expNameKey]measurementSummary),
		}
		as := []float64{}
		bs := []float64{}
		for l, h := range v {
			if len(h.SampleValues) == 0 {
				continue
			}
			sort.Float64s(h.SampleValues)
			switch l {
			case baseLabel:
				as = h.SampleValues
			case expLabel:
				bs = h.SampleValues
			}

			// Compute the minimum and the maximum by hand, because we don't have generics yet.
			minS := h.SampleValues[0]
			maxS := h.SampleValues[0]
			for _, s := range h.SampleValues {
				minS = math.Min(minS, s)
				maxS = math.Max(maxS, s)
			}

			ms := measurementSummary{
				Min:    minS,
				Median: stat.Quantile(0.5, stat.Empirical, h.SampleValues, nil),
				Mean:   stat.Mean(h.SampleValues, nil),
				Stddev: stat.StdDev(h.SampleValues, nil),
				Pct90:  stat.Quantile(0.9, stat.Empirical, h.SampleValues, nil),
				Pct99:  stat.Quantile(0.99, stat.Empirical, h.SampleValues, nil),
				Max:    maxS,
				Count:  len(h.SampleValues),
				Raw:    h.SampleValues,
			}
			mr.Measurements[l] = ms
		}
		// TODO: Use the unit information to determine whether to use a
		// one-tail or two-tail test.
		mwur, err := stats.MannWhitneyUTest(as, bs, stats.LocationDiffers)
		if err != nil {
			mr.ErrorMessage = err.Error()
		} else {
			p := mwur.P
			mr.PValue = &p
			pvs = append(pvs, mwur.P)
		}
		r.Reports[m] = mr
	}

	// We'll use the harmonic mean of the p-values to determine whether overall
	// we can detect a difference between the base and experiment. For more
	// explanations on why we're using this instead of the Fisher's method, see
	// https://en.wikipedia.org/wiki/Harmonic_mean_p-value.
	r.OverallPValue = math.NaN()
	if len(pvs) > 0 {
		r.OverallPValue = stat.HarmonicMean(pvs, nil)
	}
	return r, nil
}

type analyzeExperimentMixin struct {
	analyzeExperiment, check bool
}

func (aem *analyzeExperimentMixin) RegisterFlags(flags *flag.FlagSet, userCfg userConfig) {
	flags.BoolVar(&aem.analyzeExperiment, "analyze-experiment", userCfg.AnalyzeExperiment, text.Doc(`
		If set, artifacts associated with the job are downloaded (see
		-download-artifacts and -work-dir) and analyzed to generate a report.
		Override the default from the user configuration file.
	`))
	flags.BoolVar(&aem.check, "check-experiment", userCfg.CheckExperiment, text.Doc(`
		If set, the command will return an error if we end up rejecting the null
		hypothesis from the experiment (i.e. when we can detect a statistically
		significant difference). Override the default from the user
		configuration file.
	`))

}

func (aem *analyzeExperimentMixin) doAnalyzeExperiment(ctx context.Context, workDir string, job *proto.Job) (*experimentReport, error) {
	if !aem.analyzeExperiment || job.GetName() == "" {
		return nil, nil
	}
	switch job.GetJobSpec().GetJobKind().(type) {
	case *proto.JobSpec_Bisection:
		return nil, errors.Reason("not implemented").Err()
	case *proto.JobSpec_Experiment:
		return aem.analyzeTelemetryExperiment(ctx, workDir, job)
	default:
		return nil, errors.Reason("unsupported job kind").Err()
	}
}

func (aem *analyzeExperimentMixin) analyzeTelemetryExperiment(ctx context.Context, workDir string, job *proto.Job) (*experimentReport, error) {
	id, err := pinpoint.ExtractJobID(job.Name)
	if err != nil {
		return nil, errors.Annotate(err, "invalid job id").Err()
	}
	jp := filepath.Join(workDir, id)
	jm, err := loadManifestFromPath(filepath.Join(jp, "manifest.yaml"))
	if err != nil {
		return nil, errors.Annotate(err, "couldn't load manifest").Err()
	}
	r, err := analyzeExperiment(jm, jp)
	if err != nil {
		return nil, errors.Annotate(err, "failed analysis").Err()
	}
	return r, nil
}

type analyzeRun struct {
	baseCommandRun
	downloadArtifactsMixin
	params  Param
	jobName string
	check   bool
}

func (ar *analyzeRun) RegisterFlags(p Param) {
	uc := ar.baseCommandRun.RegisterFlags(p)
	ar.downloadArtifactsMixin.RegisterFlags(&ar.Flags, uc)
	ar.Flags.BoolVar(&ar.check, "check", uc.CheckExperiment, text.Doc(`
		Return a non-zero exit status if there are statistically significant
		differences detected in the experiment.
	`))
	ar.Flags.StringVar(&ar.jobName, "job-name", "", text.Doc(`
		The job id to analyze.
	`))
}

func cmdAnalyzeExperiment(p Param) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "analyze-experiment -job-name ... [-check]",
		ShortDesc: "analyzes the results of an experiment",
		LongDesc: text.Doc(`
		analyze-experiment will perform statistical analysis on the artifacts
		associated with a job. When '-check' is provided, the command will
		return a non-zero exit status in case there are statistically signficant
		differences detected in the experiment.
		`),
		CommandRun: wrapCommand(p, func() pinpointCommand {
			return &analyzeRun{}
		}),
	}
}

func (ar *analyzeRun) Run(ctx context.Context, a subcommands.Application, args []string) error {
	c, err := ar.pinpointClient(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to create a Pinpoint client").Err()
	}

	h, err := ar.httpClient(ctx)
	if err != nil {
		return errors.Annotate(err, "failed creating an http client").Err()
	}

	j, err := c.GetJob(ctx, &proto.GetJobRequest{Name: pinpoint.LegacyJobName(ar.jobName)})
	if err != nil {
		return errors.Annotate(err, "failed getting job details").Err()
	}

	if err := ar.doDownloadArtifacts(ctx, a.GetOut(), h, ar.workDir, j); err != nil {
		return errors.Annotate(err, "failed downloading artifacts").Err()
	}
	return nil
}
