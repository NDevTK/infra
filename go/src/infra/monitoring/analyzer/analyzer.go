// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analyzer

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"infra/monitoring/client"
	"infra/monitoring/messages"
)

const (
	// StaleMasterThreshold is the maximum number of seconds elapsed before a master
	// triggers a "Stale Master Data" alert.
	StaleMasterThreshold = 10 * time.Minute
)

var (
	log = logrus.New()
	// Aliased for testing purposes.
	now = func() time.Time {
		return time.Now()
	}
)

// Analyzer runs the process of checking masters, builders, test results and so on,
// in order to produce alerts.
type Analyzer struct {
	// MaxRecentBuilds is the maximum number of recent builds to check, per builder.
	MaxRecentBuilds int
	// client is the Client implementation for fetching json from CBE, builds, etc.
	Client client.Client

	// bCache is a map of build cache key to Build message.
	bCache map[string]*messages.Builds
}

// New returns a new Analyzer. If client is nil, it assigns a default implementation.
// maxBuilds is the maximum number of builds to check, per builder.
func New(c client.Client, maxBuilds int) *Analyzer {
	if c == nil {
		c = client.New()
	}

	return &Analyzer{
		Client:          c,
		MaxRecentBuilds: maxBuilds,
		bCache:          map[string]*messages.Builds{},
	}
}

// MasterAlerts returns alerts generated from the master at URL.
func (a *Analyzer) MasterAlerts(url string, be *messages.BuildExtract) []messages.Alert {
	ret := []messages.Alert{}

	// Copied logic from builder_messages.
	// No created_timestamp should be a warning sign, no?
	if be.CreatedTimestamp == messages.EpochTime(0) {
		return ret
	}

	elapsed := now().Sub(be.CreatedTimestamp.Time())
	if elapsed > StaleMasterThreshold {
		ret = append(ret, messages.Alert{
			Key:      fmt.Sprintf("stale master: %v", url),
			Title:    "Stale Master Data",
			Body:     fmt.Sprintf("%s elapsed since last update.", elapsed),
			Severity: 0,
			Time:     messages.TimeToEpochTime(now()),
			Links:    []messages.Link{{"Master Url", url}},
			// No type or extension for now.
		})
	}
	if elapsed < 0 {
		// Add this to the alerts returned, rather than just log it?
		log.Errorf("Master %s timestamp is newer than current time (%s): %s old.", url, now(), elapsed)
	}

	return ret
}

// BuilderAlerts returns alerts generated from builders connected to the master at url.
func (a *Analyzer) BuilderAlerts(url string, be *messages.BuildExtract) []messages.Alert {
	mn, err := masterName(url)
	if err != nil {
		log.Fatalf("Couldn't parse %s: %s", url, err)
	}

	// TODO: Collect activeBuilds from be.Slaves.RunningBuilds
	type r struct {
		bn     string
		b      messages.Builders
		alerts []messages.Alert
		err    []error
	}
	c := make(chan r, len(be.Builders))

	// TODO: get a list of all the running builds from be.Slaves? It
	// appears to be used later on in the original py.
	for bn, b := range be.Builders {
		go func(bn string, b messages.Builders) {
			out := r{bn: bn, b: b}
			defer func() {
				c <- out
			}()

			// This blocks on IO, hence the goroutine.
			a.warmBuildCache(mn, bn, b.CachedBuilds)

			// Each call to builderAlerts may trigger blocking json fetches,
			// but it has a data dependency on the above cache-warming call, so
			// the logic remains serial.
			out.alerts, out.err = a.builderAlerts(mn, bn, &b)
		}(bn, b)
	}

	ret := []messages.Alert{}
	for bn := range be.Builders {
		r := <-c
		if len(r.err) != 0 {
			// TODO: add a special alert for this too?
			log.Errorf("Error getting alerts for builder %s: %v", bn, r.err)
		} else {
			ret = append(ret, r.alerts...)
		}
	}

	return ret
}

// masterName extracts the name of the master from the master's URL.
func masterName(URL string) (string, error) {
	mURL, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	pathParts := strings.Split(mURL.Path, "/")
	return pathParts[len(pathParts)-1], nil
}

func cacheKeyForBuild(master, builder string, number int64) string {
	return filepath.FromSlash(
		fmt.Sprintf("%s/%s/%d.json", url.QueryEscape(master), url.QueryEscape(builder), number))
}

