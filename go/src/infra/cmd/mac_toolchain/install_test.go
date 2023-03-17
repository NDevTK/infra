// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.chromium.org/luci/common/errors"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallXcode(t *testing.T) {
	t.Parallel()

	Convey("installXcode works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		installArgs := InstallArgs{
			xcodeVersion:           "testVersion",
			xcodeAppPath:           "testdata/Xcode-old.app",
			acceptedLicensesFile:   "testdata/acceptedLicenses.plist",
			cipdPackagePrefix:      "test/prefix",
			kind:                   macKind,
			serviceAccountJSON:     "",
			packageInstallerOnBots: "testdata/dummy_installer",
			withRuntime:            false,
		}

		Convey("for accepted license, mac", func() {
			s.ReturnOutput = []string{
				"12.2.1", // MacOS Version
				"cipd dry run",
				"cipd ensures",
				"chomod prints nothing",
				"", // No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch installs packages",
				"xcrun simctl list prints a list of all simulators installed",
				"Developer mode is currently enabled.\n",
			}
			err := installXcode(ctx, installArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 9)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "testdata/Xcode-old.app",
			})

			So(s.Calls[4].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[4].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[5].Executable, ShouldEqual, "sudo")
			So(s.Calls[5].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-old.app"})

			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[7].Executable, ShouldEqual, "xcrun")
			So(s.Calls[7].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[8].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[8].Args, ShouldResemble, []string{"-status"})
		})

		Convey("for already installed package with Developer mode enabled and -runFirstLaunch needs to run", func() {
			s.ReturnError = []error{
				errors.Reason("check OS version error").Err(),
				errors.Reason("CIPD package already installed").Err(),
			}
			s.ReturnOutput = []string{
				"12.2.1", // MacOS Version
				"cipd dry run",
				"original/Xcode.app",
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch installs packages",
				"xcrun simctl list prints a list of all simulators installed",
				"xcode-select -s prints nothing",
				"Developer mode is currently enabled.\n",
			}
			err := installXcode(ctx, installArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 8)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\n")
			So(s.Calls[1].Env, ShouldResemble, []string(nil))

			So(s.Calls[2].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[2].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[3].Executable, ShouldEqual, "sudo")
			So(s.Calls[3].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-old.app"})

			So(s.Calls[4].Executable, ShouldEqual, "sudo")
			So(s.Calls[4].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[5].Executable, ShouldEqual, "xcrun")
			So(s.Calls[5].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "original/Xcode.app"})

			So(s.Calls[7].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[7].Args, ShouldResemble, []string{"-status"})

		})

		Convey("for already installed package with Developer mode disabled", func() {
			s.ReturnError = []error{
				errors.Reason("check OS version error").Err(),
				errors.Reason("already installed").Err(),
			}
			s.ReturnOutput = []string{
				"12.2.1", // MacOS Version
				"",
				"original/Xcode.app",
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch installs packages",
				"xcrun simctl list prints a list of all simulators installed",
				"Developer mode is currently disabled.",
			}
			err := installXcode(ctx, installArgs)
			So(err.Error(), ShouldContainSubstring, "Developer mode is currently disabled! Please use `sudo /usr/sbin/DevToolsSecurity -enable` to enable.")
			So(s.Calls, ShouldHaveLength, 8)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[2].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[3].Executable, ShouldEqual, "sudo")
			So(s.Calls[3].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-old.app"})

			So(s.Calls[4].Executable, ShouldEqual, "sudo")
			So(s.Calls[4].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[5].Executable, ShouldEqual, "xcrun")
			So(s.Calls[5].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "original/Xcode.app"})

			So(s.Calls[7].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[7].Args, ShouldResemble, []string{"-status"})
		})

		Convey("with a service account", func() {
			s.ReturnOutput = []string{
				"12.2.1", // MacOS Version
				"cipd dry run",
				"cipd ensures",
				"chomod prints nothing",
				"", // No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch installs packages",
				"xcrun simctl list prints a list of all simulators installed",
				"Developer mode is currently enabled.\n",
			}
			installArgs.serviceAccountJSON = "test/service-account.json"
			err := installXcode(ctx, installArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 9)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
				"-service-account-json", "test/service-account.json",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
				"-service-account-json", "test/service-account.json",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "testdata/Xcode-old.app",
			})

			So(s.Calls[4].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[4].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[5].Executable, ShouldEqual, "sudo")
			So(s.Calls[5].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-old.app"})

			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[7].Executable, ShouldEqual, "xcrun")
			So(s.Calls[7].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[8].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[8].Args, ShouldResemble, []string{"-status"})
		})

		Convey("for new license, ios", func() {
			s.ReturnError = []error{
				errors.Reason("check OS version error").Err(),
				errors.Reason("already installed").Err(),
			}
			s.ReturnOutput = []string{
				"12.2.1", // MacOS Version
				"cipd dry run",
				"old/xcode/path",
				"xcode-select -s prints nothing",
				"license accept",
				"xcode-select -s prints nothing",
				"old/xcode/path",
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch",
				"xcrun simctl list prints a list of all simulators installed",
				"xcode-select -s prints nothing",
				"Developer mode is currently enabled.",
			}

			installArgs.xcodeAppPath = "testdata/Xcode-new.app"
			installArgs.kind = iosKind
			err := installXcode(ctx, installArgs)
			So(err, ShouldBeNil)
			So(len(s.Calls), ShouldEqual, 12)

			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-new.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual,
				"test/prefix/mac testVersion\n"+
					"test/prefix/ios testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[2].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[3].Executable, ShouldEqual, "sudo")
			So(s.Calls[3].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})

			So(s.Calls[4].Executable, ShouldEqual, "sudo")
			So(s.Calls[4].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-license", "accept"})

			So(s.Calls[5].Executable, ShouldEqual, "sudo")
			So(s.Calls[5].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "old/xcode/path"})

			So(s.Calls[6].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[6].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[7].Executable, ShouldEqual, "sudo")
			So(s.Calls[7].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})

			So(s.Calls[8].Executable, ShouldEqual, "sudo")
			So(s.Calls[8].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[9].Executable, ShouldEqual, "xcrun")
			So(s.Calls[9].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[10].Executable, ShouldEqual, "sudo")
			So(s.Calls[10].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "old/xcode/path"})

			So(s.Calls[11].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[11].Args, ShouldResemble, []string{"-status"})
		})

	})

	Convey("install Xcode ios mode with/without ios runtime", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		installArgs := InstallArgs{
			xcodeVersion:           "testVersion",
			xcodeAppPath:           "testdata/Xcode-old.app",
			acceptedLicensesFile:   "testdata/acceptedLicenses.plist",
			cipdPackagePrefix:      "test/prefix",
			kind:                   iosKind,
			serviceAccountJSON:     "",
			packageInstallerOnBots: "testdata/dummy_installer",
			withRuntime:            true,
		}

		Convey("install with runtime", func() {
			s.ReturnOutput = []string{
				"12.2.1",                         // MacOS Version
				"cipd dry run",                   // 0 (index in s.Calls, same below)
				"cipd ensures",                   // 1
				"chomod prints nothing",          // 2
				"",                               // 3: successfully resolved (by returning no error)
				"cipd dry run",                   // 4
				"cipd ensures",                   // 5
				"chomod prints nothing",          // 6
				"",                               // 7: No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing", // 8
				"xcodebuild -runFirstLaunch installs packages",                // 9
				"xcrun simctl list prints a list of all simulators installed", // 10
				"Developer mode is currently enabled.\n",                      // 11
			}
			installArgsForTest := installArgs
			installArgsForTest.withRuntime = true
			// Clean up the added runtime dir.
			defer os.RemoveAll("testdata/Xcode-old.app/Contents/Developer/Platforms")
			err := installXcode(ctx, installArgsForTest)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 13)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\ntest/prefix/ios testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\ntest/prefix/ios testVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "testdata/Xcode-old.app",
			})

			So(s.Calls[4].Executable, ShouldEqual, "cipd")
			So(s.Calls[4].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testVersion",
			})

			// Normalize for win builder tests.
			runtimeInstallPath := filepath.FromSlash("testdata/Xcode-old.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes")

			So(s.Calls[5].Executable, ShouldEqual, "cipd")
			So(s.Calls[5].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", runtimeInstallPath,
			})
			So(s.Calls[5].ConsumedStdin, ShouldEqual, "test/prefix/ios_runtime testVersion\n")

			So(s.Calls[6].Executable, ShouldEqual, "cipd")
			So(s.Calls[6].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", runtimeInstallPath,
			})
			So(s.Calls[6].ConsumedStdin, ShouldEqual, "test/prefix/ios_runtime testVersion\n")

			So(s.Calls[7].Executable, ShouldEqual, "chmod")
			So(s.Calls[7].Args, ShouldResemble, []string{
				"-R", "u+w", runtimeInstallPath,
			})

			So(s.Calls[8].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[8].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[9].Executable, ShouldEqual, "sudo")
			So(s.Calls[9].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-old.app"})

			So(s.Calls[10].Executable, ShouldEqual, "sudo")
			So(s.Calls[10].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[11].Executable, ShouldEqual, "xcrun")
			So(s.Calls[11].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[12].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[12].Args, ShouldResemble, []string{"-status"})
		})

		Convey("with runtime but runtime already exist", func() {
			s.ReturnOutput = []string{
				"12.2.1",                         // MacOS Version
				"cipd dry run",                   // 0 (index in s.Calls, same below)
				"cipd ensures",                   // 1
				"chomod prints nothing",          // 2
				"",                               // 3: No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing", // 4
				"xcodebuild -runFirstLaunch installs packages",                // 5
				"xcrun simctl list prints a list of all simulators installed", // 6
				"Developer mode is currently enabled.\n",                      // 7
			}
			installArgsForTest := installArgs
			installArgsForTest.withRuntime = true
			installArgsForTest.xcodeAppPath = "testdata/Xcode-with-runtime.app"
			err := installXcode(ctx, installArgsForTest)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 9)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-with-runtime.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\ntest/prefix/ios testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "testdata/Xcode-with-runtime.app",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\ntest/prefix/ios testVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "testdata/Xcode-with-runtime.app",
			})

			So(s.Calls[4].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[4].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[5].Executable, ShouldEqual, "sudo")
			So(s.Calls[5].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-with-runtime.app"})

			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[7].Executable, ShouldEqual, "xcrun")
			So(s.Calls[7].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[8].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[8].Args, ShouldResemble, []string{"-status"})
		})

		Convey("without runtime", func() {
			s.ReturnOutput = []string{
				"12.2.1",                         // MacOS Version
				"cipd dry run",                   // 0 (index in s.Calls, same below)
				"cipd ensures",                   // 1
				"chomod prints nothing",          // 2
				"",                               // 3: No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing", // 4
				"xcodebuild -runFirstLaunch installs packages",                // 5
				"xcrun simctl list prints a list of all simulators installed", // 6
				"Developer mode is currently enabled.\n",                      // 7
			}
			installArgsForTest := installArgs
			installArgsForTest.withRuntime = false
			installArgsForTest.xcodeAppPath = "testdata/Xcode-old.app"
			err := installXcode(ctx, installArgsForTest)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 9)
			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\ntest/prefix/ios testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "testdata/Xcode-old.app",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/mac testVersion\ntest/prefix/ios testVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "testdata/Xcode-old.app",
			})

			So(s.Calls[4].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[4].Args, ShouldResemble, []string{"-p"})

			So(s.Calls[5].Executable, ShouldEqual, "sudo")
			So(s.Calls[5].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-old.app"})

			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})

			So(s.Calls[7].Executable, ShouldEqual, "xcrun")
			So(s.Calls[7].Args, ShouldResemble, []string{"simctl", "list"})

			So(s.Calls[8].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[8].Args, ShouldResemble, []string{"-status"})
		})
	})

	Convey("installXcode on MacOS 13+", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		installArgs := InstallArgs{
			xcodeVersion:           "testVersion",
			xcodeAppPath:           "testdata/Xcode-new.app",
			acceptedLicensesFile:   "testdata/acceptedLicenses.plist",
			cipdPackagePrefix:      "test/prefix",
			kind:                   macKind,
			serviceAccountJSON:     "",
			packageInstallerOnBots: "testdata/dummy_installer",
			withRuntime:            false,
		}

		Convey("current Xcode CFBundle version matches what's on cipd", func() {
			s.ReturnOutput = []string{
				"13.2.1", // MacOS Version
				"cf_bundle_version:12345",
				"", // No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing",
				"license accpet",
				"testdata/Xcode-new.app",
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch",
				"xcrun simctl list prints a list of all simulators installed",
				"xcode-select -s prints nothing",
				"Developer mode is currently enabled.\n",
			}
			err := installXcode(ctx, installArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 11)

			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testVersion",
			})
			So(s.Calls[2].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[2].Args, ShouldResemble, []string{"-p"})
			So(s.Calls[3].Executable, ShouldEqual, "sudo")
			So(s.Calls[3].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})
			So(s.Calls[4].Executable, ShouldEqual, "sudo")
			So(s.Calls[4].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-license", "accept"})
			So(s.Calls[5].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[5].Args, ShouldResemble, []string{"-p"})
			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})
			So(s.Calls[7].Executable, ShouldEqual, "sudo")
			So(s.Calls[7].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})
			So(s.Calls[8].Executable, ShouldEqual, "xcrun")
			So(s.Calls[8].Args, ShouldResemble, []string{"simctl", "list"})
			So(s.Calls[9].Executable, ShouldEqual, "sudo")
			So(s.Calls[9].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})
			So(s.Calls[10].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[10].Args, ShouldResemble, []string{"-status"})
		})

		Convey("current Xcode CFBundle version mismatch what's on cipd", func() {
			s.ReturnOutput = []string{
				"13.2.1", // MacOS Version
				// cipd describe returns nothing to ensure backward compatibility when CFBundleVersions tags
				// don't exist on older packages
				"",
				"cipd dry run",
				"cipd ensures",
				"chomod prints nothing",
				"", // No original Xcode when running xcode-select -p
				"xcode-select -s prints nothing",
				"license accpet",
				"testdata/Xcode-new.app",
				"xcode-select -s prints nothing",
				"xcodebuild -runFirstLaunch",
				"xcrun simctl list prints a list of all simulators installed",
				"xcode-select -s prints nothing",
				"Developer mode is currently enabled.\n",
			}
			err := installXcode(ctx, installArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 14)

			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testVersion",
			})
			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "testdata/Xcode-new.app",
			})
			So(s.Calls[3].Executable, ShouldEqual, "cipd")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "testdata/Xcode-new.app",
			})
			So(s.Calls[4].Executable, ShouldEqual, "chmod")
			So(s.Calls[4].Args, ShouldResemble, []string{
				"-R", "u+w", "testdata/Xcode-new.app",
			})
			So(s.Calls[5].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[5].Args, ShouldResemble, []string{"-p"})
			So(s.Calls[6].Executable, ShouldEqual, "sudo")
			So(s.Calls[6].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})
			So(s.Calls[7].Executable, ShouldEqual, "sudo")
			So(s.Calls[7].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-license", "accept"})
			So(s.Calls[8].Executable, ShouldEqual, "/usr/bin/xcode-select")
			So(s.Calls[8].Args, ShouldResemble, []string{"-p"})
			So(s.Calls[9].Executable, ShouldEqual, "sudo")
			So(s.Calls[9].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})
			So(s.Calls[10].Executable, ShouldEqual, "sudo")
			So(s.Calls[10].Args, ShouldResemble, []string{"/usr/bin/xcodebuild", "-runFirstLaunch"})
			So(s.Calls[11].Executable, ShouldEqual, "xcrun")
			So(s.Calls[11].Args, ShouldResemble, []string{"simctl", "list"})
			So(s.Calls[12].Executable, ShouldEqual, "sudo")
			So(s.Calls[12].Args, ShouldResemble, []string{"/usr/bin/xcode-select", "-s", "testdata/Xcode-new.app"})
			So(s.Calls[13].Executable, ShouldEqual, "/usr/sbin/DevToolsSecurity")
			So(s.Calls[13].Args, ShouldResemble, []string{"-status"})
		})
	})

	Convey("describeRef works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		Convey("ref exists", func() {
			s.ReturnOutput = []string{
				"Package:       test/prefix/mac",
			}
			output, err := describeRef(ctx, "test/prefix/mac", "testXcodeVersion")
			So(err, ShouldBeNil)
			So(output, ShouldNotEqual, "")
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testXcodeVersion",
			})
		})
		Convey("ref doesn't exist", func() {
			s.ReturnError = []error{errors.Reason("no such ref").Err()}
			output, err := describeRef(ctx, "test/prefix/mac", "testNonExistRef")
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testNonExistRef",
			})
			So(output, ShouldEqual, "")
			So(err.Error(), ShouldContainSubstring, "Error when describing package path test/prefix/mac with ref testNonExistRef.")
		})
	})

	Convey("shouldReInstallXcode works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		Convey("Xcode doesn't exists so it needs to be re-intalled", func() {
			result, err := shouldReInstallXcode(ctx, "test/prefix/mac", "testdata/nonexistent.app", "testXcodeVersion")
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, true)
		})

		Convey("Xcode exists but CFBundleVersion tag on cipd not found so it needs to be re-intalled", func() {
			s.ReturnOutput = []string{
				"Package:       test/prefix/mac",
			}
			result, err := shouldReInstallXcode(ctx, "test/prefix", "testdata/Xcode-new.app", "testXcodeVersion")
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testXcodeVersion",
			})
			So(result, ShouldEqual, true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Unable to parse CFBundleVersion from cipd")
		})

		Convey("Xcode exists but CFBundleVersion is different on cipd so it needs to be re-intalled", func() {
			s.ReturnOutput = []string{
				"cf_bundle_version:12346",
			}
			result, err := shouldReInstallXcode(ctx, "test/prefix", "testdata/Xcode-new.app", "testXcodeVersion")
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testXcodeVersion",
			})
			So(result, ShouldEqual, true)
			So(err, ShouldBeNil)
		})

		Convey("Xcode exists and CFBundleVersion is the same on cipd so it doesn't need to be re-intalled", func() {
			s.ReturnOutput = []string{
				"cf_bundle_version:12345",
			}
			result, err := shouldReInstallXcode(ctx, "test/prefix", "testdata/Xcode-new.app", "testXcodeVersion")
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"describe", "test/prefix/mac", "-version", "testXcodeVersion",
			})
			So(result, ShouldEqual, false)
			So(err, ShouldBeNil)
		})

	})

	Convey("resolveRef works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		Convey("ref exists", func() {
			err := resolveRef(ctx, "test/prefix/ios_runtime", "testXcodeVersion", "")
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testXcodeVersion",
			})
		})
		Convey("ref doesn't exist", func() {
			s.ReturnError = []error{errors.Reason("input ref doesn't exist").Err()}
			err := resolveRef(ctx, "test/prefix/ios_runtime", "testNonExistRef", "")
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testNonExistRef",
			})
			So(err.Error(), ShouldContainSubstring, "Error when resolving package path test/prefix/ios_runtime with ref testNonExistRef.")
		})
	})

	Convey("resolveRuntimeRef works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)
		Convey("only input xcode version", func() {
			resolveRuntimeRefArgs := ResolveRuntimeRefArgs{
				runtimeVersion:     "",
				xcodeVersion:       "testXcodeVersion",
				packagePath:        "test/prefix/ios_runtime",
				serviceAccountJSON: "",
			}
			ver, err := resolveRuntimeRef(ctx, resolveRuntimeRefArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testXcodeVersion",
			})
			So(ver, ShouldEqual, "testXcodeVersion")
		})
		Convey("only input sim runtime version", func() {
			resolveRuntimeRefArgs := ResolveRuntimeRefArgs{
				runtimeVersion:     "testSimVersion",
				xcodeVersion:       "",
				packagePath:        "test/prefix/ios_runtime",
				serviceAccountJSON: "",
			}
			ver, err := resolveRuntimeRef(ctx, resolveRuntimeRefArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion",
			})
			So(ver, ShouldEqual, "testSimVersion")
		})
		Convey("input both Xcode and sim version: default runtime exists", func() {
			resolveRuntimeRefArgs := ResolveRuntimeRefArgs{
				runtimeVersion:     "testSimVersion",
				xcodeVersion:       "testXcodeVersion",
				packagePath:        "test/prefix/ios_runtime",
				serviceAccountJSON: "",
			}
			ver, err := resolveRuntimeRef(ctx, resolveRuntimeRefArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 1)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion_testXcodeVersion",
			})
			So(ver, ShouldEqual, "testSimVersion_testXcodeVersion")
		})
		Convey("input both Xcode and sim version: fallback to uploaded runtime", func() {
			s.ReturnError = []error{errors.Reason("default runtime doesn't exist").Err()}
			resolveRuntimeRefArgs := ResolveRuntimeRefArgs{
				runtimeVersion:     "testSimVersion",
				xcodeVersion:       "testXcodeVersion",
				packagePath:        "test/prefix/ios_runtime",
				serviceAccountJSON: "",
			}
			ver, err := resolveRuntimeRef(ctx, resolveRuntimeRefArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 2)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion_testXcodeVersion",
			})
			So(s.Calls[1].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion",
			})
			So(ver, ShouldEqual, "testSimVersion")
		})
		Convey("input both Xcode and sim version: fallback to any latest runtime", func() {
			s.ReturnError = []error{
				errors.Reason("default runtime doesn't exist").Err(),
				errors.Reason("uploaded runtime doesn't exist").Err(),
			}
			resolveRuntimeRefArgs := ResolveRuntimeRefArgs{
				runtimeVersion:     "testSimVersion",
				xcodeVersion:       "testXcodeVersion",
				packagePath:        "test/prefix/ios_runtime",
				serviceAccountJSON: "",
			}
			ver, err := resolveRuntimeRef(ctx, resolveRuntimeRefArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 3)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion_testXcodeVersion",
			})
			So(s.Calls[1].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion",
			})
			So(s.Calls[2].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion_latest",
			})
			So(ver, ShouldEqual, "testSimVersion_latest")
		})
		Convey("input both Xcode and sim version: raise when all fallbacks fail", func() {
			s.ReturnError = []error{
				errors.Reason("default runtime doesn't exist").Err(),
				errors.Reason("uploaded runtime doesn't exist").Err(),
				errors.Reason("any latest runtime doesn't exist").Err(),
			}
			resolveRuntimeRefArgs := ResolveRuntimeRefArgs{
				runtimeVersion:     "testSimVersion",
				xcodeVersion:       "testXcodeVersion",
				packagePath:        "test/prefix/ios_runtime",
				serviceAccountJSON: "",
			}
			ver, err := resolveRuntimeRef(ctx, resolveRuntimeRefArgs)
			So(s.Calls, ShouldHaveLength, 3)
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion_testXcodeVersion",
			})
			So(s.Calls[1].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion",
			})
			So(s.Calls[2].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion_latest",
			})
			So(err.Error(), ShouldContainSubstring, "Failed to resolve runtime ref given runtime version: testSimVersion, xcode version: testXcodeVersion.")
			So(ver, ShouldEqual, "")
		})

	})

	Convey("installRuntime works", t, func() {
		var s MockSession
		ctx := useMockCmd(context.Background(), &s)

		Convey("install an Xcode default runtime", func() {
			runtimeInstallArgs := RuntimeInstallArgs{
				runtimeVersion:     "",
				xcodeVersion:       "testVersion",
				installPath:        "test/path/to/install/runtimes",
				cipdPackagePrefix:  "test/prefix",
				serviceAccountJSON: "",
			}
			err := installRuntime(ctx, runtimeInstallArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 4)
			So(s.Calls[0].Executable, ShouldEqual, "cipd")
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testVersion",
			})

			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "test/path/to/install/runtimes",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/ios_runtime testVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "test/path/to/install/runtimes",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/ios_runtime testVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "test/path/to/install/runtimes",
			})
		})

		Convey("install an uploaded runtime", func() {
			runtimeInstallArgs := RuntimeInstallArgs{
				runtimeVersion:     "testSimVersion",
				xcodeVersion:       "",
				installPath:        "test/path/to/install/runtimes",
				cipdPackagePrefix:  "test/prefix",
				serviceAccountJSON: "",
			}
			err := installRuntime(ctx, runtimeInstallArgs)
			So(err, ShouldBeNil)
			So(s.Calls, ShouldHaveLength, 4)
			So(s.Calls[0].Executable, ShouldEqual, "cipd")
			So(s.Calls[0].Args, ShouldResemble, []string{
				"resolve", "test/prefix/ios_runtime", "-version", "testSimVersion",
			})

			So(s.Calls[1].Executable, ShouldEqual, "cipd")
			So(s.Calls[1].Args, ShouldResemble, []string{
				"puppet-check-updates", "-ensure-file", "-", "-root", "test/path/to/install/runtimes",
			})
			So(s.Calls[1].ConsumedStdin, ShouldEqual, "test/prefix/ios_runtime testSimVersion\n")

			So(s.Calls[2].Executable, ShouldEqual, "cipd")
			So(s.Calls[2].Args, ShouldResemble, []string{
				"ensure", "-ensure-file", "-", "-root", "test/path/to/install/runtimes",
			})
			So(s.Calls[2].ConsumedStdin, ShouldEqual, "test/prefix/ios_runtime testSimVersion\n")

			So(s.Calls[3].Executable, ShouldEqual, "chmod")
			So(s.Calls[3].Args, ShouldResemble, []string{
				"-R", "u+w", "test/path/to/install/runtimes",
			})
		})
	})

	Convey("removeCipdFiles works", t, func() {
		Convey("remove cipd files whether it exists or not", func() {
			srcPath := "testdata/"
			tmpCipdPath, err := os.MkdirTemp(srcPath, "tmp")
			So(err, ShouldBeNil)
			defer os.RemoveAll(tmpCipdPath)

			// folder is empty but it should still succeed
			err = removeCipdFiles(tmpCipdPath)
			So(err, ShouldBeNil)

			// create cipd files so they can be removed
			dotCipdPath := filepath.Join(tmpCipdPath, ".cipd")
			dotXcodeVersionPath := filepath.Join(tmpCipdPath, ".xcode_versions")
			err = os.MkdirAll(dotCipdPath, 0700)
			So(err, ShouldBeNil)
			err = os.MkdirAll(dotXcodeVersionPath, 0700)
			So(err, ShouldBeNil)

			err = removeCipdFiles(tmpCipdPath)
			So(err, ShouldBeNil)

			// files should not exist after removing
			_, err = os.Stat(dotCipdPath)
			So(err, ShouldNotBeNil)
			_, err = os.Stat(dotXcodeVersionPath)
			So(err, ShouldNotBeNil)

		})
	})

}
