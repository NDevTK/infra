// Copyright (c) 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package docker

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/tsmon"
	"github.com/luci/luci-go/common/tsmon/field"
	"github.com/luci/luci-go/common/tsmon/metric"
	tsmonTypes "github.com/luci/luci-go/common/tsmon/types"

	"golang.org/x/net/context"
)

var (
	statusMetric = metric.NewString("dev/container/status",
		"Status (running, stopped, etc.) of a container.",
		nil,
		field.String("name"),
		field.String("hostname"))

	uptimeMetric = metric.NewFloat("dev/container/uptime",
		"Uptime (in seconds) of a contianer.",
		&tsmonTypes.MetricMetadata{Units: tsmonTypes.Seconds},
		field.String("name"))

	memUsedMetric = metric.NewInt("dev/container/mem/used",
		"Memory in used by a container.",
		&tsmonTypes.MetricMetadata{Units: tsmonTypes.Bytes},
		field.String("name"))
	memTotalMetric = metric.NewInt("dev/container/mem/total",
		"Total memory avaialable to a container.",
		&tsmonTypes.MetricMetadata{Units: tsmonTypes.Bytes},
		field.String("name"))

	netDownMetric = metric.NewInt("dev/container/net/down",
		"Total bytes of network ingress for the container.",
		&tsmonTypes.MetricMetadata{Units: tsmonTypes.Bytes},
		field.String("name"))
	netUpMetric = metric.NewInt("dev/container/net/up",
		"Total bytes of network egress for the container.",
		&tsmonTypes.MetricMetadata{Units: tsmonTypes.Bytes},
		field.String("name"))
)

// The following is a subset of the fields contained in the json blob returned
// when querying the engine for a container's stats. Only the fields we care
// about are listed here.
type containerStats struct {
	Name     string
	Memory   memoryStats `json:"memory_stats"`
	Networks struct {
		Eth0 struct {
			RxBytes int64 `json:"rx_bytes"`
			TxBytes int64 `json:"tx_bytes"`
		}
	}
}
type memoryStats struct {
	Usage int64
	Limit int64
}

func updateContainerMetrics(c context.Context, container dockerTypes.Container, containerInfo dockerTypes.ContainerJSON, containerStatsJSON dockerTypes.ContainerStats) error {
	// Remove leading slash from container name.
	containerName := strings.TrimPrefix(container.Names[0], "/")
	containerState := container.State
	containerHostname := containerInfo.Config.Hostname
	statusMetric.Set(c, containerState, containerName, containerHostname)
	startTime, err := time.Parse(time.RFC3339Nano, containerInfo.State.StartedAt)
	if err != nil {
		return err
	}
	uptime := clock.Now(c).Sub(startTime).Seconds()
	uptimeMetric.Set(c, uptime, containerName)

	buff := new(bytes.Buffer)
	defer containerStatsJSON.Body.Close()
	if _, err := buff.ReadFrom(containerStatsJSON.Body); err != nil {
		return err
	}
	stats := &containerStats{}
	if err := json.Unmarshal(buff.Bytes(), stats); err != nil {
		return err
	}

	netUp := stats.Networks.Eth0.TxBytes
	netDown := stats.Networks.Eth0.RxBytes
	memTotal := stats.Memory.Limit
	memUsed := stats.Memory.Usage

	memUsedMetric.Set(c, memUsed, containerName)
	memTotalMetric.Set(c, memTotal, containerName)
	netUpMetric.Set(c, netUp, containerName)
	netDownMetric.Set(c, netDown, containerName)
	return nil
}

func inspectContainer(c context.Context, dockerClient *client.Client, container dockerTypes.Container, ch chan error) {
	containerInfo, err := dockerClient.ContainerInspect(c, container.ID)
	if err != nil {
		ch <- err
		return
	}

	// The docker client returns a stream of raw json for a container's stats.
	containerStatsJSON, err := dockerClient.ContainerStats(c, container.ID, false)
	if err != nil {
		ch <- err
		return
	}

	ch <- updateContainerMetrics(c, container, containerInfo, containerStatsJSON)
}

func update(c context.Context) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	_, err = dockerClient.Ping(c)
	if err != nil {
		// Don't log an error if the ping failed. Most bots don't have
		// the docker engine installed and running.
		return nil
	}

	containers, err := dockerClient.ContainerList(c, dockerTypes.ContainerListOptions{All: true})
	if err != nil {
		return err
	}

	// Inspect each container in parallel. This is much faster than doing so in serial.
	var channels []chan error
	for _, container := range containers {
		ch := make(chan error)
		channels = append(channels, ch)
		go inspectContainer(c, dockerClient, container, ch)
	}
	for _, ch := range channels {
		err = <-ch
		if err != nil {
			logging.Errorf(c, "Failed to query docker engine: %s", err)
		}
	}
	return nil
}

// Register adds tsmon callbacks to set docker metrics.
func Register() {
	tsmon.RegisterGlobalCallback(func(c context.Context) {
		if err := update(c); err != nil {
			logging.Errorf(c, "Failed to update Docker metrics: %s", err)
		}
	})
}
