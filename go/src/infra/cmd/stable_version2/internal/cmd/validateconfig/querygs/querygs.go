// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package querygs

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"unicode"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"

	gslib "infra/cmd/stable_version2/internal/gs"

	labPlatform "go.chromium.org/chromiumos/infra/proto/go/lab_platform"
)

// BoardModel is a combined build target and model. It is used for models that aren't present
// in any metadata.json file that we read.
type BoardModel struct {
	BuildTarget string `json:"build_target"`
	Model       string `json:"model"`
}

// VersionMismatch is a manifestation of a mismatched firmware version.
// For the purposes of analyzing the stable version file, we always assume
// that the CrOS version is correct and that the firmware version is potentially
// out of sync with it, rather than the other way around.
type VersionMismatch struct {
	BuildTarget string `json:"build_target"`
	Model       string `json:"model"`
	Wanted      string `json:"wanted"`
	Got         string `json:"got"`
}

// ValidationResult is a summary of the result of validating a stable version config file.
type ValidationResult struct {
	// Non-Lowercase entries
	NonLowercaseEntries []string `json:"non_lowercase_entries"`
	// MissingBoards are the boards that don't have a metadata file in Google Storage.
	MissingBoards []string `json:"missing_boards"`
	// FailedToLookup are board/model pairs that aren't present in the descriptions fetched from Google Storage.
	FailedToLookup []*BoardModel `json:"failed_to_lookup"`
	// InvalidVersions is the list of entries where the version in the config file does not match Google Storage.
	InvalidVersions []*VersionMismatch `json:"invalid_versions"`
	// InvalidFaftVersions is the list of FAFT entries.
	InvalidFaftVersions []*VersionMismatch `json:"invalid_faft_versions"`
}

// RemoveAllowedDUTs removes DUTs that are exempted from the validation error summary.
// examples include labstations
func (r *ValidationResult) RemoveAllowedDUTs() {
	var newMissingBoards []string
	var newFailedToLookup []*BoardModel
	var newInvalidVersions []*VersionMismatch
	if len(r.MissingBoards) != 0 {
		for _, item := range r.MissingBoards {
			if !missingBoardAllowList[item] {
				newMissingBoards = append(newMissingBoards, item)
			}
		}
	}
	if len(r.FailedToLookup) != 0 {
		for _, item := range r.FailedToLookup {
			if !failedToLookupAllowList[fmt.Sprintf("%s;%s", item.BuildTarget, item.Model)] {
				newFailedToLookup = append(newFailedToLookup, item)
			}
		}
	}
	if len(r.InvalidVersions) != 0 {
		for _, item := range r.InvalidVersions {
			if !invalidVersionAllowList[fmt.Sprintf("%s;%s", item.BuildTarget, item.Model)] {
				newInvalidVersions = append(newInvalidVersions, item)
			}
		}
	}
	r.MissingBoards = newMissingBoards
	r.FailedToLookup = newFailedToLookup
	r.InvalidVersions = newInvalidVersions
}

// AnomalyCount counts the total number of issues found in the results summary.
func (r *ValidationResult) AnomalyCount() int {
	return len(r.MissingBoards) + len(r.FailedToLookup) + len(r.InvalidVersions) + len(r.NonLowercaseEntries) + len(r.InvalidFaftVersions)
}

type downloader func(gsPath gs.Path) ([]byte, error)

type existenceChecker func(gsPath gs.Path) error

type specialBoardEntry struct {
	board   string
	version string
}

// Reader reads metadata.json files from google storage and caches the result.
type Reader struct {
	dld  downloader
	exst existenceChecker
	// buildTarget > version > model > version
	cache *map[string]map[string]map[string]string
	// Special boards are boards with no firmware versions.
	specialBoards []specialBoardEntry
}

