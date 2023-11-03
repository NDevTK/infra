// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package util

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/testing/protocmp"

	ufspb "infra/unifiedfleet/api/v1/models"
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
		startIP     uint32
		length      int
		freeStartIP uint32
		freeEndIP   uint32
		want        []*ufspb.IP
	}{
		{
			name:        "one real ip",
			vlanName:    "fake-vlan",
			startIP:     makeIPv4Uint32(127, 0, 0, 0),
			length:      2,
			freeStartIP: makeIPv4Uint32(127, 0, 0, 1),
			freeEndIP:   makeIPv4Uint32(127, 0, 0, 2),
			want: []*ufspb.IP{
				FormatIP("fake-vlan", "127.0.0.0", true, false),
				FormatIP("fake-vlan", "127.0.0.1", false, false),
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := makeIPv4sInVlan(tt.vlanName, tt.startIP, tt.length, tt.freeStartIP, tt.freeEndIP)

			if diff := cmp.Diff(got, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

// TestMakeIPv4 tests the simple helper function that makes an IPv4.
func TestMakeIPv4(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		vlanName string
		ipv4     uint32
		reserved bool
		want     *ufspb.IP
	}{
		{
			name:     "simple vlan",
			vlanName: "fake-vlan",
			ipv4:     makeIPv4Uint32(127, 0, 0, 1),
			reserved: true,
			want: &ufspb.IP{
				Id:      GetIPName("fake-vlan", Int64ToStr(int64(makeIPv4Uint32(127, 0, 0, 1)))),
				Ipv4:    makeIPv4Uint32(127, 0, 0, 1),
				Ipv4Str: "127.0.0.1",
				Vlan:    "fake-vlan",
				Reserve: true,
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := makeIPv4(tt.vlanName, tt.ipv4, tt.reserved)
			if diff := cmp.Diff(got, tt.want, protocmp.Transform()); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
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
