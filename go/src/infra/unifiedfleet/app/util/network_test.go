// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"net"
	"testing"
	"testing/quick"

	"go.chromium.org/luci/common/testing/typed"

	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/util/iputil"
)

func TestParseVlan(t *testing.T) {
	Convey("ParseVlan - happy path", t, func() {
		ips, l, freeStartIP, freeEndIP, reservedNum, err := ParseVlan("fake_vlan", "192.168.40.0/22", "", "")
		So(err, ShouldBeNil)
		So(l, ShouldEqual, 1024)
		So(ips, ShouldHaveLength, 1024)
		for i, ip := range ips {
			if i >= 0 && i < 11 {
				So(ip.GetReserve(), ShouldBeTrue)
			} else if i >= 1023 {
				So(ip.GetReserve(), ShouldBeTrue)
			} else {
				So(ip.GetReserve(), ShouldBeFalse)
			}
		}
		So(freeStartIP, ShouldEqual, "192.168.40.11")
		So(freeEndIP, ShouldEqual, "192.168.43.254")
		// 12 = util.reserveFirst (11) + util.reserveLast (1)
		So(reservedNum, ShouldEqual, 12)
	})

	Convey("ParseVlan - happy path with free start/end ip", t, func() {
		ips, l, freeStartIP, freeEndIP, reservedNum, err := ParseVlan("fake_vlan", "192.168.40.0/22", "192.168.40.100", "192.168.40.200")
		So(err, ShouldBeNil)
		So(l, ShouldEqual, 1024)
		So(ips, ShouldHaveLength, 1024)
		So(freeStartIP, ShouldEqual, "192.168.40.100")
		So(freeEndIP, ShouldEqual, "192.168.40.200")
		// 2 ^ (32-22) - 101 (101 IPs available between 192.168.40.100 & 192.168.40.200 )
		expectedReservedIPs := 923
		So(reservedNum, ShouldEqual, expectedReservedIPs)
		for i, ip := range ips {
			if i < 100 {
				So(ip.GetReserve(), ShouldBeTrue)
			} else if i > 200 {
				So(ip.GetReserve(), ShouldBeTrue)
			} else {
				So(ip.GetReserve(), ShouldBeFalse)
			}
		}
	})
}

type parseVlanOutput = struct {
	ips         []*ufspb.IP
	length      int
	freeStartIP string
	freeEndIP   string
	reservedNum int
}