// Init creates a new Google Storage Client.
// TODO(gregorynisbet): make it possible to initialize a test gsClient as well
func (r *Reader) Init(ctx context.Context, t http.RoundTripper, unmarshaler jsonpb.Unmarshaler, tempPrefix string) error {
	var gsc gslib.Client
	if err := gsc.Init(ctx, t, unmarshaler); err != nil {
		return fmt.Errorf("Reader::Init: %s", err)
	}
	r.dld = func(remotePath gs.Path) ([]byte, error) {
		dir, err := ioutil.TempDir("", tempPrefix)
		if err != nil {
			return nil, fmt.Errorf("download adapter: making temporary directory: %s", err)
		}
		defer os.RemoveAll(dir)
		localPath := filepath.Join(dir, "metadata.json")
		if err := gsc.Download(remotePath, localPath); err != nil {
			return nil, fmt.Errorf("download adapter: fetching file: %s", err)
		}
		contents, err := ioutil.ReadFile(localPath)
		if err != nil {
			return nil, fmt.Errorf("download adapter: reading local file: %s", err)
		}
		return contents, nil
	}
	r.exst = func(remotePath gs.Path) error {
		// Thoroughly check for existence by downloading a few bytes and discarding
		// the result.
		return gsc.DownloadByteRange(remotePath, os.DevNull, 0, 10)
	}
	return nil
}

// RemoteFileExists checks for the existence and nonzero size of a given path in Google Storage.
func (r *Reader) RemoteFileExists(remotePath gs.Path) error {
	return r.exst(remotePath)
}

// combinedKey combines a board and a model into a single key
// and returns just the board name when the model is "".
func combinedKey(board string, model string) string {
	if model == "" {
		return board
	}
	return fmt.Sprintf("%s;%s", board, model)
}

// ValidateConfig takes a stable version protobuf and attempts to validate every entry.
func (r *Reader) ValidateConfig(ctx context.Context, sv *labPlatform.StableVersions) (*ValidationResult, error) {
	var cfgCrosVersions = make(map[string]string, len(sv.GetCros()))
	var out ValidationResult
	if sv == nil {
		return nil, fmt.Errorf("Reader::ValidateConfig: config file cannot be nil")
	}
	// use the CrOS keys in the sv file to seed the reader.
	for _, item := range sv.GetCros() {
		bt := item.GetKey().GetBuildTarget().GetName()
		model := item.GetKey().GetModelId().GetValue()
		combined := combinedKey(bt, model)
		version := item.GetVersion()
		if !isLowercase(bt) {
			out.NonLowercaseEntries = append(out.NonLowercaseEntries, bt)
			continue
		}
		if !isLowercase(model) {
			out.NonLowercaseEntries = append(out.NonLowercaseEntries, model)
			continue
		}
		if _, err := r.getAllModelsForBuildTarget(ctx, bt, version); err != nil {
			out.MissingBoards = append(out.MissingBoards, bt)
			continue
		}
		cfgCrosVersions[combined] = version
	}
	// Confirm that all faft firmware bundles exist.
	for _, item := range sv.GetFaft() {
		bt := item.GetKey().GetBuildTarget().GetName()
		model := item.GetKey().GetModelId().GetValue()
		if path, err := r.validateFaft(item.GetVersion()); err != nil {
			out.InvalidFaftVersions = append(out.InvalidFaftVersions, &VersionMismatch{bt, model, path, ""})
		}
	}
	return &out, nil
}

// allModels returns a mapping from model names to fimrware versions given a buildTaret and CrOS version.
func (r *Reader) getAllModelsForBuildTarget(ctx context.Context, buildTarget string, version string) (map[string]string, error) {
	if err := r.maybeDownloadFile(ctx, buildTarget, version); err != nil {
		logging.Infof(ctx, "failed to get contents for %q %q", buildTarget, version)
		return nil, fmt.Errorf("all models: downloading: %s", err)
	}
	m, err := getAllModels(r, buildTarget, version)
	if err != nil {
		return nil, fmt.Errorf("all models: reading: %s", err)
	}
	return m, nil
}

