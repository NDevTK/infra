// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ninjalog

import (
	"bytes"
	_ "embed"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	logTestCase = `# ninja log v5
76	187	0	resources/inspector/devtools_extension_api.js	75430546595be7c2
80	284	0	gen/autofill_regex_constants.cc	fa33c8d7ce1d8791
78	286	0	gen/angle/commit_id.py	4ede38e2c1617d8c
79	287	0	gen/angle/copy_compiler_dll.bat	9fb635ad5d2c1109
141	287	0	PepperFlash/manifest.json	324f0a0b77c37ef
142	288	0	PepperFlash/libpepflashplayer.so	1e2c2b7845a4d4fe
287	290	0	obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp	b211d373de72f455
`

	stepsTestCase = []Step{
		{
			Start:   76 * time.Millisecond,
			End:     187 * time.Millisecond,
			Out:     "resources/inspector/devtools_extension_api.js",
			CmdHash: "75430546595be7c2",
		},
		{
			Start:   80 * time.Millisecond,
			End:     284 * time.Millisecond,
			Out:     "gen/autofill_regex_constants.cc",
			CmdHash: "fa33c8d7ce1d8791",
		},
		{
			Start:   78 * time.Millisecond,
			End:     286 * time.Millisecond,
			Out:     "gen/angle/commit_id.py",
			CmdHash: "4ede38e2c1617d8c",
		},
		{
			Start:   79 * time.Millisecond,
			End:     287 * time.Millisecond,
			Out:     "gen/angle/copy_compiler_dll.bat",
			CmdHash: "9fb635ad5d2c1109",
		},
		{
			Start:   141 * time.Millisecond,
			End:     287 * time.Millisecond,
			Out:     "PepperFlash/manifest.json",
			CmdHash: "324f0a0b77c37ef",
		},
		{
			Start:   142 * time.Millisecond,
			End:     288 * time.Millisecond,
			Out:     "PepperFlash/libpepflashplayer.so",
			CmdHash: "1e2c2b7845a4d4fe",
		},
		{
			Start:   287 * time.Millisecond,
			End:     290 * time.Millisecond,
			Out:     "obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp",
			CmdHash: "b211d373de72f455",
		},
	}

	stepsSorted = []Step{
		{
			Start:   76 * time.Millisecond,
			End:     187 * time.Millisecond,
			Out:     "resources/inspector/devtools_extension_api.js",
			CmdHash: "75430546595be7c2",
		},
		{
			Start:   78 * time.Millisecond,
			End:     286 * time.Millisecond,
			Out:     "gen/angle/commit_id.py",
			CmdHash: "4ede38e2c1617d8c",
		},
		{
			Start:   79 * time.Millisecond,
			End:     287 * time.Millisecond,
			Out:     "gen/angle/copy_compiler_dll.bat",
			CmdHash: "9fb635ad5d2c1109",
		},
		{
			Start:   80 * time.Millisecond,
			End:     284 * time.Millisecond,
			Out:     "gen/autofill_regex_constants.cc",
			CmdHash: "fa33c8d7ce1d8791",
		},
		{
			Start:   141 * time.Millisecond,
			End:     287 * time.Millisecond,
			Out:     "PepperFlash/manifest.json",
			CmdHash: "324f0a0b77c37ef",
		},
		{
			Start:   142 * time.Millisecond,
			End:     288 * time.Millisecond,
			Out:     "PepperFlash/libpepflashplayer.so",
			CmdHash: "1e2c2b7845a4d4fe",
		},
		{
			Start:   287 * time.Millisecond,
			End:     290 * time.Millisecond,
			Out:     "obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp",
			CmdHash: "b211d373de72f455",
		},
	}

	metadataTestCase = Metadata{
		BuildID:  12345,
		Platform: "Linux",
		Argv:     []string{"../../../scripts/compile.py", "--target", "Release", "--clobber", "--compiler=goma", "--", "all"},
		Cwd:      "/b/build/Linux_x64/build/src",
		Compiler: "goma",
		Cmdline:  []string{"python3", "/path/to/ninja.py", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "-j50"},
		Exit:     0,
		StepName: "compile",
		Env: map[string]string{
			"LANG":    "en_US.UTF-8",
			"SHELL":   "/bin/bash",
			"HOME":    "/home/chrome-bot",
			"PWD":     "/b/build/Linux_x64/build",
			"LOGNAME": "chrome-bot",
			"USER":    "chrome-bot",
			"PATH":    "/home/chrome-bot/bin:/b/depot_tools:/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin",
		},
		CompilerProxyInfo: "/tmp/compiler_proxy.build48-m1.chrome-bot.log.INFO.20140907-203827.14676",
		Jobs:              50,
	}
)

