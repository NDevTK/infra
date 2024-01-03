// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package protoio contains helper methods for proto I/O done by the testplan
// tool.
package protoio

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/common/logging"
)

// ReadBinaryOrJSONPb reads path into m, attempting to parse as both a binary
// and json encoded proto.
//
// This function is meant as a convenience so the CLI can take either json or
// binary protos as input. This function guesses at whether to attempt to parse
// as binary or json first based on path's suffix.
func ReadBinaryOrJSONPb(ctx context.Context, path string, m proto.Message) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}

	if strings.HasSuffix(path, ".jsonpb") || strings.HasSuffix(path, ".jsonproto") {
		logging.Infof(ctx, "Attempting to parse %q as jsonpb first", path)

		err = unmarshalOpts.Unmarshal(b, m)
		if err == nil {
			return nil
		}

		logging.Warningf(ctx, "Parsing %q as jsonpb failed (%q), attempting to parse as binary pb", path, err)

		return proto.Unmarshal(b, m)
	}

	logging.Infof(ctx, "Attempting to parse %q as binary pb first", path)

	err = proto.Unmarshal(b, m)
	if err == nil {
		return nil
	}

	logging.Warningf(ctx, "Parsing %q as binarypb failed, attempting to parse as jsonpb", path)

	return unmarshalOpts.Unmarshal(b, m)
}

// ReadJsonl parses the newline-delimited jsonprotos in inPath. ctor must return
// a new empty proto to parse each line into.
func ReadJsonl[M proto.Message](inPath string, ctor func() M) (messages []M, err error) {
	f, err := os.Open(inPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
	}()

	messages = make([]M, 0)
	scanner := bufio.NewScanner(f)
	unmarshalOpts := protojson.UnmarshalOptions{DiscardUnknown: true}

	for scanner.Scan() {
		m := ctor()
		if err := unmarshalOpts.Unmarshal(scanner.Bytes(), m); err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}

	return messages, nil
}

// WriteJsonl writes a newline-delimited json file containing messages to outPath.
func WriteJsonl[M proto.Message](messages []M, outPath string) (err error) {
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() {
		err = outFile.Close()
	}()

	for _, m := range messages {
		jsonBytes, err := protojson.Marshal(m)
		if err != nil {
			return err
		}

		jsonBytes = append(jsonBytes, []byte("\n")...)

		if _, err = outFile.Write([]byte(jsonBytes)); err != nil {
			return err
		}
	}

	return nil
}

// FilepathAsJsonpb returns a copy of path, with the extension changed to
// ".jsonpb". If path is the empty string, an empty string is returned. Note
// that this function makes no attempt to check if the input path already has a
// jsonproto extension; i.e. if path is "a/b/test.jsonpb", the exact same path
// will be returned. Thus, it is up to the caller to check the returned path
// is different if this is required.
func FilepathAsJsonpb(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return path[0:len(path)-len(ext)] + ".jsonpb"
}