// getFirmwareVersion returns the firmware version for a specific model given the buildTarget and CrOS version.
func (r *Reader) getFirmwareVersion(ctx context.Context, buildTarget string, model string, version string) (string, error) {
	if err := r.maybeDownloadFile(ctx, buildTarget, version); err != nil {
		logging.Infof(ctx, "failed to get contents for %q %q %q", buildTarget, version, model)
		return "", fmt.Errorf("FirmwareVersion: %s", err)
	}
	if r.cache == nil {
		return "", fmt.Errorf("getFirmwareVersion: cache cannot be empty")
	}
	if _, ok := (*r.cache)[buildTarget]; !ok {
		// If control makes it here, then maybeDownloadFile should have returned
		// a non-nil error.
		panic(fmt.Sprintf("getFirmwareVersion: buildTarget MUST be present (%s)", buildTarget))
	}
	fwversion := get(r.cache, buildTarget, version, model)
	if fwversion == "" {
		return "", fmt.Errorf("no info for model (%s)", model)
	}
	return fwversion, nil
}

// maybeDownloadFile fetches a metadata.json corresponding to a buildTarget and version if it doesn't already exist in the cache.
func (r *Reader) maybeDownloadFile(ctx context.Context, buildTarget string, crosVersion string) error {
	if r.cache == nil {
		v := make(map[string]map[string]map[string]string)
		r.cache = &v
	}
	if m, _ := getAllModels(r, buildTarget, crosVersion); m != nil {
		return nil
	}
	rawRemotePath := fmt.Sprintf("gs://chromeos-image-archive/%s-release/%s/metadata.json", buildTarget, crosVersion)
	remotePath := gs.Path(rawRemotePath)
	contents, err := (r.dld)(remotePath)
	if err != nil {
		return fmt.Errorf("Reader::maybeDownloadFile: fetching file: %s", err)
	}
	fws, err := gslib.ParseMetadata(contents)
	if err != nil {
		return fmt.Errorf("Reader::maybeDownloadFile: parsing metadata.json: %s", err)
	}
	switch len(fws.FirmwareVersions) {
	case 0:
		logging.Infof(ctx, "no firmware versions for board %q at %q, creating special board entry", buildTarget, rawRemotePath)
		r.specialBoards = append(r.specialBoards, specialBoardEntry{buildTarget, crosVersion})
	default:
		// TODO(gregorynisbet): Consider throwing an error or panicking if we encounter
		// a duplicate when populating the cache.
		for _, fw := range fws.FirmwareVersions {
			inferredBuildTarget := fw.GetKey().GetBuildTarget().GetName()
			inferredModel := fw.GetKey().GetModelId().GetValue()
			fwversion := fw.GetVersion()
			set(r.cache, inferredBuildTarget, crosVersion, inferredModel, fwversion)
		}
	}
	return nil
}

// check if string has entirely lowercase letters
func isLowercase(s string) bool {
	for _, ch := range s {
		if unicode.IsLetter(ch) {
			if unicode.IsUpper(ch) {
				return false
			}
		}
	}
	return true
}

func set(m *map[string]map[string]map[string]string, board string, version string, model string, fwversion string) {
	if m == nil {
		v := make(map[string]map[string]map[string]string)
		m = &v
	}
	if (*m)[board] == nil {
		v := make(map[string]map[string]string)
		(*m)[board] = v
	}
	if (*m)[board][version] == nil {
		v := make(map[string]string)
		(*m)[board][version] = v
	}

	(*m)[board][version][model] = fwversion
}

func get(m *map[string]map[string]map[string]string, board string, version string, model string) string {
	if m == nil {
		v := make(map[string]map[string]map[string]string)
		m = &v
	}
	if (*m)[board] == nil {
		v := make(map[string]map[string]string)
		(*m)[board] = v
	}
	if (*m)[board][version] == nil {
		v := make(map[string]string)
		(*m)[board][version] = v
	}

	return (*m)[board][version][model]
}

func getAllModels(r *Reader, board string, version string) (map[string]string, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	// Special boards have no firmware versions, but it is not an error when this happens.
	for _, entry := range r.specialBoards {
		if entry.board == board {
			return nil, nil
		}
	}
	if r.cache == nil {
		return nil, fmt.Errorf("map is nil")
	}
	if (*r.cache)[board] == nil {
		return nil, fmt.Errorf("board %q not present", board)
	}
	if (*r.cache)[board][version] == nil {
		return nil, fmt.Errorf("board+version submap is nil")
	}
	return (*r.cache)[board][version], nil
}
