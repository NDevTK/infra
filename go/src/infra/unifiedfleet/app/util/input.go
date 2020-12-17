// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	descriptorpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/gae/service/info"
	"google.golang.org/genproto/googleapis/api/annotations"

	ufspb "infra/unifiedfleet/api/v1/models"
)

const (
	// MachineCollection refers to the prefix of the corresponding resource.
	MachineCollection string = "machines"
	// RackCollection refers to the prefix of the corresponding resource.
	RackCollection string = "racks"
	// VMCollection refers to the prefix of the corresponding resource.
	VMCollection string = "vms"
	// ChromePlatformCollection refers to the prefix of the corresponding resource.
	ChromePlatformCollection string = "chromeplatforms"
	// MachineLSECollection refers to the prefix of the corresponding resource.
	MachineLSECollection string = "machineLSEs"
	// HostCollection refers to the prefix of the corresponding resource.
	HostCollection string = "hosts"
	// RackLSECollection refers to the prefix of the corresponding resource.
	RackLSECollection string = "rackLSEs"
	// NicCollection refers to the prefix of the corresponding resource.
	NicCollection string = "nics"
	// KVMCollection refers to the prefix of the corresponding resource.
	KVMCollection string = "kvms"
	// RPMCollection refers to the prefix of the corresponding resource.
	RPMCollection string = "rpms"
	// DracCollection refers to the prefix of the corresponding resource.
	DracCollection string = "dracs"
	// SwitchCollection refers to the prefix of the corresponding resource.
	SwitchCollection string = "switches"
	// VlanCollection refers to the prefix of the corresponding resource.
	VlanCollection string = "vlans"
	// MachineLSEPrototypeCollection refers to the prefix of the corresponding resource.
	MachineLSEPrototypeCollection string = "machineLSEPrototypes"
	// RackLSEPrototypeCollection refers to the prefix of the corresponding resource.
	RackLSEPrototypeCollection string = "rackLSEPrototypes"
	// DHCPCollection refers to the prefix of the dhcp config id in change history
	DHCPCollection string = "dhcps"
	// IPCollection refers to the prefix of the ip id in change history
	IPCollection string = "ips"
	// StateCollection refers to the prefix of the states id in change history
	StateCollection string = "states"

	// DefaultImporter refers to the user of the cron job importer
	DefaultImporter string = "crimson-importer"

	defaultPageSize int32 = 100
	// MaxPageSize maximum page size for list operations
	MaxPageSize int32 = 1000
)

var collectionsRe = regexp.MustCompile(`\/{[a-zA-Z0-9]*}$`)

// Filter names for indexed properties in datastore for different entities
var (
	ZoneFilterName              string = "zone"
	RackFilterName              string = "rack"
	MachineFilterName           string = "machine"
	HostFilterName              string = "host"
	NicFilterName               string = "nic"
	DracFilterName              string = "drac"
	KVMFilterName               string = "kvm"
	KVMPortFilterName           string = "kvmport"
	MacAddressFilterName        string = "mac"
	RPMFilterName               string = "rpm"
	SwitchFilterName            string = "switch"
	SwitchPortFilterName        string = "switchport"
	ServoFilterName             string = "servo"
	ServoTypeFilterName         string = "servotype"
	TagFilterName               string = "tag"
	ChromePlatformFilterName    string = "platform"
	MachinePrototypeFilterName  string = "machineprototype"
	RackPrototypeFilterName     string = "rackprototype"
	VlanFilterName              string = "vlan"
	StateFilterName             string = "state"
	IPV4FilterName              string = "ipv4"
	IPV4StringFilterName        string = "ipv4str"
	OccupiedFilterName          string = "occupied"
	ManufacturerFilterName      string = "man"
	FreeVMFilterName            string = "free"
	ResourceTypeFilterName      string = "resourcetype"
	OSVersionFilterName         string = "osversion"
	OSFilterName                string = "os"
	VirtualDatacenterFilterName string = "vdc"
	ModelFilterName             string = "model"
	BuildTargetFilterName       string = "target"
	DeviceTypeFilterName        string = "devicetype"
	PhaseFilterName             string = "phase"
)

const separator string = "/"

// Namespace namespace to be set by clients in context metadata
// This will be used to set the actual datastore namespace in the context
var (
	// OSNamespace os namespace to be set in client context metadata. OS data is stored in os namespace in the datastore.
	OSNamespace = "os"
	// BrowserNamespace browser namespace to be set in client context metadata. Browser data is stored in default namespace in the datastore.
	BrowserNamespace = "browser"
	//Namespace key in the incoming context metadata
	Namespace = "namespace"
)

