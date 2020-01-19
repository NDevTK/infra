// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package datastore

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"go.chromium.org/chromiumos/infra/proto/go/lab"
	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

const (
	servoPortRangeUpperLimit = 9999
	servoPortRangeLowerLimit = 9900
)

type servoHostRecord struct {
	entity     *DeviceEntity
	message    *lab.ChromeOSDevice
	oldMessage *lab.ChromeOSDevice
}
type servoHostRegistry map[string]*servoHostRecord

// newServohostRegistryFromProtoMsgs creates a new servoHostRegistry instance
// with slice of lab.ChromeOSDevice to be added to datastore.
// This is useful when deploy labstations and DUTs together in one RPC call.
func newServoHostRegistryFromProtoMsgs(ctx context.Context, devices []*lab.ChromeOSDevice) servoHostRegistry {
	r := servoHostRegistry{}
	for _, d := range devices {
		l := d.GetLabstation()
		if l == nil {
			continue
		}
		r[l.GetHostname()] = &servoHostRecord{
			entity: &DeviceEntity{
				ID:       DeviceEntityID(d.GetId().GetValue()),
				Hostname: l.GetHostname(),
				Parent:   fakeAcestorKey(ctx),
			},
			// Initialize `message` and `oldMessage` to same since it will be
			// saved in `AddDevices` call if nothing changed.
			message:    d,
			oldMessage: proto.Clone(d).(*lab.ChromeOSDevice),
		}
	}
	return r
}

func (r servoHostRegistry) getServoHost(ctx context.Context, hostname string) (*lab.Labstation, error) {
	if _, ok := r[hostname]; !ok {
		q := datastore.NewQuery(DeviceKind).Ancestor(fakeAcestorKey(ctx)).Eq("Hostname", hostname)
		var servoHosts []*DeviceEntity
		if err := datastore.GetAll(ctx, q, &servoHosts); err != nil {
			return nil, errors.Annotate(err, "get servo host when add devices").Err()
		}
		switch len(servoHosts) {
		case 0:
			// Create a labstation record automatically.
			id := fmt.Sprintf("%s:%s", UUIDPrefix, uuid.New().String())
			message := &lab.ChromeOSDevice{
				Id: &lab.ChromeOSDeviceID{Value: id},
				Device: &lab.ChromeOSDevice_Labstation{
					Labstation: &lab.Labstation{Hostname: hostname}},
			}
			labConfig, err := proto.Marshal(message)
			if err != nil {
				return nil, errors.Annotate(err, "marshal labstation message").Err()
			}
			r[hostname] = &servoHostRecord{
				entity: &DeviceEntity{
					ID:        DeviceEntityID(id),
					Hostname:  hostname,
					Parent:    fakeAcestorKey(ctx),
					LabConfig: labConfig,
				},
				// `oldMessage` is nil, so this entity will be saved.
				oldMessage: nil,
				message:    message,
			}
		case 1:
			entity := servoHosts[0]
			var crosDev lab.ChromeOSDevice
			if err := proto.Unmarshal(entity.LabConfig, &crosDev); err != nil {
				return nil, errors.Annotate(err, "unmarshal labstation message").Err()
			}
			r[hostname] = &servoHostRecord{
				entity:     entity,
				message:    &crosDev,
				oldMessage: proto.Clone(&crosDev).(*lab.ChromeOSDevice),
			}
		default:
			logging.Errorf(ctx, "multiple servo host with name '%s'", hostname)
			return nil, errors.Reason("multiple servo host with same name '%s'", hostname).Err()
		}

	}
	return r[hostname].message.GetLabstation(), nil
}

// When we create/update a DUT, we must also add/update the servo information to the
// associated labstation. Optionally, we should also assign a servo port to the
// DUT.
func (r servoHostRegistry) amendServoToLabstation(ctx context.Context, d *lab.DeviceUnderTest, assignServoPort bool) error {
	servo := d.GetPeripherals().GetServo()
	if servo == nil {
		return nil
	}
	servoHostname := servo.GetServoHostname()
	servoHost, err := r.getServoHost(ctx, servoHostname)
	if err != nil {
		return err
	}
	servos := servoHost.GetServos()

	if assignServoPort && servo.GetServoPort() == 0 {
		p, err := firstFreePort(servos)
		if err != nil {
			return err
		}
		servo.ServoPort = int32(p)
	}

	servoHost.Servos = mergeServo(servos, servo)
	return nil
}

func (r servoHostRegistry) saveToDatastore(ctx context.Context) error {
	// Only save the entities that changed.
	now := time.Now().UTC()
	var entities []*DeviceEntity
	for _, v := range r {
		// Sort servos by port number.
		servos := v.message.GetLabstation().Servos
		sort.Slice(servos, func(i, j int) bool {
			return servos[i].GetServoPort() > servos[j].GetServoPort()
		})
		if !proto.Equal(v.oldMessage, v.message) {
			labConfig, err := proto.Marshal(v.message)
			if err != nil {
				return errors.Annotate(err, "marshal labstation message").Err()
			}
			v.entity.LabConfig = labConfig
			v.entity.Updated = now

			entities = append(entities, v.entity)
		}
	}
	if err := datastore.Put(ctx, entities); err != nil {
		return errors.Annotate(err, "save labstations back to datastore").Err()
	}
	return nil
}

func firstFreePort(servos []*lab.Servo) (int, error) {
	var used []int
	for _, s := range servos {
		used = append(used, int(s.GetServoPort()))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(used)))
	// This range is consistent with the range of ports generated by servod:
	// https://chromium.googlesource.com/chromiumos/third_party/hdctools/+/cf5f8027b9d3015db75df4853e37ea7a2f1ac538/servo/servod.py#36
	portCount := len(used)
	for idx := 0; idx < (servoPortRangeUpperLimit - servoPortRangeLowerLimit); idx++ {
		p := servoPortRangeUpperLimit - idx
		if idx >= portCount {
			return p, nil
		}
		if used[idx] < p {
			return p, nil
		}
	}
	return 0, errors.Reason("no free servo port available").Err()
}

// A servo is identified by its serial number. If the SN of servo to be merged
// is existing, then overwrite the existing record.
func mergeServo(servos []*lab.Servo, servo *lab.Servo) []*lab.Servo {
	mapping := map[string]*lab.Servo{}
	for _, s := range servos {
		mapping[s.GetServoSerial()] = s
	}
	mapping[servo.GetServoSerial()] = servo

	result := make([]*lab.Servo, 0, len(mapping))
	for _, v := range mapping {
		result = append(result, v)
	}
	return result
}