func TestStepsSort(t *testing.T) {
	steps := append([]Step{}, stepsTestCase...)
	sort.Sort(Steps(steps))
	if !reflect.DeepEqual(steps, stepsSorted) {
		t.Errorf("sort Steps=%v; want=%v", steps, stepsSorted)
	}
}

func TestStepsReverse(t *testing.T) {
	steps := []Step{
		{Out: "0"},
		{Out: "1"},
		{Out: "2"},
		{Out: "3"},
	}
	Steps(steps).Reverse()
	want := []Step{
		{Out: "3"},
		{Out: "2"},
		{Out: "1"},
		{Out: "0"},
	}
	if !reflect.DeepEqual(steps, want) {
		t.Errorf("steps.Reverse=%v; want=%v", steps, want)
	}
}

func TestParseBadVersion(t *testing.T) {
	_, err := Parse(".ninja_log", strings.NewReader(`# ninja log v4
0	1	0	foo	touch foo
`))
	if err == nil {
		t.Error("Parse()=_, <nil>; want=_, error")
	}
}

func TestParseSimple(t *testing.T) {
	njl, err := Parse(".ninja_log", strings.NewReader(logTestCase))
	if err != nil {
		t.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
	}

	want := &NinjaLog{
		Filename: ".ninja_log",
		Start:    1,
		Steps:    stepsTestCase,
	}
	if !reflect.DeepEqual(njl, want) {
		t.Errorf("Parse()=%v; want=%v", njl, want)
	}
}

func TestParseEmptyLine(t *testing.T) {
	njl, err := Parse(".ninja_log", strings.NewReader(logTestCase+"\n"))
	if err != nil {
		t.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
	}
	want := &NinjaLog{
		Filename: ".ninja_log",
		Start:    1,
		Steps:    stepsTestCase,
	}
	if !reflect.DeepEqual(njl, want) {
		t.Errorf("Parse()=%v; want=%v", njl, want)
	}
}

func TestParseLast(t *testing.T) {
	njl, err := Parse(".ninja_log", strings.NewReader(`# ninja log v5
1020807	1020916	0	chrome.1	e101fd46be020cfc
84	9489	0	gen/libraries.cc	9001f3182fa8210e
1024369	1041522	0	chrome	aee9d497d56c9637
76	187	0	resources/inspector/devtools_extension_api.js	75430546595be7c2
80	284	0	gen/autofill_regex_constants.cc	fa33c8d7ce1d8791
78	286	0	gen/angle/commit_id.py	4ede38e2c1617d8c
79	287	0	gen/angle/copy_compiler_dll.bat	9fb635ad5d2c1109
141	287	0	PepperFlash/manifest.json	324f0a0b77c37ef
142	288	0	PepperFlash/libpepflashplayer.so	1e2c2b7845a4d4fe
287	290	0	obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp	b211d373de72f455
`))
	if err != nil {
		t.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
	}

	want := &NinjaLog{
		Filename: ".ninja_log",
		Start:    4,
		Steps:    stepsTestCase,
	}
	if !reflect.DeepEqual(njl, want) {
		t.Errorf("Parse()=%v; want=%v", njl, want)
	}
}