// ClientToDatastoreNamespace refers a map between client namespace(set in context metadata) to actual datastore namespace
var ClientToDatastoreNamespace = map[string]string{
	BrowserNamespace: "",          // browser data is stored in default namespace
	OSNamespace:      OSNamespace, // os data in os namespace
}

// ValidClientNamespaceStr returns a valid str list for client namespace(set in incoming context metadata) strings.
func ValidClientNamespaceStr() []string {
	ks := make([]string, 0, len(ClientToDatastoreNamespace))
	for k := range ClientToDatastoreNamespace {
		ks = append(ks, k)
	}
	return ks
}

// IsClientNamespace checks if a string refers to a valid client namespace.
func IsClientNamespace(namespace string) bool {
	_, ok := ClientToDatastoreNamespace[namespace]
	return ok
}

// SetupDatastoreNamespace sets the datastore namespace in the context to access the correct namespace in the datastore
func SetupDatastoreNamespace(ctx context.Context, namespace string) (context.Context, error) {
	return info.Namespace(ctx, namespace)
}

// GetPageSize gets the correct page size for List pagination
func GetPageSize(pageSize int32) int32 {
	switch {
	case pageSize == 0:
		return defaultPageSize
	case pageSize > MaxPageSize:
		return MaxPageSize
	default:
		return pageSize
	}
}

// GetResourcePrefix gets the resource prefix given to the proto message.
//
// Returns the resource prefix for a given proto message.
// See also: https://blog.golang.org/protobuf-apiv2
func GetResourcePrefix(message proto.Message) (string, error) {
	m := proto.MessageReflect(message)
	x, ok := m.Descriptor().Options().(*descriptorpb.MessageOptions)
	if !ok {
		return "", errors.Reason("Unable to read Message Options").Err()
	}
	y, _ := proto.GetExtension(x, annotations.E_Resource)
	z, ok := y.(*annotations.ResourceDescriptor)
	if !ok {
		return "", errors.Reason("Resource descriptor not found in proto message").Err()
	}
	prefix := collectionsRe.ReplaceAllString(z.Pattern[0], "")
	return prefix, nil
}

// FormatInputNames formats a given array of resource names
func FormatInputNames(names []string) []string {
	var res []string
	for _, n := range names {
		if n != "" {
			res = append(res, RemovePrefix(n))
		}
	}
	return res
}

// FormatDHCPHostname formats a name which will be a dhcp host
func FormatDHCPHostname(old string) string {
	return strings.ToLower(old)
}

// FormatDHCPHostnames formats a given array of resource names which could be used as dhcp hostnames
func FormatDHCPHostnames(names []string) []string {
	for i, n := range names {
		names[i] = FormatDHCPHostname(n)
	}
	return names
}

// RemovePrefix extracts string appearing after a "/"
func RemovePrefix(name string) string {
	// Get substring after a string.
	name = strings.TrimSpace(name)
	pos := strings.Index(name, separator)
	if pos == -1 {
		return name
	}
	adjustedPos := pos + len(separator)
	if adjustedPos >= len(name) {
		return name
	}
	return name[adjustedPos:]
}

// AddPrefix adds the prefix for a given resource name
func AddPrefix(collection, entity string) string {
	return fmt.Sprintf("%s%s%s", collection, separator, entity)
}

// GetPrefix returns the prefix for a resource name
func GetPrefix(resourceName string) string {
	s := strings.Split(strings.TrimSpace(resourceName), separator)
	if len(s) < 1 {
		return ""
	}
	return s[0]
}

// GetRackHostname returns a rack host name.
func GetRackHostname(rackName string) string {
	return fmt.Sprintf("%s-host", rackName)
}

// FormatResourceName formats the resource name
func FormatResourceName(old string) string {
	str := strings.Replace(old, " ", "_", -1)
	return strings.Replace(str, ",", "_", -1)
}

// StrToUFSState refers a map between a string to a UFS defined state map.
var StrToUFSState = map[string]string{
	"registered":           "STATE_REGISTERED",
	"deployed_pre_serving": "STATE_DEPLOYED_PRE_SERVING",
	"deployed_testing":     "STATE_DEPLOYED_TESTING",
	"serving":              "STATE_SERVING",
	"needs_reset":          "STATE_NEEDS_RESET",
	"needs_repair":         "STATE_NEEDS_REPAIR",
	"repair_failed":        "STATE_REPAIR_FAILED",
	"disabled":             "STATE_DISABLED",
	"reserved":             "STATE_RESERVED",
	"decommissioned":       "STATE_DECOMMISSIONED",
	"deploying":            "STATE_DEPLOYING",
	"ready":                "STATE_READY",
}

