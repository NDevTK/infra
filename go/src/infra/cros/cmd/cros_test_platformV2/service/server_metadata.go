// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Plain Old Go Object for persisting Server information
package service

// ServerMetadata stores server specific information
type ServerMetadata struct {
	Port                      int
	ServiceMetadataExportPath string
	LogPath                   string
}
