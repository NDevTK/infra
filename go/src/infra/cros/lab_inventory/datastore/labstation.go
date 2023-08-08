// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package datastore

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.chromium.org/chromiumos/infra/proto/go/lab"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"

	"infra/cros/lab_inventory/changehistory"
	"infra/libs/skylab/common/heuristics"
)

const (
	servoPortRangeUpperLimit = 9999
	servoPortRangeLowerLimit = 9900
)

// LabstationNotDeployedError is the error raised when the DUT has no deployed
// labstation yet.
type LabstationNotDeployedError struct {
	hostname string
}

func (e *LabstationNotDeployedError) Error() string {
	return fmt.Sprintf("labstation %s was not deployed yet. Deploy it first by `skylab add-labstation` please.", e.hostname)
}

type servoHostRecord struct {
	entity     *DeviceEntity
	message    *lab.ChromeOSDevice
	oldMessage *lab.ChromeOSDevice
}
type servoHostRegistry map[string]*servoHostRecord

// UpdateLabstations updates a labstation's info, e.g. servos
func UpdateLabstations(ctx context.Context, hostname string, servoToDelete, addedDUTs []string) (*lab.ChromeOSDevice, error) {
	q := datastore.NewQuery(DeviceKind).Ancestor(fakeAcestorKey(ctx)).Eq("Hostname", hostname)
	var servoHosts []*DeviceEntity
	if err := datastore.GetAll(ctx, q, &servoHosts); err != nil {
		return nil, errors.Annotate(err, "get servo host %s", hostname).Err()
	}
	if len(servoHosts) == 0 {
		return nil, errors.Reason("No such labstation: %s", hostname).Err()
	}
	if len(servoHosts) > 1 {
		return nil, errors.Reason("multiple servo host with same name '%s'", hostname).Err()
	}
	entity := servoHosts[0]
	var l lab.ChromeOSDevice
	if err := proto.Unmarshal(entity.LabConfig, &l); err != nil {
		return nil, errors.Annotate(err, "unmarshal labstation message").Err()
	}
	if l.GetLabstation() == nil {
		return nil, errors.Reason("%s is not a valid labstation hostname", hostname).Err()
	}

	var oldLabstation lab.ChromeOSDevice
	proto.Merge(&oldLabstation, &l)

	l.GetLabstation().Servos = deleteServosBySN(l.GetLabstation().GetServos(), servoToDelete)
	if len(addedDUTs) > 0 {
		dutServos, err := getDUTServoByHostname(ctx, addedDUTs)
		if err != nil {
			return nil, errors.Annotate(err, "failed add get servo from DUT").Err()
		}
		servos, err := addServosFromDUTs(hostname, l.GetLabstation().GetServos(), dutServos)
		if err != nil {
			return nil, errors.Annotate(err, "failed add servo from DUT").Err()
		}
		l.GetLabstation().Servos = servos
	}
	newConfig, err := proto.Marshal(&l)
	if err != nil {
		return nil, errors.Reason("fail to marshal the new labstation message after updating servos").Err()
	}
	changes := changehistory.LogChromeOSLabstationChange(&oldLabstation, &l)
	now := time.Now().UTC()
	f := func(ctx context.Context) error {
		if err := changes.SaveToDatastore(ctx); err != nil {
			return errors.Annotate(err, "save labstation changes").Err()
		}
		entity.LabConfig = newConfig
		entity.Updated = now
		if err := datastore.Put(ctx, []*DeviceEntity{entity}); err != nil {
			return errors.Annotate(err, "save labstations back to datastore").Err()
		}
		return nil
	}
	if err := datastore.RunInTransaction(ctx, f, nil); err != nil {
		return nil, err
	}
	return &l, nil
}

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
			return nil, errors.Annotate(err, "get servo host").Err()
		}
		switch len(servoHosts) {
		case 0:
			return nil, &LabstationNotDeployedError{hostname: hostname}
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
	labstation := r[hostname].message.GetLabstation()
	if labstation == nil {
		// A labstation may be deployed as a DUT due to errors. We check here in
		// case the checking in deployment doesn't work in some cases.
		return nil, errors.Reason("device was not deployed as a labstation: %s", hostname).Err()
	}
	return labstation, nil
}

// When we create/update a DUT, we must also add/update the servo information to the
// associated labstation. Optionally, we should also assign a servo port to the
// DUT.
func (r servoHostRegistry) amendServoToLabstation(ctx context.Context, d *lab.DeviceUnderTest, oldServo *lab.Servo, assignServoPort bool) error {
	servo := d.GetPeripherals().GetServo()
	if servo == nil {
		return nil
	}
	servoHostname := servo.GetServoHostname()

	// If the servo host is a servo v3, just assign port 9999 to servo if
	// required, and do nothing else since one servo host has maximum one servo.
	if !heuristics.LooksLikeLabstation(servoHostname) {
		if assignServoPort && servo.GetServoPort() == 0 {
			servo.ServoPort = servoPortRangeUpperLimit
		}
		return nil
	}
	if servo.GetServoSerial() == "" {
		return errors.Reason("the servo serial number missed for: %q", servoHostname).Err()
	}

	// For labstation, we need to merge the current servo information to the
	// labstation servo list.
	servoHost, err := r.getServoHost(ctx, servoHostname)
	if err != nil {
		return err
	}
	servos := servoHost.GetServos()

	if oldServo != nil {
		logging.Infof(ctx, "delete the old servo: %v", oldServo)
		if servoHostname == oldServo.GetServoHostname() {
			servos = deleteServo(servos, oldServo)
		} else {
			oldServoHost, err := r.getServoHost(ctx, oldServo.GetServoHostname())
			if err != nil {
				logging.Infof(ctx, "Fail to find labstation: %v for old servo: %v", oldServo.GetServoHostname(), err)
			} else {
				oldServoHost.Servos = deleteServo(oldServoHost.GetServos(), oldServo)
			}
		}
	}

	if assignServoPort && servo.GetServoPort() == 0 {
		p, err := firstFreePort(servos)
		if err != nil {
			return err
		}
		servo.ServoPort = int32(p)
	} else {
		if err := checkDuplicatePort(servo, servos); err != nil {
			return err
		}
		if err := checkDuplicateSerial(servo, servos); err != nil {
			return err
		}
	}

	servoHost.Servos = mergeServo(servos, servo)
	return nil
}