func TestParseWithMetadata(t *testing.T) {
	njl, err := Parse(".ninja_log", strings.NewReader(`# ninja log v5
1020807	1020916	0	chrome.1	e101fd46be020cfc
84	9489	0	gen/libraries.cc	9001f3182fa8210e
1024369	1041522	0	chrome	aee9d497d56c9637
76	187	0	resources/inspector/devtools_extension_api.js	75430546595be7c2
80	284	0	gen/autofill_regex_constants.cc	fa33c8d7ce1d8791
78	286	0	gen/angle/commit_id.py	4ede38e2c1617d8c
79	287	0	gen/angle/copy_compiler_dll.bat	9fb635ad5d2c1109
141	287	0	PepperFlash/manifest.json	324f0a0b77c37ef
142	288	0	PepperFlash/libpepflashplayer.so	1e2c2b7845a4d4fe
287	290	0	obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp	b211d373de72f455

# end of ninja log
{"build_id": 12345, "platform": "Linux", "argv": ["../../../scripts/compile.py", "--target", "Release", "--clobber", "--compiler=goma", "--", "all"], "cmdline": ["python3", "/path/to/ninja.py", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "-j50"], "exit": 0, "step_name": "compile", "env": {"LANG": "en_US.UTF-8", "SHELL": "/bin/bash", "HOME": "/home/chrome-bot", "PWD": "/b/build/Linux_x64/build", "LOGNAME": "chrome-bot", "USER": "chrome-bot", "PATH": "/home/chrome-bot/bin:/b/depot_tools:/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin" }, "compiler_proxy_info": "/tmp/compiler_proxy.build48-m1.chrome-bot.log.INFO.20140907-203827.14676", "cwd": "/b/build/Linux_x64/build/src", "compiler": "goma"}
`))
	if err != nil {
		t.Errorf(`Parse()=_, %#v; want=_, <nil>`, err)
	}

	want := &NinjaLog{
		Filename: ".ninja_log",
		Start:    4,
		Steps:    stepsTestCase,
		Metadata: metadataTestCase,
	}
	njl.Metadata.Raw = ""

	if diff := cmp.Diff(want, njl); diff != "" {
		t.Errorf("Parse() mismatch (-want, +got):\n%s", diff)
	}
}

