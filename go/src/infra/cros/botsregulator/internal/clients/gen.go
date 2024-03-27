// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clients

//go:generate mockgen -source ufs.go -destination ufs.mock.go -package clients -write_package_comment=false
//go:generate mockgen -source gcep.go -destination gcep.mock.go -package clients -write_package_comment=false
