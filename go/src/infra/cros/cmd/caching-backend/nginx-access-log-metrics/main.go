// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"strings"
	"time"

	"go.chromium.org/luci/common/tsmon"
	"go.chromium.org/luci/common/tsmon/distribution"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/common/tsmon/target"

	"infra/cros/cmd/caching-backend/nginx-access-log-metrics/internal/filetailer"
)

type record struct {
	Timestamp     time.Time `json:"access_time"`
	ClientIP      string    `json:"remote_addr"`
	HTTPMethod    string    `json:"method"`
	Path          string    `json:"uri"`
	Status        int       `json:"status"`
	BodyBytesSent int       `json:"bytes_sent"`
	ExpectedSize  int       `json:"content_length"`
	RequestTime   float64   `json:"request_time"`
	CacheStatus   string    `json:"upstream_cache_status"`
	ProxyHost     string    `json:"proxy_host"`
}

var (
	inputLogFile        = flag.String("input-log-file", "/var/log/nginx/gs-cache.access.log", "Nginx access log for gs_cache")
	tsmonCredentialPath = flag.String("tsmon-credentials", "", "Path to a pkcs8 json credential file")
	tsmonEndpoint       = flag.String("tsmon-endpoint", "", "URL (including file://, https://, pubsub://project/topic) to post monitoring metrics to")
	tsmonTaskHostname   = flag.String("tsmon-task-hostname", "", "Name of the host reported to tsmon. (default is the hostname)")
)

func main() {
	if err := innerMain(); err != nil {
		log.Fatalf("Exiting due to an error: %s", err)
	}
	log.Printf("Exiting successfully")
}

func innerMain() error {
	flag.Parse()

	ctx := context.Background()
	setupTsMon(ctx)
	defer shutdownTsMon(ctx)
	// We set up context cancellation after tsmon setup because we want tsmon
	// to finish flushing.
	ctx = cancelOnSignals(ctx)

	tailer, err := filetailer.New(*inputLogFile)
	if err != nil {
		return err
	}
	defer tailer.Close()

	go func() {
		for tailer.Scan() {
			if r := parseLine(tailer.Text()); r != nil {
				reportToTsMon(r)
			}
		}
	}()
	<-ctx.Done()
	return nil
}

// parseLine parses a log line and generate a record if the line is want.
func parseLine(line string) *record {
	r := record{}
	if err := json.Unmarshal([]byte(line), &r); err != nil {
		log.Printf("Parsing error: %s", err)
		return nil
	}

	// Ignore client side errors (4xx) except NotFound(404).
	// We may receive many client bad requests from some service scanner, so
	// ignore them.
	if r.Status >= 400 && r.Status < 500 && r.Status != 404 {
		log.Printf("Ignore client error %d", r.Status)
		return nil
	}
	// Ignore non cache related requests.
	if ignoredPath(r.Path) {
		log.Printf("Ignore path %q", r.Path)
		return nil
	}

	return &r
}

// ignoredPath ignores path/URL not related to caching.
func ignoredPath(path string) bool {
	switch {
	case path == "/":
		return true
	case path == "/static/quick-provision":
		return true
	case path == "/download/chromeos-image-archive":
		return true
	case path == "/download/chromeos-releases":
		return true
	case strings.HasPrefix(path, "/check_health"):
		return true
	case strings.HasPrefix(path, "/stage"):
		return true
	case strings.HasPrefix(path, "/is_staged"):
		return true
	}
	return false
}

// setupTsMon set up tsmon.
func setupTsMon(ctx context.Context) {
	fl := tsmon.NewFlags()
	fl.Endpoint = *tsmonEndpoint
	fl.Credentials = *tsmonCredentialPath
	fl.Flush = tsmon.FlushAuto
	fl.Target.TaskHostname = *tsmonTaskHostname
	fl.Target.SetDefaultsFromHostname()
	fl.Target.TargetType = target.TaskType
	fl.Target.TaskServiceName = "caching_backend"
	fl.Target.TaskJobName = "nginx"

	if err := tsmon.InitializeFromFlags(ctx, &fl); err != nil {
		log.Printf("Skipping tsmon setup: %s", err)
	}
}

// shutdownTsMon shuts down tsmon.
func shutdownTsMon(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	log.Printf("Shutting down tsmon...")
	tsmon.Shutdown(ctx)
}

// respBytesSent is a tsmon metric for the response bytes sent to clients.
var respBytesSent = metric.NewCounter("chromeos/caching_backend/nginx/response_bytes_sent",
	"response bytes sent to clients from a caching backend",
	nil,
	field.String("http_method"),
	field.String("rpc"),
	field.Int("status"),
	field.String("cache"),
	field.Bool("full_download"),
	field.Bool("internal_traffic"),
	field.String("level"),
)

// respBytesPerSecond is a tsmon metric for the response bandwidth.
var respBytesPerSecond = metric.NewCumulativeDistribution("chromeos/caching_backend/nginx/response_bandwidth",
	"distribution of response bandwidth (Byte/second) to clients from a caching backend",
	nil,
	// The bucket covers from 2^0 to 2^25 (i.e. 128 MB/s) which is the range we
	// concern.
	distribution.GeometricBucketer(2, 25),
	field.String("http_method"),
	field.String("rpc"),
	field.Int("status"),
	field.String("cache"),
	field.Bool("full_download"),
	field.Bool("internal_traffic"),
	field.String("level"),
)

// reportToTsMon reports the parsed log line data to tsmon server.
func reportToTsMon(i *record) {
	// Extract the rpc part from the path (e.g. "/rpc/path/to/file").
	rpc := "unknown"
	if s := strings.SplitN(i.Path, "/", 3); len(s) > 2 {
		rpc = s[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fullDownload := i.ExpectedSize == i.BodyBytesSent
	internalTraffic := i.ClientIP == "127.0.0.1"

	var level string
	switch i.ProxyHost {
	case "l7_upstream":
		level = "L4"
	case "downloader":
		level = "L7"
	default:
		level = "unknown"
	}

	respBytesSent.Add(ctx, int64(i.BodyBytesSent), i.HTTPMethod, rpc, i.Status, i.CacheStatus, fullDownload, internalTraffic, level)

	// Set the response speed metric. The minimum resolution of Nginx request
	// time is 1 ms. For shorter cases, we just set the download speed to a
	// high number as the speed is supposed high enough.
	if speed := 100. * 1024 * 1024; i.RequestTime < 0.001 {
		respBytesPerSecond.Add(ctx, speed, i.HTTPMethod, rpc, i.Status, i.CacheStatus, fullDownload, internalTraffic, level)
	} else {
		respBytesPerSecond.Add(ctx, float64(i.BodyBytesSent)/i.RequestTime, i.HTTPMethod, rpc, i.Status, i.CacheStatus, fullDownload, internalTraffic, level)
	}
}
