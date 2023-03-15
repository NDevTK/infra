// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows

// Command drone-agent is the client that talks to the drone queen
// service to provide Swarming bots for running tasks against test
// devices.  See the README.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"go.opentelemetry.io/otel/sdk/trace"

	"google.golang.org/grpc/metadata"

	"infra/appengine/drone-queen/api"
	"infra/cmd/drone-agent/internal/agent"
	"infra/cmd/drone-agent/internal/bot"
	"infra/cmd/drone-agent/internal/draining"
	"infra/cmd/drone-agent/internal/metrics"
	"infra/cmd/drone-agent/internal/tokman"
	"infra/cmd/drone-agent/internal/tracing"
)

const (
	drainingFile   = "drone-agent.drain"
	oauthTokenPath = "/var/lib/swarming/oauth_bot_token.json"
)

var (
	queenService = os.Getenv("DRONE_AGENT_QUEEN_SERVICE")
	// DRONE_AGENT_SWARMING_URL is the URL of the Swarming
	// instance.  Should be a full URL without the path,
	// e.g. https://host.example.com
	swarmingURL       = os.Getenv("DRONE_AGENT_SWARMING_URL")
	dutCapacity       = getIntEnv("DRONE_AGENT_DUT_CAPACITY", 10)
	reportingInterval = time.Duration(getIntEnv("DRONE_AGENT_REPORTING_INTERVAL_MINS", 1)) * time.Minute

	authOptions = auth.Options{
		Method:                 auth.ServiceAccountMethod,
		ServiceAccountJSONPath: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}
	workingDirPath = filepath.Join(os.Getenv("HOME"), "skylab_bots")
	// hive value of the drone agent.  This is used for DUT/drone affinity.
	// A drone is assigned DUTs with same hive value.
	hive = initializeHive(os.Getenv("DRONE_AGENT_HIVE"), os.Getenv("DOCKER_DRONE_SERVER_NAME"))

	// tsmonEndpoint is the URL (including file://, https://,
	// pubsub://project/topic) to post monitoring metrics to.
	// If empty, we will try to load configuration from LUCI tsmon default
	// configuration file, i.e. /etc/chrome-infra/ts-mon.json.
	tsmonEndpoint       = os.Getenv("DRONE_AGENT_TSMON_ENDPOINT")
	tsmonCredentialPath = os.Getenv("DRONE_AGENT_TSMON_CREDENTIAL_PATH")

	// botPrefix is used as the prefix for the bot ID.
	// If DRONE_AGENT_BOT_PREFIX env is not set, then 'crossk-' will be used as default
	botPrefix = getEnv("DRONE_AGENT_BOT_PREFIX", "crossk-")

	// Bot compute resources settings.
	// Block IO throttle settings. 0 means no throttling. Only /dev/sda (device
	// number 8:0) is supported.
	botBlkIOReadBPS  = getIntEnv("DRONE_AGENT_BOT_BLKIO_READ_BPS", 0)
	botBlkIOWriteBPS = getIntEnv("DRONE_AGENT_BOT_BLKIO_WRITE_BPS", 0)
)

// Flag options.
var (
	// versionFilePath is the path to a drone-agent version file.
	// This file should only contain the version i.e. 12345.
	versionFilePath = flag.String("version-file", "", "Path for drone-agent version file."+
		" This is reported to drone queen for analytics.")
	// traceBackend denotes the backend used for OTel traces.
	traceBackend string
	// traceTarget is the destination for traces.
	traceTarget = flag.String("trace-target", "", "Traces destination. "+
		"See \"trace-backend\" description for usage.")
)

func init() {
	const desc = `Exporter for OTel traces. Valid options are "console", "grpc" and "none". Default is "none".
For values other than "none", -trace-target must be set.
For "grpc", the format is "host:port" for an OTel collector service.
For "console", -trace-target should be a path for the output file.`
	flag.Func("trace-backend", desc, func(s string) error {
		switch s {
		case "console", "grpc", "none":
			traceBackend = s
			return nil
		default:
			return errors.Reason("invalid value %s. Allowed values are: %s", s, "console, grpc, none").Err()
		}
	})
}

func main() {
	flag.Parse()
	if err := innerMain(); err != nil {
		log.Fatal(err)
	}
}

