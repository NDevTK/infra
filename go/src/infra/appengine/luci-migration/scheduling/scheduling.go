// Copyright 2017 The LUCI Authors.
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

// Package scheduling schedules Buildbot builds on LUCI.
//
// It accepts PubSub messages published by Buildbucket for each build.
// If it is a Buildbot build for a builder that we want to migrate,
// schedules an equivalent LUCI build, for a percentage of CLs.
// If it is a failed LUCI build, retries at most 2 times.
//
// Other packages of luci-migration app must not depend on the fact that this
// functionality is implemented in this app.
package scheduling

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/api/googleapi"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/gae/service/memcache"
	"go.chromium.org/luci/buildbucket"
	"go.chromium.org/luci/buildbucket/proto"
	bbapi "go.chromium.org/luci/common/api/buildbucket/buildbucket/v1"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry/transient"

	"infra/appengine/luci-migration/config"
	"infra/appengine/luci-migration/storage"
)

const (
	attemptTagKey         = "luci_migration_attempt"
	buildbotBuildIDTagKey = "luci_migration_buildbot_build_id"
)

// Build is buildbucket.Build plus
// original ParametersJSON.
type Build struct {
	buildbucket.Build
	ParametersJSON string // original ParametersJSON
}

// OutputProperties is used for parsing Build.Output.Properties.
type OutputProperties struct {
	GotRevision string      `json:"got_revision"`
	DryRun      interface{} `json:"dry_run"`
}

// Scheduler schedules Buildbot builds on LUCI.
type Scheduler struct {
	Buildbucket *bbapi.Service
}

// BuildCompleted handles a build completion notification.
func (h *Scheduler) BuildCompleted(c context.Context, build *Build) error {
	switch {
	case build.Status.Ended() && strings.HasPrefix(build.Bucket, "master."):
		return h.buildbotBuildCompleted(c, build)

	case (build.Status == buildbucketpb.Status_FAILURE || build.Status == buildbucketpb.Status_INFRA_FAILURE) &&
		strings.HasPrefix(build.Bucket, "luci."):

		// Note that result==FAILURE builds include infra-failed builds.
		return h.luciBuildFailed(c, build)

	default:
		return nil
	}
}

// buildbotBuildCompleted schedules a Buildbot build on LUCI if the build
// is for a builder that we track. Honors builder's experiment percentage and
// build creation rate limit.
func (h *Scheduler) buildbotBuildCompleted(c context.Context, build *Build) error {
	// Is it a builder that we care about?
	builder := &storage.Builder{ID: storage.BuilderID{
		Master:  strings.TrimPrefix(build.Bucket, "master."),
		Builder: build.Builder,
	}}
	switch err := datastore.Get(c, builder); {
	case err == datastore.ErrNoSuchEntity:
		return nil // no, we don't care about it
	case err != nil:
		return errors.Annotate(err, "could not read builder %s", &builder.ID).Err()
	}
	if builder.SchedulingType != config.SchedulingType_TRYJOBS {
		// we don't support non-tryjob yet
		return nil
	}
	master := config.Get(c).FindMaster(builder.ID.Master)
	if master == nil {
		return errors.Reason("master %q is not configured", builder.ID.Master).Err()
	}

	// Should we experiment with this CL?
	if !shouldExperiment(build.Tags.Get(bbapi.TagBuildSet), builder.ExperimentPercentage) {
		return nil
	}

	// This build should be scheduled on LUCI.
	// Prepare new build request.
	outProps := (build.Output.Properties).(*OutputProperties)
	props := map[string]interface{}{
		"revision": outProps.GotRevision,
		// Mark the build as experimental, so it does not confuse users of Rietveld and Gerrit.
		"category": "cq_experimental",
	}
	if outProps.DryRun != nil {
		props["dry_run"] = outProps.DryRun
	}
	newParamsJSON, err := setProps(build.ParametersJSON, props)
	if err != nil {
		return err
	}
	newTags := strpair.Map{}
	newTags.Set(buildbotBuildIDTagKey, strconv.FormatInt(build.ID, 10))
	newTags.Set(attemptTagKey, "0")
	newTags[bbapi.TagBuildSet] = build.Tags[bbapi.TagBuildSet]
	newBuild := &bbapi.ApiPutRequestMessage{
		Bucket:            master.LuciBucket,
		ClientOperationId: "luci-migration-retry-" + strconv.FormatInt(build.ID, 10),
		ParametersJson:    newParamsJSON,
		Tags:              newTags.Format(),
	}
	return withLock(c, build.ID, func() error {
		logging.Infof(
			c,
			"scheduling Buildbot build %d on LUCI for builder %q and buildset %q",
			build.ID, &builder.ID, build.Tags.Get(bbapi.TagBuildSet))
		return h.schedule(c, builder.ID.Builder, newBuild)
	})
}

