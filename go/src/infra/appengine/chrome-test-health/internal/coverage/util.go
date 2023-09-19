// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"strings"

	structpb "github.com/golang/protobuf/ptypes/struct"
)

// getStructFromCompressedData decompresses zlib-compressed
// data and returns a struct for the decompressed data.
func getStructFromCompressedData(data []byte, s *structpb.Struct) error {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, r)
	if err != nil {
		return err
	}
	uncompressedStr := buf.String()
	err = json.Unmarshal([]byte(uncompressedStr), &s)
	if err != nil {
		return err
	}

	return nil
}

// compressString compresses the given str argument using
// zlib package.
func compressString(str string) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	_, err := w.Write([]byte(str))
	if err != nil {
		return []byte{}, err
	}
	w.Close()

	compressedBytes := b.Bytes()

	return compressedBytes, nil
}
