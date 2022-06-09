// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/data/caching/lru"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/config"
	"go.chromium.org/luci/config/cfgclient"
	"go.chromium.org/luci/config/validation"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/gae/service/info"
	"go.chromium.org/luci/server/caching"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	configpb "infra/appengine/weetbix/proto/config"
)

// LRU cache, of which only one slot is used (config for all projects
// is stored in the same slot). We use LRU cache instead of cache slot
// as we sometimes want to refresh config before it has expired.
// Only the LRU Cache has the methods to do this.
var projectsCache = caching.RegisterLRUCache(1)

const projectConfigKind = "weetbix.ProjectConfig"

// ProjectCacheExpiry defines how often project configuration stored
// in the in-process cache is refreshed from datastore.
const ProjectCacheExpiry = 1 * time.Minute

// StartingEpoch is the earliest valid config version for a project.
// It is deliberately different from the timestamp zero value to be
// discernible from "timestamp not populated" programming errors.
var StartingEpoch = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

// NotExistsErr is returned if no matching configuration could be found
// for the specified project.
var NotExistsErr = errors.New("no config exists for the specified project")

var (
	importAttemptCounter = metric.NewCounter(
		"weetbix/project_config/import_attempt",
		"The number of import attempts of project config",
		nil,
		// status can be "success" or "failure".
		field.String("project"), field.String("status"))
)

type cachedProjectConfig struct {
	_extra datastore.PropertyMap `gae:"-,extra"`
	_kind  string                `gae:"$kind,weetbix.ProjectConfig"`

	ID     string      `gae:"$id"` // The name of the project for which the config is.
	Config []byte      `gae:",noindex"`
	Meta   config.Meta `gae:",noindex"`
}

func init() {
	// Registers validation of the given configuration paths with cfgmodule.
	validation.Rules.Add("regex:projects/.*", "${appid}.cfg", func(ctx *validation.Context, configSet, path string, content []byte) error {
		// Discard the returned deserialized message.
		validateProjectConfigRaw(ctx, string(content))
		return nil
	})
}

// updateProjects fetches fresh project-level configuration from LUCI Config
// service and stores it in datastore.
func updateProjects(ctx context.Context) error {
	// Fetch freshest configs from the LUCI Config.
	fetchedConfigs, err := fetchLatestProjectConfigs(ctx)
	if err != nil {
		return err
	}

	var errs []error
	parsedConfigs := make(map[string]*fetchedProjectConfig)
	for project, fetch := range fetchedConfigs {
		valCtx := validation.Context{Context: ctx}
		valCtx.SetFile(fetch.Path)
		msg := validateProjectConfigRaw(&valCtx, fetch.Content)
		if err := valCtx.Finalize(); err != nil {
			blocking := err.(*validation.Error).WithSeverity(validation.Blocking)
			if blocking != nil {
				// Continue through validation errors to ensure a validation
				// error in one project does not affect other projects.
				errs = append(errs, errors.Annotate(blocking, "validation errors for %q", project).Err())
				msg = nil
			}
		}
		// We create an entry even for invalid config (where msg == nil),
		// because we want to signal that config for this project still exists
		// and existing config should be retained instead of being deleted.
		parsedConfigs[project] = &fetchedProjectConfig{
			Config: msg,
			Meta:   fetch.Meta,
		}
	}
	forceUpdate := false
	success := true
	if err := updateStoredConfig(ctx, parsedConfigs, forceUpdate); err != nil {
		errs = append(errs, err)
		success = false
	}
	// Report success for all projects that passed validation, assuming the
	// update succeeded.
	for project, config := range parsedConfigs {
		status := "success"
		if !success || config.Config == nil {
			status = "failure"
		}
		importAttemptCounter.Add(ctx, 1, project, status)
	}

	if len(errs) > 0 {
		return errors.NewMultiError(errs...)
	}
	return nil
}

type fetchedProjectConfig struct {
	// config is the project-level configuration, if it has passed validation,
	// and nil otherwise.
	Config *configpb.ProjectConfig
	// meta is populated with config metadata.
	Meta config.Meta
}

