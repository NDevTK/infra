// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
)

// GetUsedFSSpace returns the amount of used space on the filesystem.
func GetUsedFSSpace(ctx context.Context, runner components.Runner, path string) (int64, error) {
	cmd := fmt.Sprintf("df %q --block-size=1 --output=used | sed 1d", path)
	if out, err := runner(ctx, time.Minute, cmd); err != nil {
		return 0, errors.Annotate(err, "get used space: failed to get filesystem usage stats using df").Err()
	} else if bytes, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64); err != nil {
		return 0, errors.Annotate(err, "get used space: failed to parse filesystem useage stats using df").Err()
	} else {
		return bytes, nil
	}
}

// getPartitionNumber gets the partition number of a device.
func getPartitionNumber(ctx context.Context, runner components.Runner, device string) (int, error) {
	cmd := fmt.Sprintf("partx -g -o NR %q", device)
	out, err := runner(ctx, time.Minute, cmd)
	if err != nil {
		return 0, errors.Annotate(err, "get partition number: failed to get partition number with partx").Err()
	}

	id, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, errors.Annotate(err, "get partition number: failed to parse partition number from: %q", out).Err()
	}
	return id, nil
}

// getMountDeviceForPath returns the device/partition that a specific file is associated with.
func getMountDeviceForPath(ctx context.Context, runner components.Runner, path string) (string, error) {
	cmd := fmt.Sprintf("findmnt %q -o source -n", path)
	dev, err := runner(ctx, time.Minute, cmd)
	if err != nil {
		return "", errors.Annotate(err, "get device for path: failed to get mount with findmnt").Err()
	}
	return strings.TrimSpace(dev), nil
}

// getRootDeviceForPath returns the root device that a specific file is associated with.
//
// This differs from GetMountDeviceForPath since it will return the parent device and not the individual partition.
func getRootDeviceForPath(ctx context.Context, runner components.Runner, path string) (string, error) {
	mountDev, err := getMountDeviceForPath(ctx, runner, path)
	if err != nil {
		return "", errors.Annotate(err, "get device for path: failed to get mounted device").Err()
	}

	cmd := fmt.Sprintf("lsblk -pno pkname %q", mountDev)
	rootDev, err := runner(ctx, time.Minute, cmd)
	if err != nil {
		return "", errors.Annotate(err, "get device for path: failed to get root device").Err()
	}
	return strings.TrimSpace(rootDev), nil
}

// partitionHelper is a helper utility to query and interact with device partitions.
type partitionHelper struct {
	runner components.Runner
	device string
}

// partitionInfo is information about a single device partition.
type partitionInfo struct {
	Number int
	Start  float64
	End    float64
	Size   float64
	Type   string
}

// NewPartitionHelperForPath creates a new helper for the device that owns the requested path.
func NewPartitionHelperForPath(ctx context.Context, runner components.Runner, path string) (*partitionHelper, error) {
	rootDev, err := getRootDeviceForPath(ctx, runner, path)
	if err != nil {
		return nil, errors.Annotate(err, "new partition helper: failed to get root device for path %q", path).Err()
	}

	return &partitionHelper{
		runner: runner,
		device: rootDev,
	}, nil
}

// GetPartitionSize gets the partition size of the mounted device.
func (p *partitionHelper) GetPartitionSize(ctx context.Context, path string) (int64, error) {
	device, err := getMountDeviceForPath(ctx, p.runner, path)
	if err != nil {
		return 0, errors.Annotate(err, "get partition size: failed to get mount device").Err()
	}

	partID, err := getPartitionNumber(ctx, p.runner, device)
	if err != nil {
		return 0, errors.Annotate(err, "get partition size: failed to get partition id").Err()
	}

	partitions, err := p.GetPartitionInfo(ctx, "B", true)
	if err != nil {
		return 0, errors.Annotate(err, "get partition size: failed to get partition id").Err()
	}

	for _, partition := range partitions {
		if partition.Number == partID {
			return int64(partition.Size), nil
		}
	}

	return 0, errors.Reason("get partition size: could not find partition info for partition: %d", partID).Err()
}

// GetFreeSpace returns the amount of free space at the end of the partitions.
func (p *partitionHelper) GetFreeSpace(ctx context.Context) (int64, error) {
	partitions, err := p.GetPartitionInfo(ctx, "B", false)
	if err != nil {
		return 0, errors.Annotate(err, "get free space: failed to get partition info").Err()
	}

	if len(partitions) == 0 {
		return 0, errors.Annotate(err, "get free space: no partitions found").Err()
	}

	last := partitions[len(partitions)-1]
	if strings.EqualFold(last.Type, "free") {
		return int64(last.Size), nil
	}
	return 0, nil
}

// GetPartitionInfo returns all of the device partitions.
func (p *partitionHelper) GetPartitionInfo(ctx context.Context, unit string, excludeFree bool) ([]*partitionInfo, error) {
	/*
		BYT;
		/dev/sda:100%:scsi:512:512:msdos: Patriot Memory:;
		1:0.01%:0.88%:0.87%:fat32::lba;
		2:0.88%:29.0%:28.1%:ext4::;
		3:29.0%:31.0%:1.97%:fat32::lba;
		4:31.0%:100%:69.0%:ext4::;
	*/
	cmd := fmt.Sprintf("parted -m %q unit %s print free | tr -d ';'", p.device, unit)
	out, err := p.runner(ctx, time.Minute, cmd)
	if err != nil {
		return nil, errors.Annotate(err, "get partition info: failed to get partition info from parted").Err()
	}

	res := make([]*partitionInfo, 0)
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) < 5 {
			continue
		}

		if strings.EqualFold(fields[0], p.device) {
			continue
		}

		if excludeFree && strings.EqualFold(fields[4], "free") {
			continue
		}

		partNumber, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			return nil, errors.Reason("get partition info: unable to parse partition value: %q", fields[0]).Err()
		}

		val := fields[1][:len(fields[1])-len(unit)]
		start, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, errors.Reason("get partition info: unable to parse start value: %q", val).Err()
		} else if start < 0 {
			return nil, errors.Reason("get partition info: invalid start value: %f", start).Err()
		}

		val = fields[2][:len(fields[2])-len(unit)]
		end, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, errors.Reason("get partition info: unable to parse end value: %q", val).Err()
		} else if end < 0 {
			return nil, errors.Reason("get partition info: invalid end value: %f", end).Err()
		}

		if end < start {
			return nil, errors.Reason("get partition info: invalid start/end values, end: %f,  must be > start: %f", end, start).Err()
		}

		res = append(res, &partitionInfo{
			Number: partNumber,
			Start:  start,
			End:    end,
			Size:   end - start,
			Type:   fields[4],
		})
	}
	return res, nil
}
