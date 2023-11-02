// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"go.chromium.org/luci/common/errors"

	ufspb "infra/unifiedfleet/api/v1/models"
)

// The first ip is the subnet ID
// The next first 10 ips are reserved
const reserveFirst = 11

// The last ip is broadcast address.
const reserveLast = 1

// MaxPreallocatedVlanSize is the maximum number of preallocated vlan addresses.
const MaxPreallocatedVlanSize = 2000

// StringifyIP stringifies an IP. The standard library makes the interesting
// choice of mapping an empty IP address object to "<nil>" rather than "".
//
// Here we just return an empty string given an empty IP.
func StringifyIP(ip net.IP) string {
	if len(ip) == 0 {
		return ""
	}
	return ip.String()
}

// ParseVlan parses vlan to a list of IPs
//
// vlanName here is a full vlan name, e.g. browser:123
// The first 10 and last 1 ip of this cidr block will be reserved and not returned to users
// for further operations
func ParseVlan(vlanName, cidr, freeStartIP, freeEndIP string) ([]*ufspb.IP, int, string, string, int, error) {
	ip, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, 0, "", "", 0, errors.Reason("invalid CIDR block %q for vlan %s", cidr, vlanName).Err()
	}
	ipv4 := ip.Mask(subnet.Mask).To4()
	if ipv4 == nil {
		return nil, 0, "", "", 0, errors.Reason("invalid IPv4 CIDR block %q for vlan %s", cidr, vlanName).Err()
	}
	ones, _ := subnet.Mask.Size()
	length := 1 << uint32(32-ones)
	startIP := binary.BigEndian.Uint32(ipv4)
	freeStartIPInt := startIP + reserveFirst
	freeEndIPInt := startIP + uint32(length-reserveLast-1)
	if freeStartIP != "" {
		ipInt, err := IPv4StrToInt(freeStartIP)
		if err != nil {
			return nil, 0, "", "", 0, errors.Reason("invalid free start IP %q for vlan %s", freeStartIP, vlanName).Err()
		}
		freeStartIPInt = ipInt

	} else {
		freeStartIP = IPv4IntToStr(uint32(startIP + reserveFirst))
	}
	if freeEndIP != "" {
		ipInt, err := IPv4StrToInt(freeEndIP)
		if err != nil {
			return nil, 0, "", "", 0, errors.Reason("invalid free end IP %q for vlan %s", freeEndIP, vlanName).Err()
		}
		freeEndIPInt = ipInt
	} else {
		freeEndIP = IPv4IntToStr(startIP + uint32(length-reserveLast-1))
	}
	used, err := IPv4Diff(freeStartIP, freeEndIP)
	if err != nil {
		return nil, 0, "", "", 0, err
	}
	reservedNum := length - int(used) - 1
	ips := makeIPv4sInVlan(vlanName, startIP, length, freeStartIPInt, freeEndIPInt)
	return ips, length, freeStartIP, freeEndIP, reservedNum, nil
}

// makeIPv4 makes a ufs IP object.
func makeIPv4(vlanName string, ipv4 uint32, reserved bool) *ufspb.IP {
	return &ufspb.IP{
		Id:      GetIPName(vlanName, Int64ToStr(int64(ipv4))),
		Ipv4:    ipv4,
		Ipv4Str: IPv4IntToStr(ipv4),
		Reserve: reserved,
		Vlan:    vlanName,
	}
}

// makeIPv4sInVlan creates the IP objects in a Vlan that are intended to be created in datastore later.
func makeIPv4sInVlan(vlanName string, startIP uint32, length int, freeStartIPInt uint32, freeEndIPInt uint32) []*ufspb.IP {
	endIP := startIP + uint32(length)
	ips := make([]*ufspb.IP, 0, length)
	for ; startIP < endIP; startIP++ {
		reserved := startIP < freeStartIPInt || startIP > freeEndIPInt
		ips = append(ips, makeIPv4(vlanName, startIP, reserved))
	}
	return ips
}

// makeIPv4Uint32 makes an IPv4 as a uint32 using notation that looks conventional.
//
// makeIPv4Uint32(127, 0, 0, 1) === 127.0.0.1
func makeIPv4Uint32(a, b, c, d int) uint32 {
	return uint32(a)*256*256*256 + uint32(b)*256*256 + uint32(c)*256 + uint32(d)
}