// updateStoredConfig updates the config stored in datastore. fetchedConfigs
// contains the new configs to store, setForTesting forces overwrite of existing
// configuration (ignoring whether the config revision is newer).
func updateStoredConfig(ctx context.Context, fetchedConfigs map[string]*fetchedProjectConfig, setForTesting bool) error {
	// Drop out of any existing datastore transactions.
	ctx = cleanContext(ctx)

	currentConfigs, err := fetchProjectConfigEntities(ctx)
	if err != nil {
		return err
	}

	var errs []error
	var toPut []*cachedProjectConfig
	for project, fetch := range fetchedConfigs {
		if fetch.Config == nil {
			// Config did not pass validation.
			continue
		}
		cur, ok := currentConfigs[project]
		if !ok {
			cur = &cachedProjectConfig{
				ID: project,
			}
		}
		if !setForTesting && cur.Meta.Revision == fetch.Meta.Revision {
			logging.Infof(ctx, "Cached config %s is up-to-date at rev %q", cur.ID, cur.Meta.Revision)
			continue
		}
		configToSave := proto.Clone(fetch.Config).(*configpb.ProjectConfig)
		if !setForTesting {
			var lastUpdated time.Time
			if cur.Config != nil {
				cfg := &configpb.ProjectConfig{}
				if err := proto.Unmarshal(cur.Config, cfg); err != nil {
					// Continue through errors to ensure bad config for one project
					// does not affect others.
					errs = append(errs, errors.Annotate(err, "unmarshal current config").Err())
					continue
				}
				lastUpdated = cfg.LastUpdated.AsTime()
			}
			// ContentHash updated implies Revision updated, but Revision updated
			// does not imply ContentHash updated. To avoid unnecessarily
			// incrementing the last updated time (which triggers re-clustering),
			// only update it if content has changed.
			if cur.Meta.ContentHash != fetch.Meta.ContentHash {
				// Content updated. Update version.
				now := clock.Now(ctx)
				if !now.After(lastUpdated) {
					errs = append(errs, errors.New("old config version is after current time"))
					continue
				}
				lastUpdated = now
			}
			configToSave.LastUpdated = timestamppb.New(lastUpdated)
		}
		// else: use LastUpdated time provided in SetTestProjectConfig call.

		blob, err := proto.Marshal(configToSave)
		if err != nil {
			// Continue through errors to ensure bad config for one project
			// does not affect others.
			errs = append(errs, errors.Annotate(err, "marshal fetched config").Err())
			continue
		}
		logging.Infof(ctx, "Updating cached config %s: %q -> %q", cur.ID, cur.Meta.Revision, fetch.Meta.Revision)
		toPut = append(toPut, &cachedProjectConfig{
			ID:     cur.ID,
			Config: blob,
			Meta:   fetch.Meta,
		})
	}
	if err := datastore.Put(ctx, toPut); err != nil {
		errs = append(errs, errors.Annotate(err, "updating project configs").Err())
	}

	var toDelete []*datastore.Key
	for project, cur := range currentConfigs {
		if _, ok := fetchedConfigs[project]; ok {
			continue
		}
		toDelete = append(toDelete, datastore.KeyForObj(ctx, cur))
	}

	if err := datastore.Delete(ctx, toDelete); err != nil {
		errs = append(errs, errors.Annotate(err, "deleting stale project configs").Err())
	}

	if len(errs) > 0 {
		return errors.NewMultiError(errs...)
	}
	return nil
}

func fetchLatestProjectConfigs(ctx context.Context) (map[string]config.Config, error) {
	configs, err := cfgclient.Client(ctx).GetProjectConfigs(ctx, "${appid}.cfg", false)
	if err != nil {
		return nil, err
	}
	result := make(map[string]config.Config)
	for _, cfg := range configs {
		project := cfg.ConfigSet.Project()
		if project != "" {
			result[project] = cfg
		}
	}
	return result, nil
}

// fetchProjectConfigEntities retrieves project configuration entities
// from datastore, including metadata.
func fetchProjectConfigEntities(ctx context.Context) (map[string]*cachedProjectConfig, error) {
	var configs []*cachedProjectConfig
	err := datastore.GetAll(ctx, datastore.NewQuery(projectConfigKind), &configs)
	if err != nil {
		return nil, errors.Annotate(err, "fetching project configs from datastore").Err()
	}
	result := make(map[string]*cachedProjectConfig)
	for _, cfg := range configs {
		result[cfg.ID] = cfg
	}
	return result, nil
}

