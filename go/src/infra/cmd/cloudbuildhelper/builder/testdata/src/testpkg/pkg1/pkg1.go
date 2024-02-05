// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pkg1

import (
	"embed"
	"testpkg/pkg2"
)

//go:embed all:embedded
var fs embed.FS

// A exists to make golint happy.
const A = pkg2.A
