// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"time"

	"go.chromium.org/luci/common/data/strpair"
	luciflag "go.chromium.org/luci/common/flag"

	skycmdlib "infra/cmd/skylab/internal/cmd/cmdlib"
	"infra/cmd/skylab/internal/cmd/recipe"
	skyflag "infra/cmd/skylab/internal/flagx"
	"infra/cmdsupport/cmdlib"
)

// createRunCommon encapsulates parameters that are common to
// all of the create-* subcommands.
type createRunCommon struct {
	board           string
	model           string
	pool            string
	image           string
	dimensions      map[string]string
	provisionLabels []string
	priority        int
	timeoutMins     int
	maxRetries      int
	tags            []string
	keyvals         []string
	qsAccount       string
	lacrosPath      string
	buildBucket     bool
	uploadCrashes   bool
}

func (c *createRunCommon) Register(fl *flag.FlagSet) {
	fl.StringVar(&c.image, "image", "",
		`Fully specified image name to run test against,
e.g., reef-canary/R73-11580.0.0.`)
	fl.StringVar(&c.board, "board", "", "Board to run test on.")
	fl.StringVar(&c.model, "model", "", "Model to run test on.")
	// We allow arbitrary dimensions to be passed in via the -dim and/or -dims flags.
	// For maximum flexibility, both flags can take one or more key:val or key=val
	// pairs (separated by ","), and repeated/mixed flags are allowed. To keep the
	// docs intuitive, docstrings describe the more natural arg format for each flag.
	fl.Var(skyflag.Dims(&c.dimensions), "dim", "Single additional scheduling dimension in format key=value or key:value; may be specified multiple times.")
	fl.Var(skyflag.Dims(&c.dimensions), "dims", "List of additional scheduling dimensions in format key1=value1,key2=value2,... or key1:value1,key2:value2,... .")
	fl.StringVar(&c.pool, "pool", "", "Device pool to run test on.")
	fl.Var(luciflag.StringSlice(&c.provisionLabels), "provision-label",
		`Additional provisionable labels to use for the test
(e.g. cheets-version:git_pi-arc/cheets_x86_64).  May be specified
multiple times.  Optional.`)
	fl.IntVar(&c.priority, "priority", skycmdlib.DefaultTaskPriority,
		`Specify the priority of the test [50,255].  A high value means this test
will be executed in a low priority. If the tasks runs in a quotascheduler controlled pool, this value will be ignored.`)
	fl.IntVar(&c.timeoutMins, "timeout-mins", 30, "Task runtime timeout.")
	fl.IntVar(&c.maxRetries, "max-retries", 0,
		`Maximum retries allowed in total for all child tests of this
suite. No retry if it is 0.`)
	fl.Var(luciflag.StringSlice(&c.keyvals), "keyval",
		`Autotest keyval for test. Key may not contain : character. May be
specified multiple times.`)
	fl.StringVar(&c.qsAccount, "qs-account", "", "Quota Scheduler account to use for this task.  Optional.")
	fl.StringVar(&c.lacrosPath, "lacros-gcs-path", "", "GCS path pointing to a lacros artifact.  Optional.")
	fl.Var(luciflag.StringSlice(&c.tags), "tag", "Swarming tag for test; may be specified multiple times.")
	fl.BoolVar(&c.uploadCrashes, "upload-crashes", false,
		`Report crashes from DUTs to pre-configured crash servers.
Used by release automation to report crashes similar to end-users' flow.
Most clients should leave this flag unset.`)
	fl.BoolVar(&c.buildBucket, "bb", true, "Deprecated, do not use.")

	// Deprecated flags that have no effect. To be eventually deleted where
	// possible.
	var unusedUseTestRunner bool
	var unusedEnableSynchronousOffload bool
	fl.BoolVar(&unusedUseTestRunner, "use-test-runner", false, "DEPRECATED. Has no effect.")
	fl.BoolVar(&unusedEnableSynchronousOffload, "enable-synchronous-offload", false, "DEPRECATED. Has no effect.")
}

func (c *createRunCommon) ValidateArgs(fl flag.FlagSet) error {
	if c.board == "" {
		return cmdlib.NewUsageError(fl, "missing -board")
	}
	if c.pool == "" {
		return cmdlib.NewUsageError(fl, "missing -pool")
	}
	if c.image == "" {
		return cmdlib.NewUsageError(fl, "missing -image")
	}
	if c.priority < 50 || c.priority > 255 {
		return cmdlib.NewUsageError(fl, "priority should in [50,255]")
	}
	if !c.buildBucket {
		return cmdlib.NewUsageError(fl, "-bb=False is deprecated")
	}
	return nil
}

func (c *createRunCommon) RecipeArgs(tags []string) (recipe.Args, error) {
	dimSlice := toKeyvalSlice(c.dimensions)
	keyvalMap, err := toKeyvalMap(c.keyvals)
	if err != nil {
		return recipe.Args{}, err
	}

	return recipe.Args{
		Board:                      c.board,
		Image:                      c.image,
		Model:                      c.model,
		FreeformSwarmingDimensions: dimSlice,
		ProvisionLabels:            c.provisionLabels,
		Pool:                       c.pool,
		QuotaAccount:               c.qsAccount,
		Timeout:                    time.Duration(c.timeoutMins) * time.Minute,
		MaxRetries:                 c.maxRetries,
		Keyvals:                    keyvalMap,
		Priority:                   int64(c.priority),
		Tags:                       tags,
		UploadCrashes:              c.uploadCrashes,
		LacrosPath:                 c.lacrosPath,
	}, nil
}

func (c *createRunCommon) BuildTags() []string {
	ts := c.tags
	if c.image != "" {
		ts = append(ts, fmt.Sprintf("build:%s", c.image))
	}
	if c.board != "" {
		ts = append(ts, fmt.Sprintf("label-board:%s", c.board))
	}
	if c.model != "" {
		ts = append(ts, fmt.Sprintf("label-model:%s", c.model))
	}
	if c.pool != "" {
		ts = append(ts, fmt.Sprintf("label-pool:%s", c.pool))
	}
	// Only surface the priority if Quota Account was unset. Note,
	// tags attached here will NOT be processed by CTP. The "real"
	// Quota Account or Priority is set in dimension.
	if c.qsAccount != "" {
		ts = append(ts, fmt.Sprintf("quota_account:%s", c.qsAccount))
	} else {
		ts = append(ts, fmt.Sprintf("priority:%d", c.priority))
	}
	return ts
}

func toKeyvalMap(keyvals []string) (map[string]string, error) {
	m := make(map[string]string, len(keyvals))
	for _, s := range keyvals {
		k, v := strpair.Parse(s)
		if v == "" {
			return nil, fmt.Errorf("malformed keyval with key '%s' has no value", k)
		}
		if _, ok := m[k]; ok {
			return nil, fmt.Errorf("keyval with key %s specified more than once", k)
		}
		m[k] = v
	}
	return m, nil
}

func toKeyvalSlice(keyvals map[string]string) []string {
	var s []string
	for key, val := range keyvals {
		s = append(s, fmt.Sprintf("%s:%s", key, val))
	}
	return s
}

func printScheduledTaskJSON(w io.Writer, name string, ID string, URL string) error {
	t := struct {
		Name string `json:"task_name"`
		ID   string `json:"task_id"`
		URL  string `json:"task_url"`
	}{
		Name: name,
		ID:   ID,
		URL:  URL,
	}
	return json.NewEncoder(w).Encode(t)
}
