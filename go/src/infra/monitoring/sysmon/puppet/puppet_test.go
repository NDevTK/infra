// Copyright (c) 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package puppet

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/common/clock/testclock"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/common/tsmon"
)

const validCert = `
-----BEGIN CERTIFICATE-----
MIICEDCCAXkCFHrEhtWbYRk3IpQ4n7iho79PNOZLMA0GCSqGSIb3DQEBCwUAMEcx
CzAJBgNVBAYTAlVTMQswCQYDVQQIDAJDQTEPMA0GA1UECgwGR29vZ2xlMRowGAYD
VQQDDBFjZXJ0LWZhY3RvcnktdGVzdDAeFw0yMjAyMjMyMDM5MzdaFw0yMzAyMjMy
MDM5MzdaMEcxCzAJBgNVBAYTAlVTMQswCQYDVQQIDAJDQTEPMA0GA1UECgwGR29v
Z2xlMRowGAYDVQQDDBFjZXJ0LWZhY3RvcnktdGVzdDCBnzANBgkqhkiG9w0BAQEF
AAOBjQAwgYkCgYEAvxeA5upd6BIGUTouRmf+/9yc3GIvzAOCp1aSNQTimQT8aaZs
9YWEg1YeLUsb5p5sSQk+lnVbShOk7BeSTlCtL1Nps3uCfh5J6yCUhykRGJvYEfDY
nF38mX4q6MuC+DrBcVECCTt4HnmA7saAV5zDY7zYqTXZf8CjCYT9h9qQmQMCAwEA
ATANBgkqhkiG9w0BAQsFAAOBgQCjW0TGqW4tjBZNtPMk31EGwYrrPFQ4KbUJOYvT
pE4Z7oIshI06JZB+DJsuyFUVcTWVafCUmiuqBHjVKlajY0Aa8vFJ/hXKDR4FSgTv
NHxMXl29/1NC0bGuiQl3C/8foLAMBUT6Bg0tgBVFN9t8ADNVWb/Mr+o7SS4f8Fr0
QmR8MA==
-----END CERTIFICATE-----`

