// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
)

// getSecureClient gets a secure http.Client pointed at a specific host.
//
// TODO(gregorynisbet): Remove this function as well as the dependency on authcli.Flags.
//
//	We should be able to manually construct an auth.Options object with the settings that we want.
//	However, I know for a fact that using authFlags to produce an authFlags.Options() object
//	produces a usable client. Sometime in the future, I will opportunistically replace this
//	function with something more reasonable.
func getSecureClient(ctx context.Context, host string, authFlags authcli.Flags) (*http.Client, error) {
	authOptions, err := authFlags.Options()
	if err != nil {
		return nil, errors.Annotate(err, "creating secure client").Err()
	}
	authOptions.UseIDTokens = true
	if authOptions.Audience == "" {
		authOptions.Audience = "https://" + host
	}
	authenticator := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOptions)
	httpClient, err := authenticator.Client()
	if err != nil {
		return nil, errors.Annotate(err, "creating secure client").Err()
	}
	return httpClient, nil
}

// Message showProto writes a proto message as an indentend object.
func showProto(dst io.Writer, message proto.Message) (int, error) {
	if dst == nil {
		return 0, errors.New("dest cannot be nil")
	}
	bytes, err := (&protojson.MarshalOptions{
		Indent: "  ",
	}).Marshal(message)
	if err != nil {
		return 0, errors.Annotate(err, "show proto").Err()
	}
	return dst.Write(bytes)
}