// FormatIP initialize an IP object
func FormatIP(vlanName, ipAddress string, reserve, occupied bool) *ufspb.IP {
	ipv4, err := IPv4StrToInt(ipAddress)
	if err != nil {
		return nil
	}
	return &ufspb.IP{
		Id:       GetIPName(vlanName, Int64ToStr(int64(ipv4))),
		Ipv4:     ipv4,
		Ipv4Str:  ipAddress,
		Vlan:     vlanName,
		Occupied: occupied,
		Reserve:  reserve,
	}
}

// IPv4Diff takes the difference between a startIPv4 and an endIPv4.
//
// It returns an error if and only if A) at least one argument is an invalid IP address or B) the end strictly precedes the start
func IPv4Diff(startIPv4 string, endIPv4 string) (uint64, error) {
	start, err := IPv4StrToInt(startIPv4)
	if err != nil {
		return 0, errors.Annotate(err, "diffing IPs %q and %q", startIPv4, endIPv4).Err()
	}
	end, err := IPv4StrToInt(endIPv4)
	if err != nil {
		return 0, errors.Annotate(err, "diffing IPs %q and %q", startIPv4, endIPv4).Err()
	}
	if start > end {
		return 0, errors.Reason("end IP %q precedes start IP %q", end, start).Err()
	}
	return uint64(end) - uint64(start), nil
}

// Uint32Iter runs a command over a range of Uint32's. Useful for iterating over IP addresses.
func Uint32Iter(start uint32, end uint32, f func(uint32) error) error {
	for num := start; num <= end; num++ {
		if err := f(num); err != nil {
			return err
		}
	}
	return nil
}

// IPv4StrToInt returns an uint32 address from the given ip address string.
func IPv4StrToInt(ipAddress string) (uint32, error) {
	ip := net.ParseIP(ipAddress)
	if ip != nil {
		ip = ip.To4()
	}
	if ip == nil {
		return 0, errors.Reason("invalid IPv4 address %q", ipAddress).Err()
	}
	return binary.BigEndian.Uint32(ip), nil
}

// IPv4IntToStr returns a string ip address
func IPv4IntToStr(ipAddress uint32) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipAddress)
	return ip.String()
}

// parseCidrBlock returns a tuple of (cidr_block, capacity of this block)
func parseCidrBlock(subnet, mask string) (string, int) {
	maskIP := net.ParseIP(mask)
	maskAddr := maskIP.To4()
	ones, sz := net.IPv4Mask(maskAddr[0], maskAddr[1], maskAddr[2], maskAddr[3]).Size()
	return fmt.Sprintf("%s/%d", subnet, ones), 1 << uint32(sz-ones)
}

// ParseMac returns a valid mac address after parsing user input.
func ParseMac(userMac string) (string, error) {
	newUserMac := formatMac(userMac)
	m, err := net.ParseMAC(newUserMac)
	if err != nil || len(m) != 6 {
		return "", errors.Reason("invalid mac address %q (before parsing %q)", newUserMac, userMac).Err()
	}
	bytes := make([]byte, 8)
	copy(bytes[2:], m)
	mac := make(net.HardwareAddr, 8)
	binary.BigEndian.PutUint64(mac, binary.BigEndian.Uint64(bytes))
	return mac[2:].String(), nil
}

func formatMac(userMac string) string {
	if strings.Contains(userMac, ":") {
		return userMac
	}

	var newMac string
	for i := 0; ; i += 2 {
		if i+2 > len(userMac)-1 {
			newMac += userMac[i:]
			break
		}
		newMac += userMac[i:i+2] + ":"
	}
	return newMac
}

// IsMacFormatValid check if the given mac address is in valid format
func IsMacFormatValid(userMac string) error {
	newUserMac := formatMac(userMac)
	m, err := net.ParseMAC(newUserMac)
	if err != nil || len(m) != 6 {
		return errors.Reason("Invalid mac address %q (before parsing %q)", newUserMac, userMac).Err()
	}
	return nil
}