func TestParseBadMetadata(t *testing.T) {
	// https://bugs.chromium.org/p/chromium/issues/detail?id=667571
	njl, err := Parse(".ninja_log", strings.NewReader(`# ninja log v5
1020807	1020916	0	chrome.1	e101fd46be020cfc
84	9489	0	gen/libraries.cc	9001f3182fa8210e
1024369	1041522	0	chrome	aee9d497d56c9637
76	187	0	resources/inspector/devtools_extension_api.js	75430546595be7c2
80	284	0	gen/autofill_regex_constants.cc	fa33c8d7ce1d8791
78	286	0	gen/angle/commit_id.py	4ede38e2c1617d8c
79	287	0	gen/angle/copy_compiler_dll.bat	9fb635ad5d2c1109
141	287	0	PepperFlash/manifest.json	324f0a0b77c37ef
142	288	0	PepperFlash/libpepflashplayer.so	1e2c2b7845a4d4fe
287	290	0	obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp	b211d373de72f455
# end of ninja log
{"platform": "Linux", "argv": ["/b/build/scripts/upload_goma_logs.py", "--upload-compiler-proxy-info", "--json-status", "/b/build/cache/cipd/goma/jsonstatus", "--ninja-log-outdir", "/b/build/pdfium/build/pdfium/out/debug_xfa_v8", "--ninja-log-compiler", "unknown", "--ninja-log-command", "['ninja', '-C', Path('checkout', 'out','debug_xfa_v8'), '-j', 80]", "--ninja-log-exit-status", "0", "--goma-stats-file", "/b/build/pdfium/.recipe_runtime/tmpOgwx97/build_data/goma_stats_proto", "--goma-crash-report-id-file", "/b/build/pdfium/.recipe_runtime/tmpOgwx97/build_data/crash_report_id_file", "--build-data-dir", "/b/build/pdfium/.recipe_runtime/tmpOgwx97/build_data", "--buildbot-buildername", "linux_xfa", "--buildbot-mastername", "tryserver.client.pdfium", "--buildbot-botname", "bot1386-c4"], "cmdline": "['python3', '/path/to/ninja.py', '-C', Path('checkout','out','debug_xfa_v8'), '-j', 80]", "exit": 0, "env": {"GOMA_SERVICE_ACCOUNT_JSON_FILE": "/creds/service_accounts/service-account-goma-client.json", "BUILDBOT_BUILDERNAME": "linux_xfa", "USER": "chrome-bot", "HOME": "/home/chrome-bot", "BOTO_CONFIG": "/b/build/scripts/../../site_config/.boto", "PATH": "/home/chrome-bot/bin:/b/depot_tools:/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin", "PYTHONUNBUFFERED": "1", "BUILDBOT_BUILDBOTURL": "https://build.chromium.org/p/tryserver.client.pdfium/", "DISPLAY": ":0.0", "LANG": "en_US.UTF-8", "BUILDBOT_BLAMELIST": "[u'dsinclair@chromium.org']", "BUILDBOT_MASTERNAME": "tryserver.client.pdfium", "GOMACTL_CRASH_REPORT_ID_FILE": "/b/build/pdfium/.recipe_runtime/tmpOgwx97/build_data/crash_report_id_file", "USERNAME": "chrome-bot", "BUILDBOT_GOT_REVISION": "None", "PYTHONPATH": "/b/build/site_config:/b/build/scripts:/b/build/scripts/release:/b/build/third_party:/b/build/third_party/requests_2_10_0:/b/build_internal/site_config:/b/build_internal/symsrc:/b/build:/b/build/third_party/buildbot_8_4:/b/build/third_party/twisted_10_2:", "BUILDBOT_SCHEDULER": "None", "BUILDBOT_REVISION": "", "AWS_CREDENTIAL_FILE": "/b/build/scripts/../../site_config/.boto", "CHROME_HEADLESS": "1", "BUILDBOT_BRANCH": "", "GIT_USER_AGENT": "linux2 git/2.10.2 bot1386-c4.c.chromecompute.google.com.internal", "TESTING_BOTNAME": "bot1386-c4", "GOMA_DUMP_STATS_FILE": "/b/build/pdfium/.recipe_runtime/tmpOgwx97/build_data/goma_stats_proto", "BUILDBOT_BUILDNUMBER": "2937", "PWD": "/b/build/pdfium/build", "BUILDBOT_BOTNAME": "bot1386-c4", "BUILDBOT_CLOBBER": "", "PAGER": "cat"}, "compiler_proxy_info": "/tmp/compiler_proxy.bot1386-c4.chrome-bot.log.INFO.20161121-165459.5790", "cwd": "/b/build/pdfium/build", "compiler": "unknown"}
`))
	if err != nil {
		t.Errorf(`Parse()=_, %#v; want=_, <nil>`, err)
	}

	if njl.Metadata.Error == "" {
		t.Errorf("Parse().Metadata.Error='', want some error")
	}
	njl.Metadata = Metadata{}

	want := &NinjaLog{
		Filename: ".ninja_log",
		Start:    4,
		Steps:    stepsTestCase,
	}
	if diff := cmp.Diff(want, njl); diff != "" {
		t.Errorf("Parse() mismatch (-want, +got):\n%s", diff)
	}
}

func TestDump(t *testing.T) {
	var b bytes.Buffer
	err := Dump(&b, stepsTestCase)
	if err != nil {
		t.Errorf("Dump()=%v; want=<nil>", err)
	}
	if b.String() != logTestCase {
		t.Errorf("Dump %q; want %q", b.String(), logTestCase)
	}
}

