// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/common/logging/gologger"

	"infra/cmd/cloudbuildhelper/bundledesc"
	"infra/cmd/cloudbuildhelper/fileset"
	"infra/cmd/cloudbuildhelper/manifest"
)

func init() {
	// Our test module has "vendor" directory and has go >=1.14 in go.mod, so
	// "-mod=vendor" is the mode that Go should be picking. But on CI builder
	// GOFLAGS may override it to "-mod=readonly" which breaks the test. Set it
	// explicitly.
	os.Setenv("GOFLAGS", "-mod=vendor")
}

func TestBuilder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = gologger.StdConfig.Use(ctx)

	Convey("With temp dir", t, func() {
		srcDir, err := filepath.Abs("testdata")
		So(err, ShouldBeNil)

		tmpDir, err := os.MkdirTemp("", "builder_test")
		So(err, ShouldBeNil)
		Reset(func() { os.RemoveAll(tmpDir) })

		b, err := New()
		So(err, ShouldBeNil)
		defer b.Close()

		put := func(path, body string) {
			fp := filepath.Join(tmpDir, filepath.FromSlash(path))
			So(os.MkdirAll(filepath.Dir(fp), 0777), ShouldBeNil)
			So(os.WriteFile(fp, []byte(body), 0666), ShouldBeNil)
		}

		build := func(manifestBody string) (*fileset.Set, error) {
			manifestPath := filepath.Join(tmpDir, "manifest.yaml")
			So(os.WriteFile(manifestPath, []byte(manifestBody), 0600), ShouldBeNil)
			loaded, err := manifest.Load(manifestPath)
			So(err, ShouldBeNil)
			So(loaded.Finalize(), ShouldBeNil)
			return b.Build(ctx, loaded)
		}

		Convey("ContextDir only", func() {
			put("ctx/f1", "file 1")
			put("ctx/f2", "file 2")

			out, err := build(`{
				"name": "test",
				"contextdir": "ctx"
			}`)
			So(err, ShouldBeNil)
			So(out.Files(), ShouldHaveLength, 2)

			So(b.Close(), ShouldBeNil)
			So(b.Close(), ShouldBeNil) // idempotent
		})

		Convey("A bunch of steps", func() {
			put("ctx/f1", "file 1")
			put("ctx/f2", "file 2")

			put("copy/f1", "overridden")
			put("copy/dir/f", "f")

			out, err := build(fmt.Sprintf(`{
				"name": "test",
				"contextdir": "ctx",
				"inputsdir": %q,
				"build": [
					{
						"copy": "${manifestdir}/copy",
						"dest": "${contextdir}"
					},
					{
						"go_binary": "testpkg/helloworld",
						"dest": "${contextdir}/gocmd",
					},
					{
						"run": [
							"go",
							"run",
							"testpkg/helloworld",
							"${contextdir}/say_hi"
						],
						"outputs": ["${contextdir}/say_hi"]
					}
				]
			}`, filepath.Join(srcDir, "src", "testpkg")))
			So(err, ShouldBeNil)

			names := make([]string, out.Len())
			byName := make(map[string]fileset.File, out.Len())
			for i, f := range out.Files() {
				names[i] = f.Path
				byName[f.Path] = f
			}
			So(names, ShouldResemble, []string{
				"dir", "dir/f", "f1", "f2", "gocmd", "say_hi",
			})

			r, err := byName["f1"].Body()
			So(err, ShouldBeNil)
			blob, err := io.ReadAll(r)
			So(err, ShouldBeNil)
			So(string(blob), ShouldEqual, "overridden")
		})

		Convey("Go GAE bundling", func() {
			// To test .gitignore handling, create a gitignored file manually, since
			// we can't check it in.
			err := os.WriteFile(filepath.FromSlash("testdata/src/testpkg/helloworld/static/ignored"), nil, 0600)
			So(err, ShouldBeNil)

			buildBundle := func(manifestPath string) ([]string, map[string]*fileset.File) {
				m, err := manifest.Load(manifestPath)
				So(err, ShouldBeNil)
				m.ContextDir = tmpDir
				So(m.Finalize(), ShouldBeNil)

				out, err := b.Build(ctx, m)
				So(err, ShouldBeNil)

				files := make([]string, 0, out.Len())
				byName := make(map[string]*fileset.File, out.Len())
				for _, f := range out.Files() {
					if !f.Directory {
						files = append(files, f.Path)
						cpy := f
						byName[f.Path] = &cpy
					}
				}

				return files, byName
			}

			Convey("GOPATH bundle", func() {
				files, byName := buildBundle(filepath.FromSlash("testdata/src/testpkg/gaebundle_gopath.yaml"))

				So(files, ShouldResemble, []string{
					".cloudbuildhelper.json",
					"_gopath/goenv",
					"_gopath/src/example.com/another/another_a.go",
					"_gopath/src/example.com/pkg/pkg_a.go",
					"_gopath/src/testpkg/helloworld/.gcloudignore",
					"_gopath/src/testpkg/helloworld/anotherpkg.go",
					"_gopath/src/testpkg/helloworld/buildflags_amd64.go",
					"_gopath/src/testpkg/helloworld/buildflags_linux.go",
					"_gopath/src/testpkg/helloworld/curgo.go",
					"_gopath/src/testpkg/helloworld/fake-app.yaml",
					"_gopath/src/testpkg/helloworld/main.go",
					"_gopath/src/testpkg/helloworld/static.txt",
					"_gopath/src/testpkg/helloworld/static/static.txt",
					"_gopath/src/testpkg/helloworld/vendor.go",
					"_gopath/src/testpkg/pkg1/embedded/_embedded",
					"_gopath/src/testpkg/pkg1/embedded/embedded",
					"_gopath/src/testpkg/pkg1/pkg1.go",
					"_gopath/src/testpkg/pkg1/vendor.go",
					"_gopath/src/testpkg/pkg2/pkg2.go",
					"helloworld",
				})

				So(byName["helloworld"], ShouldResemble, &fileset.File{
					Path:          "helloworld",
					SymlinkTarget: "_gopath/src/testpkg/helloworld",
				})

				desc, err := byName[".cloudbuildhelper.json"].ReadAll()
				So(err, ShouldBeNil)
				So(string(desc), ShouldEqual, fmt.Sprintf(`{
  "format_version": "%s",
  "go_gae_bundles": [
    {
      "app_yaml": "_gopath/src/testpkg/helloworld/fake-app.yaml"
    }
  ]
}`, bundledesc.FormatVersion))

			})

			Convey("Modules bundle", func() {
				files, byName := buildBundle(filepath.FromSlash("testdata/src/testpkg/gaebundle_modules.yaml"))

				So(files, ShouldResemble, []string{
					".cloudbuildhelper.json",
					"_gomod/go.mod",
					"_gomod/goenv",
					"_gomod/helloworld/.gcloudignore",
					"_gomod/helloworld/anotherpkg.go",
					"_gomod/helloworld/buildflags_amd64.go",
					"_gomod/helloworld/buildflags_linux.go",
					"_gomod/helloworld/curgo.go",
					"_gomod/helloworld/fake-app.yaml",
					"_gomod/helloworld/main.go",
					"_gomod/helloworld/static.txt",
					"_gomod/helloworld/static/static.txt",
					"_gomod/helloworld/vendor.go",
					"_gomod/pkg1/embedded/_embedded",
					"_gomod/pkg1/embedded/embedded",
					"_gomod/pkg1/pkg1.go",
					"_gomod/pkg1/vendor.go",
					"_gomod/pkg2/pkg2.go",
					"_gomod/vendor/example.com/another/another_a.go",
					"_gomod/vendor/example.com/pkg/pkg_a.go",
					"_gomod/vendor/modules.txt",
					"helloworld",
				})

				So(byName["helloworld"], ShouldResemble, &fileset.File{
					Path:          "helloworld",
					SymlinkTarget: "_gomod/helloworld",
				})

				desc, err := byName[".cloudbuildhelper.json"].ReadAll()
				So(err, ShouldBeNil)
				So(string(desc), ShouldEqual, fmt.Sprintf(`{
  "format_version": "%s",
  "go_gae_bundles": [
    {
      "app_yaml": "_gomod/helloworld/fake-app.yaml"
    }
  ]
}`, bundledesc.FormatVersion))

				appYaml, err := byName["_gomod/helloworld/fake-app.yaml"].ReadAll()
				So(err, ShouldBeNil)
				So(string(appYaml), ShouldEqual, `entrypoint: |
    cd helloworld && main -auth-service-host ${AUTH_SERVICE_HOST}
handlers:
    - static_files: helloworld/frontend/static/robots.txt
      upload: helloworld/frontend/static/robots.txt
      url: /robots.txt
    - secure: always
      static_dir: helloworld/frontend/static
      url: /static
    - expiration: 7d
      secure: always
      static_dir: helloworld/ui/out/immutable
      url: /ui/immutable
    - secure: always
      static_files: helloworld/ui/out/\1
      upload: helloworld/ui/out/root_sw\.js(\.map)?$
      url: /(root_sw\.js(\.map)?)$
    - script: auto
      secure: always
      url: /.*
runtime: go111
`)
			})
		})
	})
}
