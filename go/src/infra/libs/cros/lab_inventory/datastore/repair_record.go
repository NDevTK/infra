package datastore

import (
	"context"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/common/errors"

	inv "infra/appengine/cros/lab_inventory/api/v1"
)

// DeviceManualRepairRecordsOpRes is for use in Datastore to RPC conversions
type DeviceManualRepairRecordsOpRes struct {
	Record *inv.DeviceManualRepairRecord
	Entity *DeviceManualRepairRecordEntity
	Err    error
}

func (r *DeviceManualRepairRecordsOpRes) logError(e error) {
	r.Err = e
}

// GetDeviceManualRepairRecords returns the DeviceManualRepairRecord matching
// the device id ($hostname-$assetTag).
func GetDeviceManualRepairRecords(ctx context.Context, ids []string) []*DeviceManualRepairRecordsOpRes {
	queryResults := make([]*DeviceManualRepairRecordsOpRes, len(ids))
	qrMap := make(map[string]*DeviceManualRepairRecordsOpRes)
	entities := make([]*DeviceManualRepairRecordEntity, 0, len(ids))
	for _, id := range ids {
		res := &DeviceManualRepairRecordsOpRes{
			Entity: &DeviceManualRepairRecordEntity{
				ID: id,
			},
		}
		qrMap[id] = res
		entities = append(entities, res.Entity)
	}
	if err := datastore.Get(ctx, entities); err != nil {
		for i, e := range err.(errors.MultiError) {
			qrMap[entities[i].ID].logError(e)
		}
	}
	for i, id := range ids {
		queryResults[i] = qrMap[id]
	}
	return queryResults
}

// AddDeviceManualRepairRecords creates a DeviceManualRepairRecord with the
// device hostname and adds it to the datastore.
func AddDeviceManualRepairRecords(ctx context.Context, records []*inv.DeviceManualRepairRecord) ([]*DeviceManualRepairRecordsOpRes, error) {
	allResponses := make([]*DeviceManualRepairRecordsOpRes, len(records))
	putEntities := make([]*DeviceManualRepairRecordEntity, 0, len(records))
	putResponses := make([]*DeviceManualRepairRecordsOpRes, 0, len(records))
	var err error

	for i, r := range records {
		res := &DeviceManualRepairRecordsOpRes{
			Record: r,
		}
		allResponses[i] = res
		recordEntity, err := NewDeviceManualRepairRecordEntity(r)
		if err != nil {
			res.logError(err)
			continue
		}
		res.Entity = recordEntity

		putEntities = append(putEntities, recordEntity)
		putResponses = append(putResponses, res)
	}

	f := func(ctx context.Context) error {
		finalEntities := make([]*DeviceManualRepairRecordEntity, 0, len(records))
		finalResponses := make([]*DeviceManualRepairRecordsOpRes, 0, len(records))

		existsArr, err := deviceManualRepairRecordsExists(ctx, putEntities)
		if err == nil {
			for i, pe := range putEntities {
				_, exists := existsArr[i]
				if exists {
					putResponses[i].logError(errors.Reason("Record exists in the datastore").Err())
					continue
				}
				finalEntities = append(finalEntities, pe)
				finalResponses = append(finalResponses, putResponses[i])
			}
		} else {
			finalEntities = putEntities
			finalResponses = putResponses
		}

		if err := datastore.Put(ctx, finalEntities); err != nil {
			for i, e := range err.(errors.MultiError) {
				finalResponses[i].logError(e)
			}
		}
		return nil
	}

	err = datastore.RunInTransaction(ctx, f, nil)
	return allResponses, err
}

// TODO: This is in another CL.
// UpdateDeviceManualRepairRecords updates the DeviceManualRepairRecord matching
// the device hostname in the datastore.
// func UpdateDeviceManualRepairRecords(ctx context.Context, records []*inv.DeviceManualRepairRecord, update bool) ([]*DeviceManualRepairRecordsOpRes, error) {
//
// }

// Checks if the davice manual repair records exist in the datastore.
func deviceManualRepairRecordsExists(ctx context.Context, entities []*DeviceManualRepairRecordEntity) (map[int]bool, error) {
	existsMap := make(map[int]bool, 0)
	res, err := datastore.Exists(ctx, entities)
	if res == nil {
		return existsMap, err
	}
	for i, r := range res.List(0) {
		if r {
			existsMap[i] = true
		}
	}
	return existsMap, err
}