// StateToDescription refers a map between a State to its description.
var StateToDescription = map[string]string{
	"registered":           "Needs deploy",
	"deployed_pre_serving": "Deployed but not placed in prod",
	"deployed_testing":     "Deployed to the prod, but for testing",
	"serving":              "Deployed to the prod, serving",
	"needs_reset":          "Deployed to the prod, but required cleanup and verify",
	"needs_repair":         "Deployed to the prod, needs repair",
	"repair_failed":        "Deployed to the prod, failed to be repaired in previous step and requires new repair attempt",
	"disabled":             "Deployed to the prod, but disabled",
	"reserved":             "Deployed to the prod, but reserved (e.g. locked)",
	"decommissioned":       "Decommissioned from the prod, but still lives in UFS record",
	"deploying":            "Deploying the resource with required configs just before it is READY",
	"ready":                "Resource is ready for use or free to use",
}

// IsUFSState checks if a string refers to a valid UFS state.
func IsUFSState(state string) bool {
	_, ok := StrToUFSState[state]
	return ok
}

// ValidStateStr returns a valid str list for state strings.
func ValidStateStr() []string {
	ks := make([]string, 0, len(StrToUFSState))
	for k := range StrToUFSState {
		ks = append(ks, k)
	}
	return ks
}

// RemoveStatePrefix removes the "state_" prefix from the string
func RemoveStatePrefix(state string) string {
	state = strings.ToLower(state)
	if idx := strings.Index(state, "state_"); idx != -1 {
		state = state[idx+len("state_"):]
	}
	return state
}

// ToUFSState converts state string to a UFS state enum.
func ToUFSState(state string) ufspb.State {
	state = RemoveStatePrefix(state)
	v, ok := StrToUFSState[state]
	if !ok {
		return ufspb.State_STATE_UNSPECIFIED
	}
	return ufspb.State(ufspb.State_value[v])
}

// StrToUFSZone refers a map between a string to a UFS defined map.
var StrToUFSZone = map[string]string{
	"atlanta":           "ZONE_ATLANTA",
	"chromeos1":         "ZONE_CHROMEOS1",
	"chromeos4":         "ZONE_CHROMEOS4",
	"chromeos6":         "ZONE_CHROMEOS6",
	"chromeos2":         "ZONE_CHROMEOS2",
	"chromeos3":         "ZONE_CHROMEOS3",
	"chromeos5":         "ZONE_CHROMEOS5",
	"chromeos7":         "ZONE_CHROMEOS7",
	"chromeos15":        "ZONE_CHROMEOS15",
	"atl97":             "ZONE_ATL97",
	"iad97":             "ZONE_IAD97",
	"mtv96":             "ZONE_MTV96",
	"mtv97":             "ZONE_MTV97",
	"fuchsia":           "ZONE_FUCHSIA",
	"unspecified":       "ZONE_UNSPECIFIED",
	"cros_googler_desk": "ZONE_CROS_GOOGLER_DESK",
}

// IsUFSZone checks if a string refers to a valid UFS zone.
func IsUFSZone(zone string) bool {
	_, ok := StrToUFSZone[zone]
	return ok
}

// IsAssetType checks if a strings is a valid asset type
func IsAssetType(assetType string) bool {
	for _, x := range ValidAssetTypeStr() {
		if x == assetType {
			return true
		}
	}
	return false
}

// ToAssetType returns an AssetType object corresponding to string
func ToAssetType(assetType string) ufspb.AssetType {
	aType := strings.ReplaceAll(assetType, "AssetType_", "")
	for k, v := range ufspb.AssetType_value {
		if strings.EqualFold(k, aType) {
			return ufspb.AssetType(v)
		}
	}
	return ufspb.AssetType_UNDEFINED
}

// ValidAssetTypeStr returns a valid str list for AssetTypes
func ValidAssetTypeStr() []string {
	keys := make([]string, 0, len(ufspb.AssetType_name))
	for k, v := range ufspb.AssetType_name {
		// 0 is UNDEFINED
		if k != 0 {
			keys = append(keys, strings.ToLower(v))
		}
	}
	return keys
}

// ValidZoneStr returns a valid str list for zone strings.
func ValidZoneStr() []string {
	ks := make([]string, 0, len(StrToUFSZone))
	for k := range StrToUFSZone {
		ks = append(ks, k)
	}
	return ks
}

