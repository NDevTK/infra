// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"infra/unifiedfleet/app/util"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/google/go-cmp/cmp"
)

// testServerStream stores all messages sent in a slice.
// As a result, only strings should be sent to it.
type testServerStream struct {
	serverStream
	msgs []string
}

// SendMsg overrides behavior of sending a message to store it in a slice.
func (s *testServerStream) SendMsg(m interface{}) error {
	s.msgs = append(s.msgs, m.(string))
	return nil
}

// TestStreamNamespaceInterceptor tests the interceptor correctly sets the
// context. It does so by having the stream handler simply send the namespace
// set in the outgoing context to the server stream. From there, we can look at
// the messages received by that stream to ensure we have the right context.
func TestStreamNamespaceInterceptor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		incomingNamespace     string
		wantOutgoingNamespace string
		wantErr               bool
	}{
		{
			name:                  "no ns",
			incomingNamespace:     "",
			wantOutgoingNamespace: "os",
			wantErr:               false,
		},
		{
			name:                  "valid ns",
			incomingNamespace:     "os-partner",
			wantOutgoingNamespace: "os",
			wantErr:               false,
		},
		{
			name:                  "invalid ns",
			incomingNamespace:     "fake",
			wantOutgoingNamespace: "os",
			wantErr:               false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// This simply sends the *outgoing* context (what we forward to
			// UFS) to the stream.
			namespaceContextStreamHandler := func(srv interface{}, stream grpc.ServerStream) error {
				md, ok := metadata.FromOutgoingContext(stream.Context())
				if !ok {
					return errors.New("no metadata")
				}
				namespace, ok := md[util.Namespace]
				if !ok {
					return errors.New("no namespace in metadata")
				}

				return stream.SendMsg(namespace[0])
			}

			// Set up the context like it would be for a "real" incoming GRPC
			// request by setting the incoming metadata.
			incomingCtx := context.Background()
			if tt.incomingNamespace != "" {
				md := metadata.Pairs("namespace", tt.incomingNamespace)
				incomingCtx = metadata.NewIncomingContext(incomingCtx, md)
			}
			testStream := &testServerStream{serverStream: serverStream{ctx: incomingCtx}}

			err := streamNamespaceInterceptor(struct{}{}, testStream, &grpc.StreamServerInfo{}, namespaceContextStreamHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err: %t, want err: %t", (err != nil), tt.wantErr)
			}
			// Compare the "messages received" aka namespace we sent via
			// namespaceContextStreamHandler
			if diff := cmp.Diff(testStream.msgs, []string{tt.wantOutgoingNamespace}); diff != "" {
				t.Errorf("unexpected diff: %s", diff)
			}
		})
	}
}