// TestParseVlanTableTest tests the edge cases of parsing vlans.
func TestParseVlanTableTest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		vlanName    string
		cidr        string
		freeStartIP string
		freeEndIP   string
		want        parseVlanOutput
		ok          bool
	}{
		{
			name:        "ipv4 happy path with 1 free IP",
			vlanName:    "fake-vlan",
			cidr:        "127.0.0.0/30",
			freeStartIP: "127.0.0.1",
			freeEndIP:   "127.0.0.1",
			want: parseVlanOutput{
				ips: []*ufspb.IP{
					FormatIP("fake-vlan", "127.0.0.0", true, false),
					FormatIP("fake-vlan", "127.0.0.1", false, false),
					FormatIP("fake-vlan", "127.0.0.2", true, false),
					FormatIP("fake-vlan", "127.0.0.3", true, false),
				},
				length:      4,
				freeStartIP: "127.0.0.1",
				freeEndIP:   "127.0.0.1",
				reservedNum: 3,
			},
			ok: true,
		},
		{
			name:        "ipv4 happy path with 2 free IPs",
			vlanName:    "fake-vlan",
			cidr:        "127.0.0.0/30",
			freeStartIP: "127.0.0.1",
			freeEndIP:   "127.0.0.2",
			want: parseVlanOutput{
				ips: []*ufspb.IP{
					FormatIP("fake-vlan", "127.0.0.0", true, false),
					FormatIP("fake-vlan", "127.0.0.1", false, false),
					FormatIP("fake-vlan", "127.0.0.2", false, false),
					FormatIP("fake-vlan", "127.0.0.3", true, false),
				},
				length:      4,
				freeStartIP: "127.0.0.1",
				freeEndIP:   "127.0.0.2",
				reservedNum: 2,
			},
			ok: true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error
			got := parseVlanOutput{}

			got.ips, got.length, got.freeStartIP, got.freeEndIP, got.reservedNum, err = ParseVlan(tt.vlanName, tt.cidr, tt.freeStartIP, tt.freeEndIP)

			if diff := typed.Diff(tt.want, got, cmp.AllowUnexported(parseVlanOutput{})); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
			switch {
			case err == nil && !tt.ok:
				t.Error("error is unexpectedly nil")
			case err != nil && tt.ok:
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestParseMac(t *testing.T) {
	Convey("ParseMac - happy path", t, func() {
		mac, err := ParseMac("12:34:56:78:90:ab")
		So(err, ShouldBeNil)
		So(mac, ShouldEqual, "12:34:56:78:90:ab")
	})

	Convey("ParseMac - happy path without colon separators", t, func() {
		mac, err := ParseMac("1234567890ab")
		So(err, ShouldBeNil)
		So(mac, ShouldEqual, "12:34:56:78:90:ab")
	})

	Convey("ParseMac - invalid characters", t, func() {
		invalidMacs := []string{
			"1234567890,b",
			"hello world",
			"123455678901234567890",
		}
		for _, userMac := range invalidMacs {
			mac, err := ParseMac(userMac)
			So(err, ShouldNotBeNil)
			So(mac, ShouldBeEmpty)
		}
	})
}

func TestFormatMac(t *testing.T) {
	Convey("formatMac - happy path with colon separators", t, func() {
		So(formatMac("12:34:56:78:90:ab"), ShouldEqual, "12:34:56:78:90:ab")
	})

	Convey("formatMac - happy path without colon separators", t, func() {
		So(formatMac("1234567890ab"), ShouldEqual, "12:34:56:78:90:ab")
	})

	Convey("formatMac - odd length", t, func() {
		So(formatMac("1234567890abcde"), ShouldEqual, "12:34:56:78:90:ab:cd:e")
	})
}

// TestUint32Iter tests that we, by iterating, add the correct number of things to an array.
func TestUint32Iter(t *testing.T) {
	Convey("test uint32 iteration", t, func() {
		var data []uint32

		err := Uint32Iter(0, 1000, func(x uint32) error {
			data = append(data, x)
			return nil
		})

		So(err, ShouldBeNil)
		So(len(data), ShouldEqual, 1001)
	})
}

// TestMakeIPv4sInVlan tests the helper function makeIPv4sInVlan.
func TestMakeIPv4sInVlan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		vlanName    string
		startIP     net.IP
		length      int
		freeStartIP net.IP
		freeEndIP   net.IP
		want        []*ufspb.IP
		ok          bool
	}{
		{
			name:        "one real ip",
			vlanName:    "fake-vlan",
			startIP:     iputil.MustParseIP("127.0.0.0"),
			length:      2,
			freeStartIP: iputil.MustParseIP("127.0.0.1"),
			freeEndIP:   iputil.MustParseIP("127.0.0.1"),
			want: []*ufspb.IP{
				FormatIP("fake-vlan", "127.0.0.0", true, false),
				FormatIP("fake-vlan", "127.0.0.1", false, false),
			},
			ok: true,
		},
		{
			name:        "two real ips",
			vlanName:    "fake-vlan",
			startIP:     iputil.MustParseIP("127.0.0.0"),
			length:      2,
			freeStartIP: iputil.MustParseIP("127.0.0.1"),
			freeEndIP:   iputil.MustParseIP("127.0.0.2"),
			want: []*ufspb.IP{
				FormatIP("fake-vlan", "127.0.0.0", true, false),
				FormatIP("fake-vlan", "127.0.0.1", false, false),
				FormatIP("fake-vlan", "127.0.0.2", false, false),
			},
			ok: true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := makeIPv4sInVlan(tt.vlanName, tt.startIP, tt.length, tt.freeStartIP, tt.freeEndIP)

			if diff := typed.Diff(got, tt.want); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
			if diff := typed.Diff(err == nil, tt.ok); diff != "" {
				if err != nil {
					t.Error(err)
				}
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestMakeReservedIPv4sInVlan tests making a range of IPv4s.
func TestMakeReservedIPv4sInVlan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		vlanName string
		begin    net.IP
		end      net.IP
		maximum  int
		want     []*ufspb.IP
		ok       bool
	}{
		{
			name:     "singleton",
			vlanName: "fake-vlan",
			begin:    iputil.MustParseIP("0.0.0.0"),
			end:      iputil.MustParseIP("0.0.0.0"),
			maximum:  1,
			want:     []*ufspb.IP{FormatIP("fake-vlan", "0.0.0.0", true, false)},
			ok:       true,
		},
		{
			name:     "too short",
			vlanName: "fake-vlan",
			begin:    iputil.MustParseIP("0.0.0.0"),
			end:      iputil.MustParseIP("0.0.0.0"),
			maximum:  0,
			want:     nil,
			ok:       false,
		},
		{
			name:     "beginning too big",
			vlanName: "fake-vlan",
			begin:    iputil.MustParseIP("0.0.0.10"),
			end:      iputil.MustParseIP("0.0.0.0"),
			maximum:  100,
			want:     nil,
			ok:       false,
		},
		{
			name:     "array too long",
			vlanName: "fake-vlan",
			begin:    iputil.MustParseIP("0.0.0.0"),
			end:      iputil.MustParseIP("0.0.255.255"),
			maximum:  100000,
			want:     nil,
			ok:       false,
		},
		{
			name:     "1 2 3 4",
			vlanName: "fake-vlan",
			begin:    iputil.MustParseIP("0.0.0.1"),
			end:      iputil.MustParseIP("0.0.0.4"),
			maximum:  4,
			want: []*ufspb.IP{
				FormatIP("fake-vlan", "0.0.0.1", true, false),
				FormatIP("fake-vlan", "0.0.0.2", true, false),
				FormatIP("fake-vlan", "0.0.0.3", true, false),
				FormatIP("fake-vlan", "0.0.0.4", true, false),
			},
			ok: true,
		},
		{
			name:     "1 2 3 4 ipv6",
			vlanName: "fake-vlan",
			begin:    iputil.MustParseIP("aaaa::1"),
			end:      iputil.MustParseIP("aaaa::4"),
			maximum:  4,
			want: []*ufspb.IP{
				FormatIP("fake-vlan", "aaaa::1", true, false),
				FormatIP("fake-vlan", "aaaa::2", true, false),
				FormatIP("fake-vlan", "aaaa::3", true, false),
				FormatIP("fake-vlan", "aaaa::4", true, false),
			},
			ok: true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := makeReservedIPsInVlan(tt.vlanName, tt.begin, tt.end, tt.maximum)

			if diff := typed.Diff(got, tt.want); diff != "" {
				t.Errorf("unexpected error (-want +got): %s", diff)
			}
			switch {
			case err == nil && !tt.ok:
				t.Error("error unexpectedly nil")
			case err != nil && tt.ok:
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestIPv4Diff tests diffing two IPv4 addresses.
func TestIPv4Diff(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		startIPv4 string
		endIPv4   string
		result    uint64
		ok        bool
	}{
		{
			name:      "smoke test",
			startIPv4: "",
			endIPv4:   "",
			result:    0,
			ok:        false,
		},
		{
			name:      "start > end",
			startIPv4: "127.0.0.2",
			endIPv4:   "127.0.0.1",
			result:    0,
			ok:        false,
		},
		{
			name:      "happy path",
			startIPv4: "127.0.0.1",
			endIPv4:   "127.0.0.2",
			result:    1,
			ok:        true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dist, err := ipv4Diff(tt.startIPv4, tt.endIPv4)
			switch {
			case err == nil && !tt.ok:
				t.Error("err unexpectly nil")
			case err != nil && tt.ok:
				t.Errorf("unexpected error: %s", err)
			}
			if diff := cmp.Diff(dist, tt.result); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestUint32ToIPRoundTrip checks that Uint32ToIP and IPv4ToUint32 are inveses of each other.
func TestUint32ToIPRoundTrip(t *testing.T) {
	t.Parallel()

	checker := func(addr uint32) bool {
		ipv4 := uint32ToIP(addr)
		roundTripped, err := IPv4ToUint32(ipv4)
		if err != nil {
			panic(err)
		}
		return roundTripped == addr
	}

	if err := quick.Check(checker, nil); err != nil {
		t.Error(err)
	}
}

// TestFormatIP tests that format IP stuff.
func TestFormatIP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		vlanName  string
		ipAddress string
		reserve   bool
		occupied  bool
		want      *ufspb.IP
	}{
		{
			name:      "invalid IP",
			vlanName:  "fake-vlan",
			ipAddress: "not a valid IP address",
			reserve:   false,
			occupied:  false,
			want:      nil,
		},
		{
			name:      "happy path IPv4",
			vlanName:  "fake-vlan",
			ipAddress: "127.0.0.1",
			reserve:   false,
			occupied:  false,
			want: &ufspb.IP{
				Vlan:    "fake-vlan",
				Id:      "fake-vlan/2130706433",
				Ipv4:    makeIPv4Uint32(127, 0, 0, 1),
				Ipv4Str: "127.0.0.1",
				Ipv6:    iputil.MustParseIP("127.0.0.1"),
			},
		},
		{
			name:      "happy path IPv6",
			vlanName:  "fake-vlan",
			ipAddress: "1234:1234:1234:1234:aaaa:aaaa:aaaa:aaaa",
			reserve:   false,
			occupied:  false,
			want: &ufspb.IP{
				Vlan:    "fake-vlan",
				Id:      "fake-vlan/0x1234123412341234aaaaaaaaaaaaaaaa",
				Ipv4:    0,
				Ipv4Str: "",
				Ipv6:    iputil.MustParseIP("1234:1234:1234:1234:aaaa:aaaa:aaaa:aaaa"),
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FormatIP(tt.vlanName, tt.ipAddress, tt.reserve, tt.occupied)

			if diff := typed.Diff(got, tt.want); diff != "" {
				t.Errorf("unexpected diff (-want+got): %s", diff)
			}
		})
	}
}
