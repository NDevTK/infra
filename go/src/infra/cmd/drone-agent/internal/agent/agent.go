// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package agent implements an agent which talks to a drone queen
// service and manages Swarming bots.
package agent

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"go.chromium.org/luci/common/errors"

	"infra/appengine/drone-queen/api"
	"infra/cmd/drone-agent/internal/agent/state"
	"infra/cmd/drone-agent/internal/bot"
	"infra/cmd/drone-agent/internal/draining"
	"infra/libs/otil"
)

// Agent talks to a drone queen service and manages Swarming bots.
// This struct stores the static configuration for the agent.  The
// dynamic state is stored in state.State.
type Agent struct {
	Client api.DroneClient
	// SwarmingURL is the URL of the Swarming instance.  Should be
	// a full URL without the path, e.g. https://host.example.com
	SwarmingURL string
	// WorkingDir is used for Swarming bot working dirs.  It is
	// the caller's responsibility to create this.
	WorkingDir        string
	ReportingInterval time.Duration
	DUTCapacity       int
	// StartBotFunc is used to start Swarming bots.
	// This must be set.
	StartBotFunc func(bot.Config) (bot.Bot, error)

	// logger is used for Agent logging.  If nil, use the log package.
	logger logger
	// wrapStateFunc is called to wrap the agent state.  This is
	// used for instrumenting the state for testing.  If nil, this
	// is a no-op.
	wrapStateFunc func(*state.State) stateInterface
	// hive value of the drone agent.  This is used for DUT/drone affinity.
	// A drone is assigned DUTs with same hive value.
	Hive string
	// botPrefix is used to prefix hostnames for bots.
	BotPrefix string
	// BotResources is the compute resources (CPU, RAM, disk I/O etc.) assigned
	// to each bot.
	BotResources *specs.LinuxResources
}

// logger defines the logging interface used by Agent.
type logger interface {
	Printf(string, ...interface{})
}

// stateInterface is the state interface used by the agent.  The usual
// implementation of the interface is in the state package.
type stateInterface interface {
	UUID() string
	WithExpire(ctx context.Context, t time.Time) context.Context
	SetExpiration(t time.Time)
	AddDUT(dutID string)
	DrainDUT(dutID string)
	TerminateDUT(dutID string)
	DrainAll()
	TerminateAll()
	Wait()
	BlockDUTs()
	ActiveDUTs() []string
}

// Run runs the agent until it is canceled via the context.
func (a *Agent) Run(ctx context.Context) {
	a.log("Agent starting")
	for {
		if draining.IsDraining(ctx) || ctx.Err() != nil {
			a.log("Agent exited")
			return
		}
		if err := a.runOnce(ctx); err != nil {
			a.log("Lost drone assignment: %v", err)
		}
	}
}

// runOnce runs one instance of registering and maintaining a drone
// assignment with the queen.
//
// If the context is canceled, this function terminates quickly and
// gracefully (e.g., like handling a SIGTERM or an abort).  If the
// context is drained, this function terminates slowly and gracefully.
// In either case, this function returns nil.
//
// If the assignment is lost or expired for whatever reason, this
// function returns an error.
func (a *Agent) runOnce(ctx context.Context) (err error) {
	ctx, span := otil.FuncSpan(ctx)
	defer func() { otil.EndSpan(span, err) }()
	a.log("Registering with queen")
	res, err := a.Client.ReportDrone(ctx, a.reportRequest(ctx, ""))
	if err != nil {
		return errors.Annotate(err, "register with queen").Err()
	}
	if s := res.GetStatus(); s != api.ReportDroneResponse_OK {
		// TODO(ayatane): We should handle the potential unknown UUID error specially.
		return errors.Reason("register with queen: got unexpected status %v", s).Err()
	}

	// Set up state.
	uuid := res.GetDroneUuid()
	if uuid == "" {
		return errors.Reason("register with queen: got empty UUID").Err()
	}
	a.log("UUID assigned: %s", uuid)
	s := a.wrapState(state.New(uuid, hook{a: a, uuid: uuid}))

	// Set up expiration context.
	t, err := ptypes.Timestamp(res.GetExpirationTime())
	if err != nil {
		return errors.Annotate(err, "register with queen: read expiration").Err()
	}
	ctx = s.WithExpire(ctx, t)

	// Do normal report update.
	if err := applyUpdateToState(res, s); err != nil {
		return errors.Annotate(err, "register with queen").Err()
	}

	return a.reportLoop(ctx, s)
}

// reportLoop implements the core reporting loop of the agent.
// See also runOnce.
func (a *Agent) reportLoop(ctx context.Context, s stateInterface) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	defer wg.Wait()
	readyToExit := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			a.log("Terminating all DUTs due to expired context")
			s.BlockDUTs()
			s.TerminateAll()
		case <-readyToExit:
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-draining.C(ctx):
			a.log("Draining all DUTs")
			s.BlockDUTs()
			s.DrainAll()
		case <-readyToExit:
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-draining.C(ctx):
		case <-ctx.Done():
		}
		s.BlockDUTs()
		s.Wait()
		close(readyToExit)
	}()

	for {
		select {
		case <-time.After(a.ReportingInterval):
		case <-readyToExit:
			return nil
		}
		a.log("Reporting to queen")
		if err := a.reportDrone(ctx, s); err != nil {
			if _, ok := err.(fatalError); ok {
				a.log("Terminating due to fatal error: %s", err)
				cancel()
				return err
			}
			a.log("Error reporting to queen: %s", err)
		}
	}
}

