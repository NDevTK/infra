// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package coverage

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	structpb "github.com/golang/protobuf/ptypes/struct"
	. "github.com/smartystreets/goconvey/convey"
)

func getFakeCompressedString() (string, []byte) {
	str := `{"name":"foo","type":"foo1"}`
	strBytes, _ := compressString(str)
	return str, strBytes
}

func getFakeCompressedStringWithInvalidJson() (string, []byte) {
	str := `{"name":"foo","type":"foo1"`
	strBytes, _ := compressString(str)
	return str, strBytes
}

func TestGetStructFromCompressedData(t *testing.T) {
	t.Parallel()

	Convey(`Should be able to return decompressed data`, t, func() {
		Convey(`Not compressed in correct format`, func() {
			compressedMalformedData := []byte("malformed data")
			st := structpb.Struct{}
			err := getStructFromCompressedData(compressedMalformedData, &st)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("zlib: invalid header"))
		})
		Convey(`Compressed in correct format but invalid json`, func() {
			_, compressedInvalidJsonData := getFakeCompressedStringWithInvalidJson()
			st := structpb.Struct{}
			err := getStructFromCompressedData(compressedInvalidJsonData, &st)
			So(err, ShouldNotBeNil)
			So(fmt.Sprintf("%s", err), ShouldContainSubstring, "unexpected end of JSON input")
		})
		Convey(`Well formed data`, func() {
			str, compressedWellFormedData := getFakeCompressedString()
			st := structpb.Struct{}
			err := getStructFromCompressedData(compressedWellFormedData, &st)
			jsonStr, _ := json.Marshal(st.Fields)
			So(err, ShouldBeNil)
			So(string(jsonStr), ShouldResemble, str)
		})
	})
}