// TODO: actually write the on-disk cache.
func filenameForCacheKey(cc string) string {
	cc = strings.Replace(cc, "/", "_", -1)
	return fmt.Sprintf("/tmp/dispatcher.cache/%s", cc)
}

func alertKey(master, builder, step, reason string) string {
	return fmt.Sprintf("%s.%s.%s.%s", master, builder, step, reason)
}

func (a *Analyzer) warmBuildCache(master, builder string, recentBuildIDs []int64) {
	v := url.Values{}
	v.Add("master", master)
	v.Add("builder", builder)

	URL := fmt.Sprintf("https://chrome-build-extract.appspot.com/get_builds?%s", v.Encode())
	res := struct {
		Builds []messages.Builds `json:"builds"`
	}{}

	// TODO: add FetchBuilds to the client interface. Take a list of {master, builder} and
	// return (map[{master, builder}][]Builds, map [{master, builder}]error)
	// That way we can do all of these in parallel.
	status, err := a.Client.JSON(URL, &res)
	if err != nil {
		log.Errorf("Error (%d) fetching %s: %s", status, URL, err)
	}

	for _, b := range res.Builds {
		a.bCache[cacheKeyForBuild(master, builder, b.Number)] = &b
	}
}

// This type is used for sorting build IDs.
type buildIDs []int64

func (a buildIDs) Len() int           { return len(a) }
func (a buildIDs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a buildIDs) Less(i, j int) bool { return a[i] > a[j] }

// TODO: also check the build slaves to see if there are alerts for currently running builds that
// haven't shown up in CBE yet.
func (a *Analyzer) builderAlerts(mn string, bn string, b *messages.Builders) ([]messages.Alert, []error) {
	recentBuildIDs := b.CachedBuilds
	// Should be a *reverse* sort.
	sort.Sort(buildIDs(recentBuildIDs))
	if len(recentBuildIDs) > a.MaxRecentBuilds {
		recentBuildIDs = recentBuildIDs[:a.MaxRecentBuilds]
	}

	log.Infof("Checking %d most recent builds for %s/%s", len(recentBuildIDs), mn, bn)

	alerts := []messages.Alert{}
	errs := []error{}

	// Check for alertable step failures.
	for _, buildID := range recentBuildIDs {
		failures, err := a.stepFailures(mn, bn, buildID)
		if err != nil {
			errs = append(errs, err)
		}
		as, err := a.stepFailureAlerts(failures)
		if err != nil {
			errs = append(errs, err)
		}
		alerts = append(alerts, as...)
	}

	return alerts, errs
}

// stepFailures returns the steps that have failed recently on builder bn.
func (a *Analyzer) stepFailures(mn string, bn string, bID int64) ([]stepFailure, error) {
	cc := cacheKeyForBuild(mn, bn, bID)

	var err error // To avoid re-scoping b in the nested conditional below with a :=.
	b, ok := a.bCache[cc]
	if !ok {
		log.Infof("Cache miss for %s", cc)
		b, err = a.Client.Build(mn, bn, bID)
		if err != nil || b == nil {
			log.Errorf("Error fetching build: %v", err)
			return nil, err
		}
	}

	ret := []stepFailure{}
	for _, s := range b.Steps {
		if !s.IsFinished || len(s.Results) == 0 {
			continue
		}
		// Because Results in the json data is a homogeneous array, the unmarshaler
		// doesn't have any type information to assert about it. We have to do
		// some ugly runtime type assertion ourselves.
		if r, ok := s.Results[0].(float64); ok {
			if r == 0 || r == 1 {
				// This 0/1 check seems to be a convention or heuristic. A 0 or 1
				// result is apparently "ok", accoring to the original python code.
				continue
			}
		} else {
			log.Errorf("Couldn't unmarshal first step result into an int: %v", s.Results[0])
		}

		// We have a failure of some kind, so queue it up to check later.
		ret = append(ret, stepFailure{
			masterName:  mn,
			builderName: bn,
			build:       *b,
			step:        s,
		})
	}

	return ret, nil
}

