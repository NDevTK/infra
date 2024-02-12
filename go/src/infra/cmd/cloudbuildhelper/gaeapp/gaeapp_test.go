// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gaeapp

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const input = `service: default-go
runtime: go121

# Comment.
luci_gae_vars:
  app1:
    VAR1: val
    VAR2: "1.0"
  app2:
    VAR1: val
    VAR2: "1.0"

vpc_access_connector:
  name: ${VPC_CONNECTOR}

automatic_scaling:
  min_idle_instances: 1
  max_concurrent_requests: 80

inbound_services:
- warmup

instance_class: F4

entrypoint: >
  main
  -flag ${FLAG}
  -stuff stuff

handlers:
- url: /.*
  script: auto
  secure: always
- url: /static
  static_dir: static
- url: /(root_sw\.js(\.map)?)$
  secure: always
  static_files: ui/out/\1
  upload: ui/out/root_sw\.js(\.map)?$
`

const output = `automatic_scaling:
    max_concurrent_requests: 80
    min_idle_instances: 1
entrypoint: |
    main -flag ${FLAG} -stuff stuff
handlers:
    - script: auto
      secure: always
      url: /.*
    - static_dir: static
      url: /static
    - secure: always
      static_files: ui/out/\1
      upload: ui/out/root_sw\.js(\.map)?$
      url: /(root_sw\.js(\.map)?)$
inbound_services:
    - warmup
instance_class: F4
luci_gae_vars:
    app1:
        VAR1: val
        VAR2: "1.0"
    app2:
        VAR1: val
        VAR2: "1.0"
runtime: go121
service: default-go
vpc_access_connector:
    name: ${VPC_CONNECTOR}
`

func TestApp(t *testing.T) {
	t.Parallel()

	Convey("Marshaling", t, func() {
		app, err := LoadAppYAML([]byte(input))
		So(err, ShouldBeNil)
		out, err := app.Save()
		So(err, ShouldBeNil)
		fmt.Printf("%s\n", string(out))
		So(string(out), ShouldEqual, output)
	})

	Convey("Replacing string", t, func() {
		app, err := LoadAppYAML([]byte("entrypoint: abc"))
		So(err, ShouldBeNil)
		So(app.Entrypoint, ShouldEqual, "abc")
		app.Entrypoint = "def"
		blob, err := app.Save()
		So(err, ShouldBeNil)
		So(string(blob), ShouldEqual, "entrypoint: def\n")
	})

	Convey("Removing string", t, func() {
		app, err := LoadAppYAML([]byte("entrypoint: abc"))
		So(err, ShouldBeNil)
		So(app.Entrypoint, ShouldEqual, "abc")
		app.Entrypoint = ""
		blob, err := app.Save()
		So(err, ShouldBeNil)
		So(string(blob), ShouldEqual, "{}\n")
	})

	Convey("Modifying handlers", t, func() {
		app, err := LoadAppYAML([]byte(`handlers:
      - url: /.*
        script: auto
        secure: always
      - url: url1
        static_dir: static_dir
      - url: url2
        static_files: static_files
        upload: upload
    `))
		So(err, ShouldBeNil)

		for _, h := range app.Handlers {
			if h.StaticDir != "" {
				h.StaticDir += "-sfx1"
			}
			if h.StaticFiles != "" {
				h.StaticFiles += "-sfx2"
			}
			if h.Upload != "" {
				h.Upload += "-sfx3"
			}
		}

		blob, err := app.Save()
		So(err, ShouldBeNil)
		So(string(blob), ShouldEqual, `handlers:
    - script: auto
      secure: always
      url: /.*
    - static_dir: static_dir-sfx1
      url: url1
    - static_files: static_files-sfx2
      upload: upload-sfx3
      url: url2
`)
	})
}