func TestDedup(t *testing.T) {
	steps := append([]Step{}, stepsTestCase...)
	for _, out := range []string{
		"gen/ui/keyboard/webui/keyboard.mojom.cc",
		"gen/ui/keyboard/webui/keyboard.mojom.h",
		"gen/ui/keyboard/webui/keyboard.mojom.js",
		"gen/ui/keyboard/webui/keyboard.mojom-internal.h",
	} {
		steps = append(steps, Step{
			Start:   302 * time.Millisecond,
			End:     5764 * time.Millisecond,
			Out:     out,
			CmdHash: "a551cc46f8c21e5a",
		})
	}
	got := Dedup(steps)
	want := append([]Step{}, stepsSorted...)
	want = append(want, Step{
		Start: 302 * time.Millisecond,
		End:   5764 * time.Millisecond,
		Out:   "gen/ui/keyboard/webui/keyboard.mojom-internal.h",
		Outs: []string{
			"gen/ui/keyboard/webui/keyboard.mojom.cc",
			"gen/ui/keyboard/webui/keyboard.mojom.h",
			"gen/ui/keyboard/webui/keyboard.mojom.js",
		},
		CmdHash: "a551cc46f8c21e5a",
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Dedup=%v; want=%v", got, want)
	}
}

func TestFlow(t *testing.T) {
	steps := append([]Step{}, stepsTestCase...)
	steps = append(steps, Step{
		Start:   187 * time.Millisecond,
		End:     21304 * time.Millisecond,
		Out:     "obj/third_party/pdfium/core/src/fpdfdoc/fpdfdoc.doc_formfield.o",
		CmdHash: "2ac7111aa1ae86af",
	})

	flow := Flow(steps, false)

	wantSortByStart := [][]Step{
		{
			{
				Start:   76 * time.Millisecond,
				End:     187 * time.Millisecond,
				Out:     "resources/inspector/devtools_extension_api.js",
				CmdHash: "75430546595be7c2",
			},
			{
				Start:   187 * time.Millisecond,
				End:     21304 * time.Millisecond,
				Out:     "obj/third_party/pdfium/core/src/fpdfdoc/fpdfdoc.doc_formfield.o",
				CmdHash: "2ac7111aa1ae86af",
			},
		},
		{
			{
				Start:   78 * time.Millisecond,
				End:     286 * time.Millisecond,
				Out:     "gen/angle/commit_id.py",
				CmdHash: "4ede38e2c1617d8c",
			},
			{
				Start:   287 * time.Millisecond,
				End:     290 * time.Millisecond,
				Out:     "obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp",
				CmdHash: "b211d373de72f455",
			},
		},
		{
			{
				Start:   79 * time.Millisecond,
				End:     287 * time.Millisecond,
				Out:     "gen/angle/copy_compiler_dll.bat",
				CmdHash: "9fb635ad5d2c1109",
			},
		},
		{
			{
				Start:   80 * time.Millisecond,
				End:     284 * time.Millisecond,
				Out:     "gen/autofill_regex_constants.cc",
				CmdHash: "fa33c8d7ce1d8791",
			},
		},
		{
			{
				Start:   141 * time.Millisecond,
				End:     287 * time.Millisecond,
				Out:     "PepperFlash/manifest.json",
				CmdHash: "324f0a0b77c37ef",
			},
		},
		{
			{
				Start:   142 * time.Millisecond,
				End:     288 * time.Millisecond,
				Out:     "PepperFlash/libpepflashplayer.so",
				CmdHash: "1e2c2b7845a4d4fe",
			},
		},
	}

	if !reflect.DeepEqual(flow, wantSortByStart) {
		t.Errorf("Flow()=\n%#v\n want=\n%#v", flow, wantSortByStart)
	}

	flow = Flow(steps, true)
	wantSortByEnd := [][]Step{
		{
			{
				Start:   76 * time.Millisecond,
				End:     187 * time.Millisecond,
				Out:     "resources/inspector/devtools_extension_api.js",
				CmdHash: "75430546595be7c2",
			},
			{
				Start:   187 * time.Millisecond,
				End:     21304 * time.Millisecond,
				Out:     "obj/third_party/pdfium/core/src/fpdfdoc/fpdfdoc.doc_formfield.o",
				CmdHash: "2ac7111aa1ae86af",
			},
		},
		{
			{
				Start:   141 * time.Millisecond,
				End:     287 * time.Millisecond,
				Out:     "PepperFlash/manifest.json",
				CmdHash: "324f0a0b77c37ef",
			},
			{
				Start:   287 * time.Millisecond,
				End:     290 * time.Millisecond,
				Out:     "obj/third_party/angle/src/copy_scripts.actions_rules_copies.stamp",
				CmdHash: "b211d373de72f455",
			},
		},
		{
			{
				Start:   142 * time.Millisecond,
				End:     288 * time.Millisecond,
				Out:     "PepperFlash/libpepflashplayer.so",
				CmdHash: "1e2c2b7845a4d4fe",
			},
		},
		{
			{
				Start:   79 * time.Millisecond,
				End:     287 * time.Millisecond,
				Out:     "gen/angle/copy_compiler_dll.bat",
				CmdHash: "9fb635ad5d2c1109",
			},
		},
		{
			{
				Start:   78 * time.Millisecond,
				End:     286 * time.Millisecond,
				Out:     "gen/angle/commit_id.py",
				CmdHash: "4ede38e2c1617d8c",
			},
		},
		{
			{
				Start:   80 * time.Millisecond,
				End:     284 * time.Millisecond,
				Out:     "gen/autofill_regex_constants.cc",
				CmdHash: "fa33c8d7ce1d8791",
			},
		},
	}

	if !reflect.DeepEqual(flow, wantSortByEnd) {
		t.Errorf("Flow()=\n%#v\n want=\n%#v", flow, wantSortByEnd)
	}
}

func TestWeightedTime(t *testing.T) {
	steps := []Step{
		{
			Start:   0 * time.Millisecond,
			End:     3 * time.Millisecond,
			Out:     "target-a",
			CmdHash: "hash-target-a",
		},
		{
			Start:   2 * time.Millisecond,
			End:     5 * time.Millisecond,
			Out:     "target-b",
			CmdHash: "hash-target-b",
		},
		{
			Start:   2 * time.Millisecond,
			End:     8 * time.Millisecond,
			Out:     "target-c",
			CmdHash: "hash-target-c",
		},
		{
			Start:   2 * time.Millisecond,
			End:     3 * time.Millisecond,
			Out:     "target-d",
			CmdHash: "hash-target-d",
		},
	}

	// 0 1 2 3 4 5 6 7 8
	// +-+-+-+-+-+-+-+-+
	// <--A-->
	//     <--B-->
	//     <------C---->
	//     <D>
	got := WeightedTime(steps)
	want := map[string]time.Duration{
		"target-a": 2*time.Millisecond + 1*time.Millisecond/4,
		"target-b": 1*time.Millisecond/4 + 2*time.Millisecond/2,
		"target-c": 1*time.Millisecond/4 + 2*time.Millisecond/2 + 3*time.Millisecond,
		"target-d": 1 * time.Millisecond / 4,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WeightedTime(%v)=%v; want=%v", steps, got, want)
	}
}

func TestGetJobs(t *testing.T) {
	var getjobsTests = []struct {
		in        []string
		wantJobs  int
		wantError bool
	}{
		{
			in:       []string{"ninja", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "-j50"},
			wantJobs: 50,
		},
		{
			in:       []string{"ninja", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "-j", "100"},
			wantJobs: 100,
		},
		{
			in:       []string{"ninja", "-C", "/b/build/Linux_x64-j/build/src/out/Release", "all", "200-j"},
			wantJobs: 0,
		},
		{
			in:       []string{"ninja", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "50"},
			wantJobs: 0,
		},
		{
			in:        []string{"ninja", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "-j"},
			wantError: true,
		},
		{
			in:        []string{"ninja", "-C", "/b/build/Linux_x64/build/src/out/Release", "all", "-j", "abc"},
			wantError: true,
		},
	}
	for _, cmdline := range getjobsTests {
		got, err := getJobs(cmdline.in)
		if err != nil && !cmdline.wantError {
			t.Errorf("getjobs()=%v; want=<nil>", err)
		}
		if got != cmdline.wantJobs {
			t.Errorf("getJobs(%v)=%v; want=%v", cmdline, got, cmdline.wantJobs)
		}
	}
}

func TestParseMetadata(t *testing.T) {
	var m Metadata
	json := `{"jobs": 1000, "platform": "Linux", "cpu_core": 48, "targets": ["chrome"], "build_configs": {"use_goma": "true", "target_cpu": "\"\"", "is_component_build": "true", "symbol_level": "-1", "is_debug": "false", "enable_nacl": "false", "host_cpu": "\"x64\"", "host_os": "\"linux\"", "target_os": "\"\""}}`

	err := parseMetadata([]byte(json), &m)

	if err != nil {
		t.Errorf("failed to parse medatadata %q: %v", json, err)
	}

	want := Metadata{
		Platform: "Linux",
		CPUCore:  48,
		BuildConfigs: map[string]string{
			"use_goma":           "true",
			"is_component_build": "true",
			"enable_nacl":        "false",
			"host_cpu":           "\"x64\"",
			"target_os":          "\"\"",
			"target_cpu":         "\"\"",
			"symbol_level":       "-1",
			"is_debug":           "false",
			"host_os":            "\"linux\"",
		},
		Jobs:    1000,
		Targets: []string{"chrome"},
	}

	if !reflect.DeepEqual(m, want) {
		t.Errorf("parseMetadata(%q, ...): got %#v, want %#v", json, m, want)
	}
}

func TestGetTargets(t *testing.T) {
	type testData struct {
		m    *Metadata
		want []string
	}
	tests := []testData{
		{
			m:    &Metadata{Targets: []string{"foo", "/path/to/bar"}},
			want: []string{"foo"},
		},
		{
			m:    &Metadata{Cmdline: []string{"python3", "/path/to/ninja.py", "-C", "/path/to/out/Default", "foo", "/path/to/bar"}},
			want: []string{"foo"},
		},
		{
			m:    &Metadata{Targets: []string{"foo", "C:\\path\\to\\bar"}},
			want: []string{"foo"},
		},
		{
			m:    &Metadata{Cmdline: []string{"python3.exe", "C:\\path\\to\\ninja.py", "-C", "C:\\path\\to\\out\\Default", "foo", "C:\\path\\to\\bar"}},
			want: []string{"foo"},
		},
	}
	for _, test := range tests {
		if diff := cmp.Diff(test.want, test.m.getTargets()); diff != "" {
			t.Errorf("getTargets() mismatch (-want, +got):\n%s", diff)
		}
	}
}

//go:embed testdata/ninja_log
var ninjaLogData []byte

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Parse(".ninja_log", bytes.NewReader(ninjaLogData))
		if err != nil {
			b.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
		}
	}
}

