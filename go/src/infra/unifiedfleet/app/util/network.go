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
	reservedNum := reserveFirst + reserveLast
	if freeStartIP != "" {
		ipInt, err := IPv4StrToInt(freeStartIP)
		if err != nil {
			return nil, 0, "", "", 0, errors.Reason("invalid free start IP %q for vlan %s", freeStartIP, vlanName).Err()
		}
		reservedNum = reservedNum + int(ipInt) - int(freeStartIPInt)
		freeStartIPInt = ipInt

	} else {
		freeStartIP = IPv4IntToStr(uint32(startIP + reserveFirst))
	}
	if freeEndIP != "" {
		ipInt, err := IPv4StrToInt(freeEndIP)
		if err != nil {
			return nil, 0, "", "", 0, errors.Reason("invalid free end IP %q for vlan %s", freeEndIP, vlanName).Err()
		}
		reservedNum = reservedNum + int(freeEndIPInt) - int(ipInt)
		freeEndIPInt = ipInt
	} else {
		freeEndIP = IPv4IntToStr(startIP + uint32(length-reserveLast-1))
	}
	ips := make([]*ufspb.IP, 0, length)
	endIP := startIP + uint32(length)
	for ; startIP < endIP; startIP++ {
		if startIP < freeStartIPInt || startIP > freeEndIPInt {
			ips = append(ips, &ufspb.IP{
				Id:      GetIPName(vlanName, Int64ToStr(int64(startIP))),
				Ipv4:    startIP,
				Ipv4Str: IPv4IntToStr(startIP),
				Vlan:    vlanName,
				Reserve: true,
			})
		} else {
			ips = append(ips, &ufspb.IP{
				Id:      GetIPName(vlanName, Int64ToStr(int64(startIP))),
				Ipv4:    startIP,
				Ipv4Str: IPv4IntToStr(startIP),
				Vlan:    vlanName,
			})
		}
	}
	return ips, length, freeStartIP, freeEndIP, reservedNum, nil
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