// projectsWithMinimumVersion retrieves projects configurations, with
// the specified project at at least the specified minimumVersion.
// If no particular minimum version is desired, specify a project of ""
// or a minimumVersion of time.Time{}.
func projectsWithMinimumVersion(ctx context.Context, project string, minimumVersion time.Time) (map[string]*configpb.ProjectConfig, error) {
	var pc map[string]*configpb.ProjectConfig
	var err error
	cache := projectsCache.LRU(ctx)
	if cache == nil {
		// A fallback useful in unit tests that may not have the process cache
		// available. Production environments usually have the cache installed
		// by the framework code that initializes the root context.
		pc, err = fetchProjects(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		value, _ := projectsCache.LRU(ctx).Mutate(ctx, "projects", func(it *lru.Item) *lru.Item {
			var pc map[string]*configpb.ProjectConfig
			if it != nil {
				pc = it.Value.(map[string]*configpb.ProjectConfig)
				projectCfg, ok := pc[project]
				if project == "" || (ok && !projectCfg.LastUpdated.AsTime().Before(minimumVersion)) {
					// Projects contains the specified project at the given minimum version.
					// There is no need to update it.
					return it
				}
			}
			if pc, err = fetchProjects(ctx); err != nil {
				// Error refreshing config. Keep existing entry (if any).
				return it
			}
			return &lru.Item{
				Value: pc,
				Exp:   ProjectCacheExpiry,
			}
		})
		if err != nil {
			return nil, err
		}
		pc = value.(map[string]*configpb.ProjectConfig)
	}

	projectCfg, ok := pc[project]
	if project != "" && (ok && projectCfg.LastUpdated.AsTime().Before(minimumVersion)) {
		return nil, fmt.Errorf("could not obtain projects configuration with project %s at minimum version (%v)", project, minimumVersion)
	}
	return pc, nil
}

// Projects returns all project configurations, in a map by project name.
// Uses in-memory cache to avoid hitting datastore all the time.
func Projects(ctx context.Context) (map[string]*configpb.ProjectConfig, error) {
	return projectsWithMinimumVersion(ctx, "", time.Time{})
}

// fetchProjects retrieves all project configurations from datastore.
func fetchProjects(ctx context.Context) (map[string]*configpb.ProjectConfig, error) {
	ctx = cleanContext(ctx)

	cachedCfgs, err := fetchProjectConfigEntities(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "fetching cached config").Err()
	}
	result := make(map[string]*configpb.ProjectConfig)
	for project, cached := range cachedCfgs {
		cfg := &configpb.ProjectConfig{}
		if err := proto.Unmarshal(cached.Config, cfg); err != nil {
			return nil, errors.Annotate(err, "unmarshalling cached config").Err()
		}
		result[project] = cfg
	}
	return result, nil
}

// cleanContext returns a context with datastore using the default namespace
// and not using transactions.
func cleanContext(ctx context.Context) context.Context {
	return datastore.WithoutTransaction(info.MustNamespace(ctx, ""))
}

// SetTestProjectConfig sets test project configuration in datastore.
// It should be used from unit/integration tests only.
func SetTestProjectConfig(ctx context.Context, cfg map[string]*configpb.ProjectConfig) error {
	fetchedConfigs := make(map[string]*fetchedProjectConfig)
	for project, pcfg := range cfg {
		fetchedConfigs[project] = &fetchedProjectConfig{
			Config: pcfg,
			Meta:   config.Meta{},
		}
	}
	setForTesting := true
	if err := updateStoredConfig(ctx, fetchedConfigs, setForTesting); err != nil {
		return err
	}
	testable := datastore.GetTestable(ctx)
	if testable == nil {
		return errors.New("SetTestProjectConfig should only be used with testable datastore implementations")
	}
	// An up-to-date index is required for fetch to retrieve the project
	// entities we just saved.
	testable.CatchupIndexes()
	return nil
}

// Project returns the configuration of the requested project.
func Project(ctx context.Context, project string) (*configpb.ProjectConfig, error) {
	return ProjectWithMinimumVersion(ctx, project, time.Time{})
}

// ProjectWithMinimumVersion returns the configuration of the requested
// project, which has a LastUpdated time of at least minimumVersion.
// This bypasses the in-process cache if the cached version is older
// than the specified version.
func ProjectWithMinimumVersion(ctx context.Context, project string, minimumVersion time.Time) (*configpb.ProjectConfig, error) {
	configs, err := projectsWithMinimumVersion(ctx, project, minimumVersion)
	if err != nil {
		return nil, err
	}
	if c, ok := configs[project]; ok {
		return c, nil
	}
	return nil, NotExistsErr
}
