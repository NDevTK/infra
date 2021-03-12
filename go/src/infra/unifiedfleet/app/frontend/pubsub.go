package frontend

import (
	"context"
	"regexp"

	"github.com/golang/protobuf/proto"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/router"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/util"
)

var macAddress = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$`)

// HaRTPushHandler handles the pubsub push responses from HaRT pubsub
//
// Decodes the response sent by PubSub and updates datastore. It doesn't
// return anything as required by https://cloud.google.com/pubsub/docs/push,
// this is because by default the return is 200 OK for http POST requests.
// It does return a http error if the datastore update fails.
func HaRTPushHandler(routerContext *router.Context) {
	// Add namespace as the response from HaRT doesn't have namespace.
	ctx, err := util.SetupDatastoreNamespace(routerContext.Context, util.OSNamespace)
	if err != nil {
		logging.Errorf(ctx, "HaRTPushHandler - Failed to add namespace to context")
		return
	}
	res, err := util.NewPSRequest(routerContext.Request)
	if err != nil {
		logging.Errorf(ctx, "HaRTPushHandler - Failed to read push req %v", err)
		return
	}
	data, err := res.DecodeMessage()
	if err != nil {
		logging.Errorf(ctx, "HaRTPushHandler - Failed to read data %v", err)
		return
	}
	// Decode the proto contained in the payload
	var response ufspb.AssetInfoResponse
	perr := proto.Unmarshal(data, &response)
	if perr != nil {
		// Avoid returning error, as the data contains some assets not
		// known to HaRT and those will always fail.
		logging.Errorf(ctx, "HaRTPushHandler - Failed to decode proto %v", perr)
		return
	}
	if response.GetRequestStatus() == ufspb.RequestStatus_OK {
		allinfo := response.GetAssets()
		logging.Infof(ctx, "HaRTPushHandler - Updating %v assets", len(allinfo))
		assetsToUpdate := make([]*ufspb.Asset, 0, len(allinfo))

		f := func(ctx context.Context) error {
			for _, iv2assetinfo := range allinfo {
				ufsAsset, err := registration.GetAsset(ctx, iv2assetinfo.GetAssetTag())
				if err != nil {
					logging.Warningf(ctx, "HaRTPushHandler - Cannot update asset [%v], not found in DS", iv2assetinfo.GetAssetTag())
					continue
				}
				if info := updateAssetInfoFromHart(ufsAsset.GetInfo(), iv2assetinfo); info != nil {
					logging.Debugf(ctx, "HaRTPushHandler - Updating %v", ufsAsset.GetName())
					ufsAsset.Info = info
					// Copy model if not there, Avoiding a situation where hart does a wrong name update
					if ufsAsset.Model != "" {
						ufsAsset.Model = ufsAsset.Info.Model
					}
					assetsToUpdate = append(assetsToUpdate, ufsAsset)
				}
			}
			if _, err = registration.BatchUpdateAssets(ctx, assetsToUpdate); err != nil {
				return err
			}
			return nil
		}
		if datastore.RunInTransaction(ctx, f, nil); err != nil {
			logging.Errorf(ctx, "Failed to update assets %v", err)
		}
	}
	logging.Debugf(ctx, "Status: %v", response.GetRequestStatus())
	missing := response.GetMissingAssetTags()
	logging.Debugf(ctx, "Missing[%v]: %v", len(missing), missing)
	failed := response.GetFailedAssetTags()
	logging.Debugf(ctx, "Failed[%v]: %v", len(failed), failed)
	logging.Debugf(ctx, "Success reported for %v assets", len(response.GetAssets()))
	return
}

// updateAssetInfoFromHart copies cost_center, google_code_name, model, build_target, reference_board and phase
// from hartAssetInfo if any of these were updated.
func updateAssetInfoFromHart(ufsAssetInfo, hartAssetInfo *ufspb.AssetInfo) *ufspb.AssetInfo {
	var updated bool
	if ufsAssetInfo == nil {
		ufsAssetInfo = &ufspb.AssetInfo{
			AssetTag: hartAssetInfo.GetAssetTag(),
		}
	}
	if ufsAssetInfo.GetCostCenter() != hartAssetInfo.GetCostCenter() {
		updated = true
		// Update CostCenter if it's changed
		ufsAssetInfo.CostCenter = hartAssetInfo.GetCostCenter()
	}
	if ufsAssetInfo.GetGoogleCodeName() != hartAssetInfo.GetGoogleCodeName() {
		updated = true
		// Update GoogleCodeName if it's changed
		ufsAssetInfo.GoogleCodeName = hartAssetInfo.GetGoogleCodeName()
	}
	if ufsAssetInfo.GetModel() != hartAssetInfo.GetModel() {
		updated = true
		// Update Model if it's changed
		ufsAssetInfo.Model = hartAssetInfo.GetModel()
	}
	if ufsAssetInfo.GetBuildTarget() != hartAssetInfo.GetBuildTarget() {
		updated = true
		// Update BuildTarget if it's changed
		ufsAssetInfo.BuildTarget = hartAssetInfo.GetBuildTarget()
	}
	if ufsAssetInfo.GetReferenceBoard() != hartAssetInfo.GetReferenceBoard() {
		updated = true
		// Update ReferenceBoard if it's changed
		ufsAssetInfo.ReferenceBoard = hartAssetInfo.GetReferenceBoard()
	}
	if ufsAssetInfo.GetPhase() != hartAssetInfo.GetPhase() {
		updated = true
		// Update Phase if it's changed
		ufsAssetInfo.Phase = hartAssetInfo.GetPhase()
	}
	// Avoid write to DB if nothing was updated
	if updated {
		return ufsAssetInfo
	}
	return nil
}
