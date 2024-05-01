// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v2"

	. "go.chromium.org/luci/common/testing/assertions"
)

func TestManifest(t *testing.T) {
	t.Parallel()

	load := func(body, path string) (*Manifest, error) {
		m, err := parse(strings.NewReader(body), filepath.FromSlash(path))
		if err != nil {
			return nil, err
		}
		return m, m.Finalize()
	}

	Convey("Minimal", t, func() {
		m, err := load("name: zzz\ncontextdir: ../../../blarg/", "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m, ShouldResemble, &Manifest{
			Name:        "zzz",
			ManifestDir: filepath.FromSlash("root/1/2/3/4"),
			ContextDir:  filepath.FromSlash("root/1/blarg"),
			InputsDir:   filepath.FromSlash("root/1/blarg"),
			Sources:     []string{filepath.FromSlash("root/1/blarg")},
		})
	})

	Convey("No name", t, func() {
		_, err := load("", "some/dir")
		So(err, ShouldErrLike, `bad "name" field: can't be empty, it's required`)
	})

	Convey("Bad name", t, func() {
		_, err := load(`name: cheat:tag`, "some/dir")
		So(err, ShouldErrLike, `bad "name" field: "cheat:tag" contains forbidden symbols (any of "\\:@")`)
	})

	Convey("Not yaml", t, func() {
		_, err := load(`im not a YAML`, "")
		So(err, ShouldErrLike, "unmarshal errors")
	})

	Convey("Deriving contextdir from dockerfile", t, func() {
		m, err := load("name: zzz\ndockerfile: ../../../blarg/Dockerfile", "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m, ShouldResemble, &Manifest{
			Name:        "zzz",
			ManifestDir: filepath.FromSlash("root/1/2/3/4"),
			Dockerfile:  filepath.FromSlash("root/1/blarg/Dockerfile"),
			ContextDir:  filepath.FromSlash("root/1/blarg"),
			InputsDir:   filepath.FromSlash("root/1/blarg"),
			Sources:     []string{filepath.FromSlash("root/1/blarg")},
		})
	})

	Convey("Resolving imagepins", t, func() {
		m, err := load("name: zzz\ncontextdir: .\nimagepins: ../../../blarg/pins.yaml", "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m, ShouldResemble, &Manifest{
			Name:        "zzz",
			ManifestDir: filepath.FromSlash("root/1/2/3/4"),
			ContextDir:  filepath.FromSlash("root/1/2/3/4"),
			InputsDir:   filepath.FromSlash("root/1/2/3/4"),
			Sources:     []string{filepath.FromSlash("root/1/2/3/4")},
			ImagePins:   filepath.FromSlash("root/1/blarg/pins.yaml"),
		})
	})

	Convey("Empty build step", t, func() {
		_, err := load(`{"name": "zzz", "contextdir": ".", "build": [
			{"dest": "zzz"}
		]}`, "root/1/2/3/4")
		So(err, ShouldErrLike, "bad build step #1: unrecognized or empty")
	})

	Convey("Ambiguous build step", t, func() {
		_, err := load(`{"name": "zzz", "contextdir": ".", "build": [
			{"copy": "zzz", "go_binary": "zzz"}
		]}`, "root/1/2/3/4")
		So(err, ShouldErrLike, "bad build step #1: ambiguous")
	})

	Convey("CopyBuildStep", t, func() {
		m, err := load(`{"name": "zzz", "contextdir": "ctx", "build": [
				{"copy": "${manifestdir}/../../../blarg/zzz"}
			]}`, "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m.Build, ShouldHaveLength, 1)
		So(m.Build[0].Dest, ShouldEqual, filepath.FromSlash("root/1/2/3/4/ctx/zzz"))
		So(m.Build[0].Concrete(), ShouldResemble, &CopyBuildStep{
			Copy: filepath.FromSlash("root/1/blarg/zzz"),
		})
	})

	Convey("GoBuildStep", t, func() {
		m, err := load(`{"name": "zzz", "contextdir": "ctx", "build": [
				{"go_binary": "go.pkg/some/tool"}
			]}`, "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m.Build, ShouldHaveLength, 1)
		So(m.Build[0].Dest, ShouldEqual, filepath.FromSlash("root/1/2/3/4/ctx/tool"))
		So(m.Build[0].Cwd, ShouldEqual, filepath.FromSlash("root/1/2/3/4/ctx"))
		So(m.Build[0].Concrete(), ShouldResemble, &GoBuildStep{
			GoBinary: "go.pkg/some/tool",
		})
	})

	Convey("RunBuildStep", t, func() {
		m, err := load(`{"name": "zzz", "contextdir": "ctx", "build": [
				{"run": ["a", "b"]}
			]}`, "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m.Build, ShouldHaveLength, 1)
		So(m.Build[0].Cwd, ShouldEqual, filepath.FromSlash("root/1/2/3/4/ctx"))
		So(m.Build[0].Concrete(), ShouldResemble, &RunBuildStep{
			Run: []string{"a", "b"},
		})
	})

	Convey("GoGAEBundleBuildStep", t, func() {
		m, err := load(`{"name": "zzz", "contextdir": "ctx", "inputsdir": "in", "build": [
				{"go_gae_bundle": "${inputsdir}/pkg", "dest": "${contextdir}/pkg"}
			]}`, "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m.Build, ShouldHaveLength, 1)
		So(m.Build[0].Concrete(), ShouldResemble, &GoGAEBundleBuildStep{
			GoGAEBundle: filepath.FromSlash("root/1/2/3/4/in/pkg"),
		})
		So(m.Build[0].Dest, ShouldEqual, filepath.FromSlash("root/1/2/3/4/ctx/pkg"))
	})

	Convey("Good infra", t, func() {
		m, err := load(`{"name": "zzz", "contextdir": ".", "infra": {
			"infra1": {
				"storage": "gs://bucket",
				"notify": [
					{
						"kind": "git",
						"repo": "https://repo.example.com",
						"script": "some/script.py"
					}
				]
			},
			"infra2": {
				"storage": "gs://bucket/path"
			}
		}}`, "root/1/2/3/4")
		So(err, ShouldBeNil)
		So(m.Infra, ShouldResemble, map[string]Infra{
			"infra1": {
				Storage: "gs://bucket",
				Notify: []NotifyConfig{
					{
						Kind:   "git",
						Repo:   "https://repo.example.com",
						Script: "some/script.py",
					},
				},
			},
			"infra2": {Storage: "gs://bucket/path"},
		})
	})

	Convey("Unsupported storage", t, func() {
		_, err := load(`{"name": "zzz", "contextdir": ".", "infra": {
			"infra1": {"storage": "ftp://bucket"}
		}}`, "root/1/2/3/4")
		So(err, ShouldErrLike, `in infra section "infra1": bad storage "ftp://bucket", only gs:// is supported currently`)
	})

	Convey("No bucket in storage", t, func() {
		_, err := load(`{"name": "zzz", "contextdir": ".", "infra": {
			"infra1": {"storage": "gs:///zzz"}
		}}`, "root/1/2/3/4")
		So(err, ShouldErrLike, `in infra section "infra1": bad storage "gs:///zzz", bucket name is missing`)
	})

	Convey("Bad notify", t, func() {
		check := func(notify string) error {
			_, err := load(fmt.Sprintf(`{"name": "zzz", "contextdir": ".", "infra": {
			"infra1": {"notify": [%s]}
		}}`, notify), "root/1/2/3/4")
			return err
		}

		Convey("Bad kind", func() {
			So(check(`{"kind": "bad"}`), ShouldErrLike, `unsupported notify kind`)
		})

		Convey("Bad repo", func() {
			So(check(`{"kind": "git", "repo": "ftp://zzz"}`),
				ShouldErrLike, `should be an https:// URL`)
		})

		Convey("Bad script path", func() {
			for _, p := range []string{"a/../b", "", "a/./b"} {
				So(check(fmt.Sprintf(`{"kind": "git", "repo": "https://zzz", "script": "%s"}`, p)),
					ShouldErrLike, `should be a normalized slash-separate path`)
			}
		})

		Convey("Not relative path", func() {
			for _, p := range []string{"/", "/a/b"} {
				So(check(fmt.Sprintf(`{"kind": "git", "repo": "https://zzz", "script": "%s"}`, p)),
					ShouldErrLike, `not a path inside the repo`)
			}
		})
	})
}

func TestExtends(t *testing.T) {
	t.Parallel()

	Convey("With temp dir", t, func() {
		dir, err := ioutil.TempDir("", "cloudbuildhelper")
		So(err, ShouldBeNil)
		Reset(func() { os.RemoveAll(dir) })

		write := func(path string, m Manifest) {
			blob, err := yaml.Marshal(&m)
			So(err, ShouldBeNil)
			p := filepath.Join(dir, filepath.FromSlash(path))
			So(os.MkdirAll(filepath.Dir(p), 0777), ShouldBeNil)
			So(ioutil.WriteFile(p, blob, 0666), ShouldBeNil)
		}

		abs := func(path string) string {
			p, err := filepath.Abs(filepath.Join(dir, filepath.FromSlash(path)))
			So(err, ShouldBeNil)
			return p
		}

		Convey("Works", func() {
			var falseVal = false

			notifyBase := NotifyConfig{
				Kind:   "git",
				Repo:   "https://base.example.com",
				Script: "base",
			}
			notifyMid := NotifyConfig{
				Kind:   "git",
				Repo:   "https://mid.example.com",
				Script: "mid",
			}

			write("base.yaml", Manifest{
				Name:      "base",
				ImagePins: "pins.yaml",
				Sources:   []string{"base-1", "base-2"},
				Infra: map[string]Infra{
					"base": {
						Storage:  "gs://base-storage",
						Registry: "base-registry",
						Notify:   []NotifyConfig{notifyBase},
					},
				},
				Build: []*BuildStep{
					{CopyBuildStep: CopyBuildStep{Copy: "${manifestdir}/manifest_base.copy"}},
					{CopyBuildStep: CopyBuildStep{Copy: "${contextdir}/context_base.copy"}},
				},
			})

			write("deeper/mid.yaml", Manifest{
				Name:          "mid",
				Extends:       "../base.yaml",
				Deterministic: &falseVal,
				Sources:       []string{"mid-1", "../mid-2", "../base-2"},
				Infra: map[string]Infra{
					"mid": {
						Storage:  "gs://mid-storage",
						Registry: "mid-registry",
						CloudBuild: map[string]CloudBuildBuilder{
							"builder1": {Project: "project-1"},
							"builder2": {Project: "project-2"},
						},
						Notify: []NotifyConfig{notifyMid},
					},
				},
				Build: []*BuildStep{
					{CopyBuildStep: CopyBuildStep{Copy: "${manifestdir}/manifest_mid.copy"}},
					{CopyBuildStep: CopyBuildStep{Copy: "${contextdir}/context_mid.copy"}},
				},
			})

			write("deeper/leaf.yaml", Manifest{
				Name:       "leaf",
				Extends:    "mid.yaml",
				Dockerfile: "dockerfile",
				ContextDir: "context-dir",
				InputsDir:  "inputs-dir",
				Infra: map[string]Infra{
					"mid": { // partial override
						Registry: "leaf-registry",
						CloudBuild: map[string]CloudBuildBuilder{
							"builder1": {Project: "project-1-override"},
							"builder3": {Project: "project-3"},
						},
					},
				},
				Build: []*BuildStep{
					{CopyBuildStep: CopyBuildStep{Copy: "${manifestdir}/manifest_leaf.copy"}},
					{CopyBuildStep: CopyBuildStep{Copy: "${contextdir}/context_leaf.copy"}},
				},
			})

			m, err := Load(filepath.Join(dir, "deeper", "leaf.yaml"))
			So(err, ShouldBeNil)
			So(m.Finalize(), ShouldBeNil)

			// We'll deal with them separately below.
			steps := m.Build
			m.Build = nil

			So(m, ShouldResemble, &Manifest{
				Name:        "leaf",
				ManifestDir: abs("deeper"),
				Dockerfile:  abs("deeper/dockerfile"),
				ContextDir:  abs("deeper/context-dir"),
				InputsDir:   abs("deeper/inputs-dir"),
				Sources: []string{
					abs("deeper/mid-1"),
					abs("mid-2"),
					abs("base-2"),
					abs("base-1"),
				},
				ImagePins:     abs("pins.yaml"),
				Deterministic: &falseVal,
				Infra: map[string]Infra{
					"base": {
						Storage:    "gs://base-storage",
						Registry:   "base-registry",
						Notify:     []NotifyConfig{notifyBase},
						CloudBuild: map[string]CloudBuildBuilder{},
					},
					"mid": {
						Storage:  "gs://mid-storage",
						Registry: "leaf-registry",
						CloudBuild: map[string]CloudBuildBuilder{
							"builder1": {Project: "project-1-override", Args: []string{}},
							"builder2": {Project: "project-2", Args: []string{}},
							"builder3": {Project: "project-3", Args: []string{}},
						},
						Notify: []NotifyConfig{notifyMid},
					},
				},
			})

			var copySrc []string
			for _, s := range steps {
				copySrc = append(copySrc, s.Copy)
			}
			So(copySrc, ShouldResemble, []string{
				abs("manifest_base.copy"),
				abs("deeper/context-dir/context_base.copy"),
				abs("deeper/manifest_mid.copy"),
				abs("deeper/context-dir/context_mid.copy"),
				abs("deeper/manifest_leaf.copy"),
				abs("deeper/context-dir/context_leaf.copy"),
			})
		})

		Convey("Recursion", func() {
			write("a.yaml", Manifest{Name: "a", Extends: "b.yaml"})
			write("b.yaml", Manifest{Name: "b", Extends: "a.yaml"})

			_, err := Load(filepath.Join(dir, "a.yaml"))
			So(err, ShouldErrLike, "too much nesting")
		})

		Convey("Deep error", func() {
			write("a.yaml", Manifest{Name: "a", Extends: "b.yaml"})
			write("b.yaml", Manifest{
				Name: "b",
				Infra: map[string]Infra{
					"base": {Storage: "bad url"},
				},
			})

			_, err := Load(filepath.Join(dir, "a.yaml"))
			So(err, ShouldErrLike, `bad storage`)
		})
	})
}

func TestRenderPath(t *testing.T) {
	t.Parallel()

	Convey("Works", t, func() {
		out, err := renderPath("var", "${a}", map[string]string{"a": "zzz"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "zzz")

		out, err = renderPath("var", "${a}/", map[string]string{"a": "zzz"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "zzz")

		out, err = renderPath("var", "${a}/.", map[string]string{"a": "zzz"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, "zzz")

		out, err = renderPath("var", "${a}/b/c", map[string]string{"a": "zzz"})
		So(err, ShouldBeNil)
		So(out, ShouldEqual, filepath.FromSlash("zzz/b/c"))
	})

	Convey("Errors", t, func() {
		_, err := renderPath("var", ".", map[string]string{"a": "zzz", "b": "yyy"})
		So(err, ShouldErrLike, "must start with ${a} or ${b}")

		_, err = renderPath("var", "${c}", map[string]string{"a": "zzz", "b": "yyy"})
		So(err, ShouldErrLike, "unknown dir variable ${c}, expecting ${a} or ${b}")
	})
}