// reportDrone does one cycle of calling the ReportDrone queen RPC and
// handling the response.
func (a *Agent) reportDrone(ctx context.Context, s stateInterface) (err error) {
	ctx, span := otil.FuncSpan(ctx)
	defer func() { otil.EndSpan(span, err) }()
	res, err := a.Client.ReportDrone(ctx, a.reportRequest(ctx, s.UUID()))
	if err != nil {
		return errors.Annotate(err, "report to queen").Err()
	}
	switch rs := res.GetStatus(); rs {
	case api.ReportDroneResponse_OK:
	case api.ReportDroneResponse_UNKNOWN_UUID:
		s.TerminateAll()
		return fatalError{reason: "queen returned UNKNOWN_UUID"}
	default:
		return errors.Reason("report to queen: got unexpected status %v", rs).Err()
	}
	if err := applyUpdateToState(res, s); err != nil {
		return errors.Annotate(err, "report to queen").Err()
	}
	return nil
}

// applyUpdateToState applies the response from a ReportDrone call to the agent state.
func applyUpdateToState(res *api.ReportDroneResponse, s stateInterface) error {
	t, err := ptypes.Timestamp(res.GetExpirationTime())
	if err != nil {
		return errors.Annotate(err, "apply update to state").Err()
	}
	s.SetExpiration(t)
	draining := make(map[string]bool)
	for _, d := range res.GetDrainingDuts() {
		s.DrainDUT(d)
		draining[d] = true
	}
	assigned := make(map[string]bool)
	for _, d := range res.GetAssignedDuts() {
		assigned[d] = true
		if !draining[d] {
			s.AddDUT(d)
		}
	}
	for _, d := range s.ActiveDUTs() {
		if !assigned[d] {
			s.TerminateDUT(d)
		}
	}
	return nil
}

// reportRequest returns the api.ReportDroneRequest to use when
// reporting to the drone queen.
func (a *Agent) reportRequest(ctx context.Context, uuid string) *api.ReportDroneRequest {
	hostname, err := os.Hostname()
	if err != nil {
		a.log("Error getting drone hostname: %s", err)
	}

	req := api.ReportDroneRequest{
		DroneUuid: uuid,
		LoadIndicators: &api.ReportDroneRequest_LoadIndicators{
			DutCapacity: intToUint32(a.DUTCapacity),
		},
		DroneDescription: hostname,
		Hive:             a.Hive,
	}
	if shouldRefuseNewDUTs(ctx) {
		req.LoadIndicators.DutCapacity = 0
	}
	return &req
}

// shouldRefuseNewDUTs returns true if we should refuse new DUTs.
func shouldRefuseNewDUTs(ctx context.Context) bool {
	return draining.IsDraining(ctx) || ctx.Err() != nil
}

