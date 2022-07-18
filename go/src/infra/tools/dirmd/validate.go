// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmd

// ValidateFile returns a non-nil error if the metadata file is invalid.
//
// A valid file has a base filename "DIR_METADATA" or "OWNERS".
// The format of its contents correspond to the base name.
func ValidateFile(fileName string) error {
	_, err := ParseFile(fileName)
	return err
}