func innerMain() error {
	// TODO(ayatane): Add environment validation.
	ctx, cancel := context.WithCancel(context.Background())

	version := readVersionFile(*versionFilePath)
	log.Printf("version: %v\n", version)
	ctx = metadata.AppendToOutgoingContext(ctx, "drone-agent-version", version)

	ctx = notifySIGTERM(ctx)
	ctx = notifyDraining(ctx, filepath.Join(workingDirPath, drainingFile))

	if err := metrics.Setup(ctx, tsmonEndpoint, tsmonCredentialPath); err != nil {
		log.Printf("Skipping metrics setup: %s", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()
	defer metrics.Shutdown(ctx)

	if traceBackend != "" && traceBackend != "none" {
		// Initialize tracing.
		exp, err, close := initSpanExporter(ctx, traceBackend, *traceTarget)
		if err != nil {
			return err
		}
		defer close()
		cleanup := tracing.InitTracer(ctx, exp, version)
		defer cleanup(ctx)
	}

	authn := auth.NewAuthenticator(ctx, auth.SilentLogin, authOptions)

	r, err := tokman.Make(authn, oauthTokenPath, time.Minute)
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		r.KeepNew(ctx)
		wg.Done()
	}()

	h, err := authn.Client()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(workingDirPath, 0777); err != nil {
		return err
	}

	a := agent.Agent{
		Client: api.NewDronePRPCClient(&prpc.Client{
			C:    h,
			Host: queenService,
		}),
		SwarmingURL:       swarmingURL,
		WorkingDir:        workingDirPath,
		ReportingInterval: reportingInterval,
		DUTCapacity:       dutCapacity,
		StartBotFunc:      bot.NewStarter(h).Start,
		Hive:              hive,
		BotPrefix:         botPrefix,
		BotResources:      makeBotResources(),
	}
	a.Run(ctx)
	return nil
}

// readVersionFile reads drone agent version from a given version file.
func readVersionFile(versionFilePath string) string {
	const fallback = "unknown"
	if versionFilePath == "" {
		log.Println("no path to version file provided")
		return fallback
	}
	fileContent, err := os.ReadFile(versionFilePath)
	if err != nil {
		log.Printf("cannot read version file: %v", err)
		return fallback
	}
	version := string(fileContent)
	// Simple validation for now, to check that the version string only contains numbers.
	if _, err := strconv.Atoi(version); err != nil {
		log.Printf("illegal version string passed, version should only contain numbers")
		return fallback
	}
	return version
}

const checkDrainingInterval = time.Minute

// notifyDraining returns a context that is marked as draining when a
// file exists at the given path.
func notifyDraining(ctx context.Context, path string) context.Context {
	ctx, drain := draining.WithDraining(ctx)
	_, err := os.Stat(path)
	if err == nil {
		drain()
		return ctx
	}
	go func() {
		for {
			time.Sleep(checkDrainingInterval)
			_, err := os.Stat(path)
			if err == nil {
				drain()
				return
			}
		}
	}()
	return ctx
}

// getIntEnv gets an int value from an environment variable.  If the
// environment variable is not valid or is not set, use the default value.
func getIntEnv(key string, defaultValue int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("Invalid %s, using default value (error: %v)", key, err)
		return defaultValue
	}
	return n
}

// getEnv gets a string value from an environment variable.  If the
// environment variable not set, use the default value.
// If the environment variable is set with enplty value, the function
// will return empty.
func getEnv(key string, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}

// dcLabRegex is the regular expression to identify the Drone server is in a
// data center like lab, e.g. SFO36, in which the server name is like
// 'kube<N>-<SITE>'. If matched, we use the part of '<SITE>' as the hive.
var dcLabRegex = regexp.MustCompile(`^kube[0-9]+-([a-z]+)`)

// initializeHive returns the hive for the agent.
// If hive is not specified, we try to guess it from the hostname.
// The input args are from some envvars, but we don't get them from inside
// the function, so we can keep all code using envvars in a single code block at
// the head of this file for better readability.
func initializeHive(explicitHive, hostname string) string {
	if explicitHive != "" {
		return explicitHive
	}
	log.Printf("Hive not explicitly specified, now guess it by hostname %q", hostname)
	if m := dcLabRegex.FindStringSubmatch(hostname); m != nil {
		return m[1]
	}
	return ""
}

// makeBotResources returns a struct which defines the resources assigned to
// each bot.
func makeBotResources() *specs.LinuxResources {
	// 8 and 0 is major/minor device number of /dev/sda mounted to
	// drone containers. So far I don't see any other number than it.
	var diskMajor int64 = 8
	var diskMinor int64 = 0

	var spec specs.LinuxBlockIO
	if rate := uint64(botBlkIOReadBPS); rate > 0 {
		spec.ThrottleReadBpsDevice = []specs.LinuxThrottleDevice{*newThrottleDevice(diskMajor, diskMinor, rate)}
	}
	if rate := uint64(botBlkIOWriteBPS); rate > 0 {
		spec.ThrottleWriteBpsDevice = []specs.LinuxThrottleDevice{*newThrottleDevice(diskMajor, diskMinor, rate)}
	}
	return &specs.LinuxResources{
		BlockIO: &spec,
	}
}

// newThrottleDevice returns a new instance of LinuxThrottleDevice.
func newThrottleDevice(major, minor int64, rate uint64) *specs.LinuxThrottleDevice {
	// We cannot use struct literals to initialize this struct because "Major"
	// and "Minor" belong to a nested unexported struct. It has been fixed in
	// the upstream repo
	// https://github.com/opencontainers/runtime-spec/commit/84251a48404b19a99cc1b4a8f00c5b523e0d22d0
	// but is not included in the latest release (v1.0.2) yet.
	// TODO(guocb): initialize with struct literals when a newer release is
	// available.
	dev := specs.LinuxThrottleDevice{Rate: rate}
	dev.Major = major
	dev.Minor = minor
	return &dev
}

// initSpanExporter uses the traceBackend flag to instantiate the relevant span exporter.
// initSpanExporter expects "target" to be specified and will error out if it is not.
func initSpanExporter(ctx context.Context, traceBackend, target string) (_ trace.SpanExporter, _ error, close func() error) {
	log.Printf("trace backend: %v", traceBackend)
	if target == "" {
		return nil, errors.Reason("no trace target provided").Err(), nil
	}
	var exp trace.SpanExporter
	var err error
	var cleanup func() error
	switch traceBackend {
	case "console":
		f, err := os.Create(target)
		if err != nil {
			return nil, err, nil
		}
		cleanup = f.Close
		exp, err = tracing.NewConsoleExporter(f)
		if err != nil {
			cleanup()
			return nil, err, nil
		}
	case "grpc":
		exp, err = tracing.NewGRPCExporter(ctx, target)
		if err != nil {
			return nil, err, nil
		}
	default:
		log.Panicf("unexpected value for trace backend: %v", traceBackend)
	}
	return exp, nil, cleanup
}
