// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import "infra/cros/cmd/kron/builds"

var (
	// allowedConfigs is a quick access tool to check if the SuSch config is
	// being allowed through during the migration.
	//
	// A map was used here to reduce on search complexity.
	// NOTE: not included: multi = multi-dut
	// multi: AU_P2P
	// multi: cellular_callbox_multidut
	// multi: UWB_OTA_uwb_cros_peers_1
	// multi: UWB_OTA_Flaky_uwb_cros_peers_1
	// multi: UWB_OTA_uwb_android_peers_1
	// multi: UWB_OTA_Flaky_uwb_android_peers_1
	// multi: UWB_OTA_uwb_android_peers_2
	// multi: UWB_OTA_Flaky_uwb_android_peers_2
	// multi: UWB_OTA_uwb_cros_peers_1_uwb_android_peers_1
	// multi: UWB_OTA_Flaky_uwb_cros_peers_1_uwb_android_peers_1
	// multi: UWB_OTA_uwb_cros_peers_1_uwb_android_peers_2
	// multi: UWB_OTA_Flaky_uwb_cros_peers_1_uwb_android_peers_2
	// multi: bluetooth_multi_dut_config
	// multi: kernel_wifi__multidut__cross_device_multi_cb_wifi__wifi_cross_device_sap__tauto
	// multi: kernel_wifi__multidut__cross_device_multi_cb_wifi__wifi_cross_device_sap_flaky__tauto
	// multi: kernel_wifi__multidut__cross_device_multi_cb_wifi__wifi_cross_device_p2p__tauto
	// multi: kernel_wifi__multidut__cross_device_multi_cb_wifi__wifi_cross_device_p2p_flaky__tauto
	// multi: kernel_wifi__multidut__cross_device_multi_cb_wifi__wifi_cross_device_multidut__tauto
	// multi: kernel_wifi__multidut__cross_device_multi_cb_wifi__wifi_cross_device_multidut_flaky__tauto
	// multi: NearbyShareArc
	// multi: NearbyShare
	// multi: NearbyShareDev
	// multi: NearbyShareProd
	// multi: NearbyShareRemote
	// multi: CrossDevice
	// multi: CrossDeviceCellular
	// multi: PVSPerBuild
	// multi: PVSPerBuildLatestShop
	// firmware: NissaFaftbiosPerBuild
	// firmware: NissaFaftecPerBuild
	// firmware: DededeFaftbiosPerBuild
	// firmware: DededeFaftecPerBuild
	allowedConfigs = map[string]bool{
		"ARCCorePerBuild":                                  true,
		"ARC_T_PerBuild":                                   true,
		"AUPerBuild":                                       true,
		"AUPerBuildChrome":                                 true,
		"AUPerBuildInterrupt":                              true,
		"AUPerBuildOmahaResponse":                          true,
		"AUPerBuildTast":                                   true,
		"AUPerBuildWiFi":                                   true,
		"AU_M2N":                                           true,
		"AU_M2N_CHROME":                                    true,
		"AU_M2N_FIRMWARE":                                  true,
		"AU_OOBE":                                          true,
		"AVWebcam":                                         true,
		"AppCompatRelease":                                 true,
		"AppCompatRelease-0":                               true,
		"AppCompatRelease-1":                               true,
		"AppCompatSmokebeta":                               true,
		"AppCompatSmokecanary":                             true,
		"AppCompatSmokedev":                                true,
		"AppCompatSmokestable":                             true,
		"ArcCtsDevVMDaily":                                 true,
		"ArcDataSnapshotPerBuild":                          true,
		"AssistantAudiobox":                                true,
		"AudioBasicKernelnextLtToT-1":                      true,
		"AudioBasicKernelnextToT":                          true,
		"AudioBasicKernelnextToT-1":                        true,
		"AudioBasicLtToT-1":                                true,
		"AudioBasicToT":                                    true,
		"AudioBasicToT-1":                                  true,
		"AudioEssentialKernelnextLtToT-1":                  true,
		"AudioEssentialKernelnextToT":                      true,
		"AudioEssentialKernelnextToT-1":                    true,
		"AudioEssentialLtToT-1":                            true,
		"AudioEssentialToT":                                true,
		"AudioEssentialToT-1":                              true,
		"Bluetooth_Flaky_Every_Build":                      true,
		"Bluetooth_Flaky_Floss_Perbuild":                   true,
		"Bluetooth_Floss_Flaky_Every_Build":                true,
		"Bluetooth_Floss_Sa_Kernelnext_Perbuild":           true,
		"Bluetooth_Floss_Sa_Perbuild":                      true,
		"Bluetooth_Sa2_Kernelnext_Perbuild":                true,
		"Bluetooth_Sa2_Perbuild":                           true,
		"Bluetooth_Sa_Kernelnext_Perbuild":                 true,
		"Bluetooth_Sa_Perbuild":                            true,
		"Bluetooth_Stable_Floss_Perbuild":                  true,
		"Bluetooth_Stable_Perbuild":                        true,
		"BorealisPerBuild":                                 true,
		"BvtTastParallelsCritical":                         true,
		"BvtTastParallelsInformational":                    true,
		"CFTNewBuild":                                      true,
		"CTPV2Demo":                                        true,
		"CUJExperimentalPerBuild0":                         true,
		"CUJExperimentalPerBuild1":                         true,
		"CUJExperimentalPerBuild10":                        true,
		"CUJExperimentalPerBuild11":                        true,
		"CUJExperimentalPerBuild12":                        true,
		"CUJExperimentalPerBuild2":                         true,
		"CUJExperimentalPerBuild3":                         true,
		"CUJExperimentalPerBuild4":                         true,
		"CUJExperimentalPerBuild5":                         true,
		"CUJExperimentalPerBuild6":                         true,
		"CUJExperimentalPerBuild7":                         true,
		"CUJExperimentalPerBuild8":                         true,
		"CUJExperimentalPerBuild9":                         true,
		"CUJLoginPerf0":                                    true,
		"CUJLoginPerf1":                                    true,
		"CUJLoginPerf2":                                    true,
		"CUJPerBuild0":                                     true,
		"CUJPerBuild1":                                     true,
		"CUJPerBuild10":                                    true,
		"CUJPerBuild11":                                    true,
		"CUJPerBuild12":                                    true,
		"CUJPerBuild2":                                     true,
		"CUJPerBuild3":                                     true,
		"CUJPerBuild4":                                     true,
		"CUJPerBuild5":                                     true,
		"CUJPerBuild6":                                     true,
		"CUJPerBuild7":                                     true,
		"CUJPerBuild8":                                     true,
		"CUJPerBuild9":                                     true,
		"Camera_Libcamera_HAL":                             true,
		"Camerabox_Facing_Back":                            true,
		"Camerabox_Facing_Front":                           true,
		"CbxDisabledPerBuild":                              true,
		"CbxEnabledPerBuild":                               true,
		"Cellular_OTA_2nd_Source_Perbuild_Dev_Beta_Stable": true,
		"Cellular_OTA_3rd_Source_Perbuild_Dev_Beta_Stable": true,
		"Cellular_OTA_Flaky_Perbuild":                      true,
		"Cellular_OTA_Perbuild":                            true,
		"ChameleonHdmiPerbuild":                            true,
		"ChromeosTastInformational":                        true,
		"ChromeosTastInformationalVM":                      true,
		"ChromeosTastInformationalVM_destructive_func":     true,
		"ChromeosTastInformational_destructive_func":       true,
		"ChromeosTastStaging":                              true,
		"ChromeosTastStagingVM":                            true,
		"Cr_Chromebox_Chromebase":                          true,
		"CrosAVAnalysis":                                   true,
		"CrosAVAnalysisPerBuild":                           true,
		"CrosAVAnalysisTracePerBuild":                      true,
		"CrosboltArcPerfPerbuild":                          true,
		"CrosboltFsiCheck":                                 true,
		"CrosboltPerfParallelsPerbuild":                    true,
		"CrosboltPerfPerbuild":                             true,
		"CrosboltPerfPerbuildBringup":                      true,
		"CrosboltPerfPerbuildForExperimentalBoards":        true,
		"CrosboltReleaseGate":                              true,
		"DededeBvtOemPerBuild":                             true,
		"DededeStorageOemPerBuild":                         true,
		"EnrollmentPerBuild":                               true,
		"FaftFingerprintPerBuild":                          true,
		"FaftFingerprintPerBuildKernelnext":                true,
		"FleetLabqualInformationalTests":                   true,
		"FlexCriticalstagingPerbuild":                      true,
		"GraphicsPerBuild":                                 true,
		"GraphicsPerBuildKernelnext":                       true,
		"GraphicsPerBuild_VC":                              true,
		"HeartdPerbuild":                                   true,
		"Hotrod":                                           true,
		"HotrodRemora":                                     true,
		"IAD65PilotInformational":                          true,
		"InputsAppCompatArcPerbuildTFC":                    true,
		"InputsAppCompatCitrixPerbuildTFC":                 true,
		"LacrosBeta":                                       true,
		"LacrosCanary":                                     true,
		"LacrosDev":                                        true,
		"LacrosStable":                                     true,
		"LauncherImageSearchPerbuildTFC":                   true,
		"LauncherSearchQualityTFC":                         true,
		"MTP":                                              true,
		"NBR":                                              true,
		"NBR_AU":                                           true,
		"NBR_AU_VARIANTS":                                  true,
		"NBR_AU_WITH_DLC":                                  true,
		"NBR_AU_WITH_DLC_VARIANTS":                         true,
		"NBR_VARIANTS":                                     true,
		"NBR_WIFI":                                         true,
		"NBR_WIFI_VARIANTS":                                true,
		"NissaBvtOemPerBuild":                              true,
		"NissaStorageOemPerBuild":                          true,
		"NissaStoragePerBuild":                             true,
		"PasitFastPerbuild":                                true,
		"PasitFullPerbuild":                                true,
		"PlaystorePerBuild":                                true,
		"PlaystoreVMPerBuild":                              true,
		"PowerDashboardPerbuild":                           true,
		"PowerRegression":                                  true,
		"RLZ":                                              true,
		"RaccConfigInstalledPerbuild":                      true,
		"RaccConfigInstalledPerbuildKernelnext":            true,
		"RaccGeneralPerbuild":                              true,
		"RaccGeneralPerbuildKernelnext":                    true,
		"ShimlessRMACalibrationPerBuild":                   true,
		"ShimlessRMANodeLockedPerBuild":                    true,
		"ShimlessRMANormalPerBuild":                        true,
		"ShimlessRMASimplifiedNormalPerBuild":              true,
		"StylusTouchReplay":                                true,
		"Syd_PnP_PerBuild":                                 true,
		"SyncOffloadsPerBuild":                             true,
		"TFCDemo":                                          true,
		"TastCrostiniSlowCq":                               true,
		"TastCrostiniSlowInformational":                    true,
		"TouchpadTouchReplay":                              true,
		"TouchscreenTouchReplay":                           true,
		"TypecDpMcci2Perbuild":                             true,
		"TypecDpMcciPerbuild":                              true,
		"TypecHpdWakeMcciPerbuild":                         true,
		"TypecSmokePerbuild":                               true,
		"TypecUsbMcci2Perbuild":                            true,
		"TypecUsbMcciPerbuild":                             true,
		"UsbDetect":                                        true,
		"UsbDetectKernelnext":                              true,
		"UsbDetectToT":                                     true,
		"UsbDetectToTKernelnext":                           true,
		"VideoConferencePerBuild":                          true,
		"WiFi_Commercial_Flaky_Perbuild":                   true,
		"WiFi_Commercial_Perbuild":                         true,
		"WiFi_EndtoEnd_Flaky_Perbuild":                     true,
		"WiFi_EndtoEnd_Perbuild":                           true,
		"audio":                                            true,
		"audioKernelnext":                                  true,
		"bluetooth_floss_perbuild_preverification":         true,
		"bluetooth_perbuild_preverification":               true,
		"cellular_callbox":                                 true,
		"fingerprint-mcu-dragonclaw":                       true,
		"kernel_wifi__functional__perbuild__wificell_perbuild__wifi_matfunc__normal__tfc":                true,
		"kernel_wifi__functional__perbuild__wificell_perbuild__wifi_matfunc__persistence__tfc":           true,
		"kernel_wifi__functional__perbuild__wificell_perbuild__wifi_matfunc__suspend__tfc":               true,
		"kernel_wifi__functional__perbuild__wificell_perbuild__wifi_matfunc_flaky__normal__tfc":          true,
		"kernel_wifi__functional__perbuild__wificell_perbuild__wifi_matfunc_flaky__persistence__tfc":     true,
		"kernel_wifi__functional__perbuild__wificell_perbuild__wifi_matfunc_flaky__suspend__tfc":         true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap__wifi_matfunc__normal__tfc":                true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap__wifi_matfunc__persistence__tfc":           true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap__wifi_matfunc__suspend__tfc":               true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap__wifi_matfunc_flaky__normal__tfc":          true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap__wifi_matfunc_flaky__persistence__tfc":     true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap__wifi_matfunc_flaky__suspend__tfc":         true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap_nuc__wifi_matfunc__normal__tfc":            true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap_nuc__wifi_matfunc__persistence__tfc":       true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap_nuc__wifi_matfunc__suspend__tfc":           true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap_nuc__wifi_matfunc_flaky__normal__tfc":      true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap_nuc__wifi_matfunc_flaky__persistence__tfc": true,
		"kernel_wifi__functional__perbuild__wificell_proto_ap_nuc__wifi_matfunc_flaky__suspend__tfc":     true,
		"kernel_wifi__groamer__perbuild__groamer__wifi_atten_perf__tauto":                                true,
		"kernel_wifi__groamer__perbuild__groamer_proto__wifi_atten_perf__tauto":                          true,
		"kernel_wifi__performance__perbuild__wificell__wifi_perf_openwrt__tfc":                           true,
		"kernel_wifi__performance__perbuild__wificell__wifi_perf_openwrt_flaky__tfc":                     true,
		"kernel_wifi__performance__perbuild__wificell_perf__wifi_perf__tauto":                            true,
		"kernel_wifi__performance__perbuild__wificell_perf__wifi_perf_flaky__tauto":                      true,
		"kernel_wifi__performance__perbuild__wificell_proto_ap__wifi_perf_openwrt__tauto":                true,
		"kernel_wifi__performance__perbuild__wificell_proto_ap_nuc__wifi_perf_openwrt_flaky__tauto":      true,
	}
)

// filterConfigs iterates through the triggered SuSch Configs and scrubs out all
// configs which are not on the allowlist.
//
// TODO(b/319273876): Remove slow migration logic upon completion of
// transition.
func filterConfigs(buildPackages []*builds.BuildPackage) []*builds.BuildPackage {
	filteredBuilds := []*builds.BuildPackage{}

	hadAllowedConfig := false
	for _, build := range buildPackages {
		// Copy the build by value so that we can clear the requests field.
		tempBuild := *build
		tempBuild.Requests = []*builds.ConfigDetails{}

		// Iterate through the requests and only add requests to the temp build
		// if their SuSch config is on the allowlist.
		for _, request := range build.Requests {
			if _, ok := allowedConfigs[request.Config.Name]; ok {
				tempBuild.Requests = append(tempBuild.Requests, request)
				hadAllowedConfig = true
			}
		}

		if hadAllowedConfig {
			filteredBuilds = append(filteredBuilds, &tempBuild)
		}
	}

	return filteredBuilds
}
