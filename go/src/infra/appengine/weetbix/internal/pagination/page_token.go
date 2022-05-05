// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pagination

import (
	"encoding/base64"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/grpc/appstatus"

	paginationpb "infra/appengine/weetbix/internal/pagination/proto"
)

// ParseToken extracts a string slice position from the given page token.
// May return an appstatus-annotated error.
func ParseToken(token string) ([]string, error) {
	if token == "" {
		return nil, nil
	}

	tokBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, InvalidToken(err)
	}

	msg := &paginationpb.PageToken{}
	if err := proto.Unmarshal(tokBytes, msg); err != nil {
		return nil, InvalidToken(err)
	}
	return msg.Position, nil
}

// Token converts an string slice representing page token position to an opaque
// token string.
func Token(pos ...string) string {
	if pos == nil {
		return ""
	}

	msgBytes, err := proto.Marshal(&paginationpb.PageToken{Position: pos})
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(msgBytes)
}

// InvalidToken annotates the error with InvalidArgument appstatus.
func InvalidToken(err error) error {
	return appstatus.Attachf(err, codes.InvalidArgument, "invalid page_token")
}