func (h *Scheduler) luciBuildFailed(c context.Context, build *Build) error {
	attempt, attemptErr := strconv.Atoi(build.Tags.Get(attemptTagKey))
	buildSet := build.Tags.Get(bbapi.TagBuildSet)
	buildbotBuildID := build.Tags.Get(buildbotBuildIDTagKey)
	switch {
	case attemptErr != nil || buildSet == "" || buildbotBuildID == "":
		return nil // we don't recognize this build

	// Do at most 3 attempts.
	case attempt >= 2:
		logging.Infof(c, "enough retries for build %s", buildbotBuildID)
		return nil
	}

	// Before retrying the build, see if there is a newer one already created.
	// It may happen if a new Buildbot build completed.
	req := h.Buildbucket.Search()
	req.Context(c)
	req.Bucket(build.Bucket)
	req.CreationTsLow(bbapi.FormatTimestamp(build.CreationTime) + 1)
	req.IncludeExperimental(true)
	req.Tag(
		strpair.Format(bbapi.TagBuildSet, buildSet),
		strpair.Format(bbapi.TagBuilder, build.Builder),
	)
	switch newerBuilds, _, err := req.Fetch(1, nil); {
	case err != nil:
		return errors.Annotate(err, "failed to search newer builds").Err()
	case len(newerBuilds) > 0:
		logging.Infof(c, "not retrying because build %d is newer", newerBuilds[0].Id)
		return nil
	}

	newTags := strpair.Map{}
	newTags.Set(buildbotBuildIDTagKey, buildbotBuildID)
	newTags.Set(attemptTagKey, strconv.Itoa(attempt+1))
	newTags.Set(bbapi.TagBuildSet, buildSet)
	newBuild := &bbapi.ApiPutRequestMessage{
		Bucket:            build.Bucket,
		ClientOperationId: "luci-migration-retry-" + strconv.FormatInt(build.ID, 10),
		ParametersJson:    build.ParametersJSON,
		Tags:              newTags.Format(),
	}

	return withLock(c, build.ID, func() error {
		logging.Infof(c, "retrying LUCI build %d", build.ID)
		return h.schedule(c, build.Builder, newBuild)
	})
}

// withLock locks on the build id and, on success, calls f.
// Does not release the lock if f returns nil.
func withLock(c context.Context, buildID int64, f func() error) error {
	// Lock the build via memcache.
	lock := memcache.NewItem(c, "retry-build-"+strconv.FormatInt(buildID, 10))
	switch err := memcache.Add(c, lock); {
	case err == memcache.ErrNotStored:
		logging.Infof(c, "build %d is locked, letting go", buildID)
		return nil

	case err != nil:
		return errors.Annotate(err, "could not lock on build %d", buildID).Err()
	}

	unlock := true
	defer func() {
		if unlock {
			if err := memcache.Delete(c, lock.Key()); err != nil {
				// too bad. The analysis algorithm will probably ignore this patchset.
				logging.WithError(err).Errorf(c, "could not unlock build")
			}
		}
	}()

	if err := f(); err != nil {
		return err
	}

	unlock = false
	return nil
}

// schedule creates a build and logs a successful result.
func (h *Scheduler) schedule(c context.Context, builder string, req *bbapi.ApiPutRequestMessage) error {
	req.Tags = append(req.Tags, "user_agent:luci-migration")
	req.Experimental = true

	res, err := h.Buildbucket.Put(req).Context(c).Do()
	transientFailure := true
	if err == nil && res.Error != nil {
		err = errors.New(res.Error.Message)
		// A LUCI builder may be not defined for a buildbot builder yet.
		// Skip such builds. When the LUCI builder is defined, builds will
		// start flowing.
		transientFailure = res.Error.Reason != "BUILDER_NOT_FOUND"
	}
	if err != nil {
		if apierr, ok := err.(*googleapi.Error); ok && (apierr.Code == http.StatusForbidden || apierr.Code == http.StatusNotFound) {
			// Retries won't help. Return a non-transient error.
			// The bucket should be configured first, it is OK to skip some
			// builds.
			transientFailure = false
		}

		ann := errors.Annotate(err, "could not schedule a build")
		if transientFailure {
			ann.Tag(transient.Tag) // Cause a retry by returning a transient error
		}
		return ann.Err()
	}

	resJSON, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		panic(errors.Annotate(err, "could not marshal JSON response back to JSON").Err())
	}
	logging.Infof(c, "scheduled new build: %s", resJSON)
	return nil
}

// shouldExperiment deterministically returns true if experiments must be done
// for the buildset.
func shouldExperiment(buildSet string, percentage int) bool {
	switch {
	case percentage <= 0:
		return false
	case percentage >= 100:
		return true
	default:
		aByte := sha256.Sum256([]byte(buildSet))[0]
		return int(aByte)*100 <= percentage*255
	}
}

func setProps(paramsJSON string, values map[string]interface{}) (string, error) {
	var parameters map[string]interface{}
	if err := json.Unmarshal([]byte(paramsJSON), &parameters); err != nil {
		return "", err
	}

	var props map[string]interface{} // a pointer to "properties" in parameters
	if propsRaw, ok := parameters["properties"]; ok {
		props, ok = propsRaw.(map[string]interface{})
		if !ok {
			return "", errors.New("properties is not a JSON object")
		}
	} else {
		props = map[string]interface{}{}
		parameters["properties"] = props
	}

	for k, v := range values {
		props[k] = v
	}

	marshalled, err := json.Marshal(parameters)
	if err != nil {
		return "", err
	}
	return string(marshalled), nil
}