// stepFailureAlerts returns alerts generated from step failures. It applies filtering
// logic specified in the gatekeeper config to ignore some failures.
func (a *Analyzer) stepFailureAlerts(failures []stepFailure) ([]messages.Alert, error) {
	ret := []messages.Alert{}
	type res struct {
		f   stepFailure
		a   *messages.Alert
		err error
	}

	// Might not need full capacity buffer, since some failures are ignored below.
	rs := make(chan res, len(failures))

	for _, f := range failures {
		// goroutine/channel because the reasonsForFailure call potentially
		// blocks on IO.
		go func(f stepFailure) {
			alr := messages.Alert{
				Title: "Builder step failure",
				Time:  messages.EpochTime(now().Unix()),
				Type:  "buildfailure",
			}

			bf := messages.BuildFailure{
				// FIXME: group builders?
				Builders: []messages.AlertedBuilder{
					{
						Name:          f.builderName,
						URL:           f.URL(),
						FirstFailure:  0,
						LatestFailure: 1,
					},
				},
				// TODO: RegressionRanges:
			}

			reasons := a.reasonsForFailure(f)
			for _, r := range reasons {
				bf.Reasons = append(bf.Reasons, messages.Reason{
					TestName: r,
					Step:     f.step.Name,
				})
			}

			alr.Extension = bf
			if len(bf.Reasons) == 0 {
				log.Warnf("No reasons for step failure: %s", alertKey(f.masterName, f.builderName, f.step.Name, ""))
				rs <- res{
					f: f,
				}
			} else {
				// Should the key include all of the reasons?
				alr.Key = alertKey(f.masterName, f.builderName, f.step.Name, reasons[0])

				rs <- res{
					f:   f,
					a:   &alr,
					err: nil,
				}
			}
		}(f)
	}

	for _ = range failures {
		r := <-rs
		if r.a != nil {
			ret = append(ret, *r.a)
		}
	}

	return ret, nil
}

// reasonsForFailure examines the step failure and applies some heuristics to
// to find the cause. It may make blocking IO calls in the process.
func (a *Analyzer) reasonsForFailure(f stepFailure) []string {
	ret := []string{}
	log.Infof("Checking for reasons for failure step: %v", f.step.Name)
	switch {
	case f.step.Name == "compile":
		log.Errorf("CompileSplitter")
		// CompileSplitter
	case f.step.Name == "webkit_tests":
		log.Errorf("LayoutSplitter")
		// LayoutTestSplitter
	case f.step.Name == "androidwebview_instrumentation_tests" || f.step.Name == "mojotest_instrumentation_tests":
		log.Errorf("JUnitSplitter")
		// JUnitSplitter
	case strings.HasSuffix(f.step.Name, "tests"):
		// GTestSplitter
		testResults, err := a.Client.TestResults(f.masterName, f.builderName, f.step.Name, f.build.Number)
		if err != nil {
			log.Errorf("Error fetching test results")
			return nil
		}
		if len(testResults.Tests) == 0 {
			log.Errorf("No test results for %v", f)
		}
		log.Infof("%d test results", len(testResults.Tests))

		for testName, testResults := range testResults.Tests {
			// This string splitting logic was copied from builder_alerts.
			// I'm not sure it's really necessary since the strings appear to always
			// be either "PASS" or "FAIL" in practice.
			expected := strings.Split(testResults.Expected, " ")
			actual := strings.Split(testResults.Actual, " ")
			ue := unexpected(expected, actual)
			if len(ue) > 0 {
				ret = append(ret, testName)
			}
		}

	default:
		log.Errorf("Unknown step type, unable to find reasons: %s", f.step.Name)
		return nil
	}

	return ret
}

// unexpected returns the set of expected xor actual.
func unexpected(expected, actual []string) []string {
	e, a := make(map[string]bool), make(map[string]bool)
	for _, s := range expected {
		e[s] = true
	}
	for _, s := range actual {
		a[s] = true
	}

	ret := []string{}
	for k := range e {
		if !a[k] {
			ret = append(ret, k)
		}
	}

	for k := range a {
		if !e[k] {
			ret = append(ret, k)
		}
	}

	return ret
}

type stepFailure struct {
	masterName  string
	builderName string
	build       messages.Builds
	step        messages.Steps
}

// Sigh.  build.chromium.org doesn't accept + as an escaped space in URL paths.
func oldEscape(s string) string {
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}

// URL returns a url to builder step failure page.
func (f stepFailure) URL() string {
	return fmt.Sprintf("https://build.chromium.org/p/%s/builders/%s/builds/%d/steps/%s",
		f.masterName, oldEscape(f.builderName), f.build.Number, oldEscape(f.step.Name))
}