// When we delete a DUT, we must also delete the servo information to the
// associated labstation.
func (r servoHostRegistry) removeDeviceFromLabstation(ctx context.Context, d *lab.DeviceUnderTest) error {
	servo := d.GetPeripherals().GetServo()
	if servo == nil || servo.GetServoSerial() == "" {
		return nil
	}
	servoHostname := servo.GetServoHostname()

	// Skip the deleting when the servo host is a servo v3
	if !heuristics.LooksLikeLabstation(servoHostname) {
		return nil
	}

	// For labstation, we need to delete the current servo information from the labstation servo list.
	servoHost, err := r.getServoHost(ctx, servoHostname)
	if err != nil {
		return err
	}
	servoHost.Servos = deleteServo(servoHost.GetServos(), servo)
	return nil
}

func (r servoHostRegistry) saveToDatastore(ctx context.Context) error {
	// Only save the entities that changed.
	now := time.Now().UTC()
	var entities []*DeviceEntity
	var changes changehistory.Changes

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
			changes = append(changes, changehistory.LogChromeOSDeviceChanges(v.oldMessage, v.message)...)
		}
	}
	logging.Infof(ctx, "save %d records of labstation changes", len(changes))
	if err := changes.SaveToDatastore(ctx); err != nil {
		logging.Errorf(ctx, "failed to save labstation changes: %s", err)
	}
	logging.Infof(ctx, "%d labstations changed because of DUT changes", len(entities))
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

// checkDuplicatePort verify that labstation does not have servo with the same port
func checkDuplicatePort(newServo *lab.Servo, servos []*lab.Servo) error {
	newPort := newServo.GetServoPort()
	for _, s := range servos {
		if newPort == s.GetServoPort() {
			return status.Errorf(codes.FailedPrecondition, "the servo port: '%d' is already used in %q", newPort, s.GetServoHostname())
		}
	}
	return nil
}

// checkDuplicateSerial verify that labstation does not have servo with the same serial number
func checkDuplicateSerial(newServo *lab.Servo, servos []*lab.Servo) error {
	newSerial := newServo.GetServoSerial()
	for _, s := range servos {
		if newSerial == s.GetServoSerial() {
			return status.Errorf(codes.FailedPrecondition, "the servo with serial number: %q is already attached to %q", newSerial, s.GetServoHostname())
		}
	}
	return nil
}

// A servo is identified by its serial number. If the SN of servo to be merged
// is existing, then overwrite the existing record.
func mergeServo(servos []*lab.Servo, servo *lab.Servo) []*lab.Servo {
	mapping := map[string]*lab.Servo{}
	for _, s := range servos {
		if s.GetServoSerial() == "" {
			// to avoid cases when serial is empty.
			continue
		}
		mapping[s.GetServoSerial()] = s
	}
	mapping[servo.GetServoSerial()] = servo

	result := make([]*lab.Servo, 0, len(mapping))
	for _, v := range mapping {
		result = append(result, v)
	}
	return result
}

// A servo is identified by its serial number. If the SN of servo to be delete is existing, then overwrite the existing record.
func deleteServo(servos []*lab.Servo, servo *lab.Servo) []*lab.Servo {
	var result []*lab.Servo
	serial := servo.GetServoSerial()
	for _, s := range servos {
		if s.GetServoSerial() == serial {
			// that is servo we need to delete
			continue
		}
		result = append(result, s)
	}
	return result
}

func deleteServosBySN(servos []*lab.Servo, servoToDelete []string) []*lab.Servo {
	servoToDeleteMap := make(map[string]bool, len(servoToDelete))
	for _, s := range servoToDelete {
		servoToDeleteMap[s] = true
	}
	var newServos []*lab.Servo
	for _, servo := range servos {
		if _, ok := servoToDeleteMap[servo.GetServoSerial()]; !ok {
			newServos = append(newServos, servo)
		}
	}
	return newServos
}

func addServosFromDUTs(hostname string, servos, dutServos []*lab.Servo) ([]*lab.Servo, error) {
	servoBySN := make(map[string]*lab.Servo, len(servos))
	servoByPort := make(map[int32]bool, len(servos))
	for _, s := range servos {
		servoBySN[s.GetServoSerial()] = s
		servoByPort[s.GetServoPort()] = true
	}
	for _, servo := range dutServos {
		if servo.GetServoHostname() != hostname {
			return nil, errors.Reason("DUT does not use selected labstation as servo_hostname for %s", hostname).Err()
		} else if servo.GetServoSerial() == "" {
			return nil, errors.Reason("DUT servo does not have serial number").Err()
		} else if _, ok := servoBySN[servo.GetServoSerial()]; ok {
			return nil, errors.Reason("The labstation %s already has servo with serial number: %s", hostname, servo.GetServoSerial()).Err()
		} else if _, ok := servoByPort[servo.GetServoPort()]; ok {
			return nil, errors.Reason("The labstation %s already has servo with port number: %d", hostname, servo.GetServoPort()).Err()
		}
		servos = append(servos, servo)
	}
	return servos, nil
}