// RemoveZonePrefix removes the "zone_" prefix from the string
func RemoveZonePrefix(zone string) string {
	zone = strings.ToLower(zone)
	if idx := strings.Index(zone, "zone_"); idx != -1 {
		zone = zone[idx+len("zone_"):]
	}
	return zone
}

// ToUFSZone converts zone string to a UFS zone enum.
func ToUFSZone(zone string) ufspb.Zone {
	zone = RemoveZonePrefix(zone)
	v, ok := StrToUFSZone[zone]
	if !ok {
		return ufspb.Zone_ZONE_UNSPECIFIED
	}
	return ufspb.Zone(ufspb.Zone_value[v])
}

// StrToUFSDeviceType refers a map between a string to a UFS defined map.
var StrToUFSDeviceType = map[string]string{
	"chromebook":  "DEVICE_CHROMEBOOK",
	"labstation":  "DEVICE_LABSTATION",
	"servo":       "DEVICE_SERVO",
	"unspecified": "CHROME_OS_DEVICE_TYPE_UNSPECIFIED",
}

// ValidDeviceTypeStr returns a valid str list for devicetype strings.
func ValidDeviceTypeStr() []string {
	ks := make([]string, 0, len(StrToUFSDeviceType))
	for k := range StrToUFSDeviceType {
		ks = append(ks, k)
	}
	return ks
}

// RemoveGivenPrefix removes the prefix from the string
func RemoveGivenPrefix(msg, prefix string) string {
	msg = strings.ToLower(msg)
	if idx := strings.Index(msg, prefix); idx != -1 {
		msg = msg[idx+len(prefix):]
	}
	return msg
}

// ToUFSDeviceType converts devicetype string to a UFS devicetype enum.
func ToUFSDeviceType(devicetype string) ufspb.ChromeOSDeviceType {
	devicetype = RemoveGivenPrefix(devicetype, "device_")
	v, ok := StrToUFSDeviceType[devicetype]
	if !ok {
		return ufspb.ChromeOSDeviceType_CHROME_OS_DEVICE_TYPE_UNSPECIFIED
	}
	return ufspb.ChromeOSDeviceType(ufspb.ChromeOSDeviceType_value[v])
}

// List of regexps for recognizing assets stored with googlers or out of lab.
var googlers = []*regexp.Regexp{
	regexp.MustCompile(`container`),
	regexp.MustCompile(`desk`),
	regexp.MustCompile(`testbed`),
}

// LabToZone converts deprecated Lab type to Zone
func LabToZone(lab string) ufspb.Zone {
	switch oslabRegexp.FindString(lab) {
	case "chromeos1":
		return ufspb.Zone_ZONE_CHROMEOS1
	case "chromeos2":
		return ufspb.Zone_ZONE_CHROMEOS2
	case "chromeos3":
		return ufspb.Zone_ZONE_CHROMEOS3
	case "chromeos4":
		return ufspb.Zone_ZONE_CHROMEOS4
	case "chromeos5":
		return ufspb.Zone_ZONE_CHROMEOS5
	case "chromeos6":
		return ufspb.Zone_ZONE_CHROMEOS6
	case "chromeos7":
		return ufspb.Zone_ZONE_CHROMEOS7
	case "chromeos15":
		return ufspb.Zone_ZONE_CHROMEOS15
	default:
		for _, r := range googlers {
			if r.MatchString(lab) {
				return ufspb.Zone_ZONE_CROS_GOOGLER_DESK
			}
		}
		return ufspb.Zone_ZONE_UNSPECIFIED
	}
}

// ToUFSDept returns the dept name based on zone string.
func ToUFSDept(zone string) string {
	ufsZone := ToUFSZone(zone)
	if IsInBrowserZone(ufsZone.String()) {
		return Browser
	}
	return CrOS
}

// GetStateDescription returns the description for the state
func GetStateDescription(state string) string {
	state = RemoveStatePrefix(state)
	v, ok := StateToDescription[state]
	if !ok {
		return ""
	}
	return v
}

// GetSuffixAfterSeparator extracts the string appearing after the separator
//
// returns the suffix after the first found separator
func GetSuffixAfterSeparator(name, seprator string) string {
	name = strings.TrimSpace(name)
	pos := strings.Index(name, seprator)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(seprator)
	if adjustedPos >= len(name) {
		return ""
	}
	return name[adjustedPos:]
}

// ServoV3HostnameRegex is used to identify servo V3 hosts.
var ServoV3HostnameRegex = regexp.MustCompile(`.*-servo`)