func BenchmarkDedup(b *testing.B) {
	njl, err := Parse(".ninja_log", bytes.NewReader(ninjaLogData))
	if err != nil {
		b.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		steps := make([]Step, len(njl.Steps))
		copy(steps, njl.Steps)
		Dedup(steps)
	}
}

func BenchmarkFlow(b *testing.B) {
	njl, err := Parse(".ninja_log", bytes.NewReader(ninjaLogData))
	if err != nil {
		b.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
	}
	steps := Dedup(njl.Steps)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flowInput := make([]Step, len(steps))
		copy(flowInput, steps)
		Flow(flowInput, false)
	}
}

func BenchmarkToTraces(b *testing.B) {
	njl, err := Parse(".ninja_log", bytes.NewReader(ninjaLogData))
	if err != nil {
		b.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
	}
	steps := Dedup(njl.Steps)
	flow := Flow(steps, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToTraces(flow, 1)
	}
}

func BenchmarkDedupFlowToTraces(b *testing.B) {
	for i := 0; i < b.N; i++ {
		njl, err := Parse(".ninja_log", bytes.NewReader(ninjaLogData))
		if err != nil {
			b.Errorf(`Parse()=_, %v; want=_, <nil>`, err)
		}

		steps := Dedup(njl.Steps)
		flow := Flow(steps, false)
		ToTraces(flow, 1)
	}
}
