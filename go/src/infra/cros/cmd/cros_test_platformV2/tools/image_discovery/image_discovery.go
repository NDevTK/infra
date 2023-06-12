// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package image_discovery

// Will be a helper to translate the given request into an image.
// might require bulling the container metadata from gcs for the given build.

// How will we handle the non-cft images, such as TTCP, or others?
// For now, we will blindly assume that a non-cft (ie TTCP, others) images
// will still publish their MD along with CFT.

func getContainerTag(version string, name string) (string, error) {
	return "", nil
}
