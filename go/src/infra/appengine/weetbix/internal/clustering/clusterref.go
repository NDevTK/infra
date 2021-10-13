// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clustering

// ClusterRef represents a reference to a cluster. The LUCI Project is
// omitted as it is assumed to be implicit from the context.
type ClusterRef struct {
	Algorithm string
	ID        []byte
}

// Key returns a value that can be used to uniquely identify the ClusterRef.
// This is designed for cases where it is desirable for cluster references
// to be used as keys in a map.
func (c *ClusterRef) Key() string {
	// string is simply a sequence of bytes.
	return c.Algorithm + ":" + string(c.ID)
}