func TestMetrics(t *testing.T) {
	c := context.Background()
	c, _ = tsmon.WithDummyInMemory(c)
	c, _ = testclock.UseTime(c, time.Unix(1440132466, 0).Add(123450*time.Millisecond))

	Convey("Puppet last_run_summary.yaml metrics", t, func() {
		file, err := os.CreateTemp("", "sysmon-puppet-test")
		So(err, ShouldBeNil)

		defer file.Close()
		defer os.Remove(file.Name())

		Convey("with an empty file", func() {
			So(updateLastRunStats(c, file.Name()), ShouldBeNil)
			So(configVersion.Get(c), ShouldEqual, 0)
			So(puppetVersion.Get(c), ShouldEqual, "")
		})

		Convey("with a missing file", func() {
			So(updateLastRunStats(c, "file does not exist"), ShouldNotBeNil)
		})

		Convey("with an invalid file", func() {
			file.Write([]byte("\""))
			file.Sync()
			So(updateLastRunStats(c, file.Name()), ShouldNotBeNil)
		})

		Convey("with a file containing an array", func() {
			file.Write([]byte("- one\n- two\n"))
			file.Sync()
			So(updateLastRunStats(c, file.Name()), ShouldNotBeNil)
		})

		Convey("metrics", func() {
			file.Write([]byte(`---
  version:
    config: 1440131220
    puppet: "3.6.2"
  resources:
    changed: 1
    failed: 2
    failed_to_restart: 3
    out_of_sync: 4
    restarted: 5
    scheduled: 6
    skipped: 7
    total: 51
  time:
    anchor: 0.01
    apt_key: 0.02
    config_retrieval: 0.03
    exec: 0.04
    file: 0.05
    filebucket: 0.06
    package: 0.07
    schedule: 0.08
    service: 0.09
    total: 0.10
    last_run: 1440132466
  changes:
    total: 4
  events:
    failure: 1
    success: 2
    total: 3`))
			file.Sync()
			So(updateLastRunStats(c, file.Name()), ShouldBeNil)

			So(configVersion.Get(c), ShouldEqual, 1440131220)
			So(puppetVersion.Get(c), ShouldEqual, "3.6.2")

			So(events.Get(c, "failure"), ShouldEqual, 1)
			So(events.Get(c, "success"), ShouldEqual, 2)
			So(events.Get(c, "total"), ShouldEqual, 0)

			So(failure.Get(c), ShouldBeTrue)

			So(resources.Get(c, "changed"), ShouldEqual, 1)
			So(resources.Get(c, "failed"), ShouldEqual, 2)
			So(resources.Get(c, "failed_to_restart"), ShouldEqual, 3)
			So(resources.Get(c, "out_of_sync"), ShouldEqual, 4)
			So(resources.Get(c, "restarted"), ShouldEqual, 5)
			So(resources.Get(c, "scheduled"), ShouldEqual, 6)
			So(resources.Get(c, "skipped"), ShouldEqual, 7)
			So(resources.Get(c, "total"), ShouldEqual, 51)

			So(times.Get(c, "anchor"), ShouldEqual, 0.01)
			So(times.Get(c, "apt_key"), ShouldEqual, 0.02)
			So(times.Get(c, "config_retrieval"), ShouldEqual, 0.03)
			So(times.Get(c, "exec"), ShouldEqual, 0.04)
			So(times.Get(c, "file"), ShouldEqual, 0.05)
			So(times.Get(c, "filebucket"), ShouldEqual, 0.06)
			So(times.Get(c, "package"), ShouldEqual, 0.07)
			So(times.Get(c, "schedule"), ShouldEqual, 0.08)
			So(times.Get(c, "service"), ShouldEqual, 0.09)
			So(times.Get(c, "total"), ShouldEqual, 0)

			So(age.Get(c), ShouldEqual, 123.45)
		})

		Convey("metrics successful run", func() {
			file.Write([]byte(`---
  events:
    failure: 0
    success: 2
    total: 2`))
			file.Sync()
			So(updateLastRunStats(c, file.Name()), ShouldBeNil)
			So(failure.Get(c), ShouldBeFalse)
		})

		Convey("metrics with completely failed run", func() {
			file.Write([]byte(`---`))
			file.Sync()
			So(updateLastRunStats(c, file.Name()), ShouldBeNil)
			So(failure.Get(c), ShouldBeTrue)
		})

	})

	Convey("Puppet is_canary metric", t, func() {
		Convey("with a missing file", func() {
			So(updateIsCanary(c, "file does not exist"), ShouldNotBeNil)
			So(isCanary.Get(c), ShouldBeFalse)
		})

		Convey("with a present file", func() {
			file, err := os.CreateTemp("", "sysmon-puppet-test")
			So(err, ShouldBeNil)

			Convey("with environment=canary", func() {
				_, err := file.Write([]byte("foo=bar\nenvironment=canary\nblah=blah\n"))
				So(err, ShouldBeNil)
				So(file.Sync(), ShouldBeNil)
				So(updateIsCanary(c, file.Name()), ShouldBeNil)
				So(isCanary.Get(c), ShouldBeTrue)
			})
		})
	})

	Convey("Puppet exit_status metric", t, func() {
		Convey("with a missing file", func() {
			err := updateExitStatus(c, []string{"file does not exist"})
			So(err, ShouldNotBeNil)
		})

		Convey("with a present file", func() {
			file, err := os.CreateTemp("", "sysmon-puppet-test")
			So(err, ShouldBeNil)

			Convey("containing a valid number", func() {
				file.Write([]byte("42"))
				file.Sync()
				So(updateExitStatus(c, []string{file.Name()}), ShouldBeNil)
				So(exitStatus.Get(c), ShouldEqual, 42)
			})

			Convey(`containing a valid number and a \n`, func() {
				file.Write([]byte("42\n"))
				file.Sync()
				So(updateExitStatus(c, []string{file.Name()}), ShouldBeNil)
				So(exitStatus.Get(c), ShouldEqual, 42)
			})

			Convey("containing an invalid number", func() {
				file.Write([]byte("not a number"))
				file.Sync()
				So(updateExitStatus(c, []string{file.Name()}), ShouldNotBeNil)
			})

			Convey("second in the list", func() {
				file.Write([]byte("42"))
				file.Sync()
				So(updateExitStatus(c, []string{"does not exist", file.Name()}), ShouldBeNil)
				So(exitStatus.Get(c), ShouldEqual, 42)
			})
		})
	})

	Convey("Puppet cert_expiry metric", t, func() {
		Convey("with no file found", func() {
			So(updateCertExpiry(c, "not_a_path"), ShouldErrLike, "cert not found")
			So(certExpiry.Get(c), ShouldEqual, 0)
		})

		Convey("with valid and invalid cert contents", func() {
			dir, err := os.MkdirTemp("", "test_cert_expiry")
			So(err, ShouldBeNil)
			defer os.RemoveAll(dir)
			fqdnHost, _ := os.Hostname()

			Convey("with invalid certificate, parsing err", func() {
				testCertInvalid := filepath.Join(dir, fqdnHost+".pem")
				f, err := os.Create(testCertInvalid)
				So(err, ShouldBeNil)

				_, err = f.WriteString("invalid cert contents")
				So(err, ShouldBeNil)
				err = f.Sync()
				So(err, ShouldBeNil)

				So(updateCertExpiry(c, dir), ShouldErrLike, "error parsing certificate")
				So(certExpiry.Get(c), ShouldEqual, 0)
			})

			Convey("with valid certificate", func() {
				testCert := filepath.Join(dir, fqdnHost+".pem")
				f, err := os.Create(testCert)
				So(err, ShouldBeNil)

				_, err = f.WriteString(validCert)
				So(err, ShouldBeNil)
				err = f.Sync()
				So(err, ShouldBeNil)

				So(updateCertExpiry(c, dir), ShouldBeNil)
				// Clock time 2015-08-21, certificate notAfter 2023-02-23
				So(certExpiry.Get(c), ShouldEqual, 237052188)
			})
		})
	})
}