func (a *Agent) log(format string, args ...interface{}) {
	if v := a.logger; v != nil {
		v.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (a *Agent) wrapState(s *state.State) stateInterface {
	if a.wrapStateFunc == nil {
		return s
	}
	return a.wrapStateFunc(s)
}

// hook implements state.ControllerHook.
type hook struct {
	a    *Agent
	uuid string
}

// workingDirPrefixLength determines the number of trailing bytes of a dash-abbreviated DUT name to use
// as a suffix.
//
// A non-positive DUT suffix directs drone agent to take the entire DUT name as the suffix.
// See b:218349208 for details.
//
// The length is currently set to 20 conservatively. Values higher than 37 WILL cause tasks to start failing
// because some generated paths to unix domain sockets within this directory will be too long.
//
// The LUCI libraries used by swarming and bbagent impose a maximum length of 104 bytes on the maximum length of a
// unix domain socket. This limit does not change depending on the operating system, although the underlying length
// of the unix domain socket does.
//
// Current working directories have the following form, where X is a disambiguation suffix and Y is a hostname.
//
//	/home/chromeos-test/skylab_bots/YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY.XXXXXXXXX
//
// The following is a hostname, appended by the swarming bot.
//
//	/w/ir/x/ld/sock.ZZZZZZZZZ
//
// This means that 67 bytes total are used for parts of the path that we do not control, leaving 37 for the hostname.
// All of these details are implementation details though, so let's conservatively pick a lower bound of 20 character,
// which should be sufficient in practice.
const workingDirPrefixLength = 20

// StartBot implements state.ControllerHook.
func (h hook) StartBot(dutID string) (bot.Bot, error) {
	workingDirPrefix := abbreviate(dutID, workingDirPrefixLength)
	dir, err := ioutil.TempDir(h.a.WorkingDir, workingDirPrefix+".")
	if err != nil {
		return nil, errors.Annotate(err, "start bot %v", dutID).Err()
	}
	if err := h.shareCIPDCacheWithBot(dir); err != nil {
		// The bot can run without problem with its own CIPD cache, though it
		// may cause higher I/O.
		h.a.log("Bot %v will use its own CIPD cache: %s", dutID, err)
	}
	b, err := h.a.StartBotFunc(h.botConfig(dutID, dir))
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, errors.Annotate(err, "start bot %v", dutID).Err()
	}
	return b, nil
}

// shareCIPDCacheWithBot try to setup a common CIPD cache directory on the
// agent level and share with all bots for better caching.
// We create a common cache dir and symlink to each bot's CIPD cache dir.
// We cannot use the common dir to replace the whole {BotDir}/cipd_cache dir
// since Swarming bots may remove/recreate files in subdirectories like
// {BotDir}/cipd_cache/bin. Thus we can only symlink the common cache dir to
// {BotDir}/cipd_cache/cache.
func (h hook) shareCIPDCacheWithBot(botDir string) error {
	agentCIPDCache := filepath.Join(h.a.WorkingDir, "cipd_cache")
	botCIPDCache := filepath.Join(botDir, "cipd_cache")
	if err := os.MkdirAll(agentCIPDCache, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("setup bot CIPD cache: cannot create common CIPD cache dir %q: %s", agentCIPDCache, err)
	}
	if err := os.MkdirAll(botCIPDCache, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("setup bot CIPD cache: cannot create bot CIPD cache dir %q: %s", botCIPDCache, err)
	}
	cacheDir := filepath.Join(botCIPDCache, "cache")
	if err := os.Symlink(agentCIPDCache, cacheDir); err != nil {
		return fmt.Errorf("setup bot CIPD cache %q: %s", cacheDir, err)
	}
	return nil
}

// botConfig returns a bot config for starting a Swarming bot.
func (h hook) botConfig(dutID string, workDir string) bot.Config {
	botID := h.a.BotPrefix + dutID
	return bot.Config{
		SwarmingURL:   h.a.SwarmingURL,
		BotID:         botID,
		WorkDirectory: workDir,
		Resources:     h.a.BotResources,
	}
}

// ReleaseDUT implements state.ControllerHook.
func (h hook) ReleaseDUT(dutID string) {
	const releaseDUTsTimeout = time.Minute
	ctx := context.Background()
	ctx, f := context.WithTimeout(ctx, releaseDUTsTimeout)
	defer f()
	req := api.ReleaseDutsRequest{
		DroneUuid: h.uuid,
		Duts:      []string{dutID},
	}
	// Releasing DUTs is best-effort.  Ignore any errors since
	// there's no way to handle them.
	//
	h.a.log("Releasing %s", dutID)
	_, _ = h.a.Client.ReleaseDuts(ctx, &req)
}

// truncate returns a suffix of a string of length at most n.
// truncate returns the entire string if given a non-positive value for n.
func truncate(str string, n int) string {
	if n <= 0 {
		return str
	}
	stop := len(str)
	if n < stop {
		stop = n
	}
	return str[:stop]
}

// NumericalSuffixes include numbers like 12 and pseudo-numbers like 14a.
var numericalSuffix = regexp.MustCompile(`[0-9]+[a-z]?\z`)

// abbreviateWord takes a word and unconditionally takes the first character and takes any
// numeric suffix.
//
// E.g. "chromeos4478" --> "c4478"
func abbreviateWord(word string) string {
	if len(word) == 0 {
		return ""
	}
	first := word[0]
	suffix := numericalSuffix.FindString(word)
	// The suffix and first character might overlap.
	if len(suffix) == len(word) {
		return word
	}
	return string(first) + suffix
}

// abbreviate takes a hostname that is dash-delimited and abbreviates each dash-delimited word.
// If the hostname contains no dashes at all, we abbreviate the raw hostname.
// dashAbbrev will never return a string longer than n.
//
// E.g. abbreviate("abc123-def456", ...) === "a123-d456"
//
//	abbreviate("aaaaaaaa", ...) === "aaaaaaaa"
//
// See b:218349208 for details. Hostnames that are too long cause the generated names of unix domain sockets
// to exceed the current 104-byte limit.
func abbreviate(str string, n int) string {
	if !strings.Contains(str, "-") {
		return truncate(str, n)
	}
	words := strings.Split(str, "-")
	for i := range words {
		words[i] = abbreviateWord(words[i])
	}
	out := strings.Join(words, "-")
	return truncate(out, n)
}

// fatalError indicates that the agent should terminate its current
// UUID assignment session and re-register with the queen.
type fatalError struct {
	reason string
}

func (e fatalError) Error() string {
	return fmt.Sprintf("agent fatal error: %s", e.reason)
}

// intToUint32 converts an int to a uint32.
// If the value is negative, return 0.
// If the value overflows, return the max value.
func intToUint32(a int) uint32 {
	if a < 0 {
		return 0
	}
	if a > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(a)
}
