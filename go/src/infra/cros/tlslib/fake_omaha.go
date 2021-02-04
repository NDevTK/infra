// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tlslib

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"go.chromium.org/chromiumos/config/go/api/test/tls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/cros/tlslib/internal/nebraska"
)

// CreateFakeOmaha implements TLS CreateFakeOmaha API.
func (s *Server) CreateFakeOmaha(ctx context.Context, req *tls.CreateFakeOmahaRequest) (*tls.FakeOmaha, error) {
	fo := req.GetFakeOmaha()
	gsPathPrefix := fo.GetTargetBuild().GetGsPathPrefix()
	if len(gsPathPrefix) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "CreateFakeOmaha: empty GS path in the target bulid")
	}
	payloads := fo.GetPayloads()
	for _, p := range payloads {
		if p.GetType() == tls.FakeOmaha_Payload_TYPE_UNSPECIFIED {
			return nil, status.Errorf(codes.InvalidArgument, "CreateFakeOmaha: payload %q has unspecified type", p.GetId())
		}
	}

	dutName := fo.GetDut()
	if len(dutName) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "CreateFakeOmaha: empty DUT name")
	}
	payloadsURL, err := s.getCacheURL(ctx, gsPathPrefix, dutName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFakeOmaha: failed to get cache payload URL: %s", err)
	}
	n := nebraska.NewServer(nebraska.NewEnvironment())
	if err := n.Start(gsPathPrefix, payloads, payloadsURL); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFakeOmaha: failed to start fake Omaha: %s", err)
	}
	u, err := s.exposePort(ctx, dutName, n.Port, fo.GetExposedViaProxy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFakeOmaha: failed to expose fake Omaha: %s", err)
	}

	name := fmt.Sprintf("fakeOmaha/%s", uuid.New().String())
	if err := s.resMgr.CreateResource(name, n); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFakeOmaha: failed to create resource: %s", err)
	}

	fo.Name = name
	q := url.Values{}
	if fo.GetCriticalUpdate() {
		q.Set("critical_update", "True")
	}
	// TODO(guocb) handle the case of 'return_noupdate_starting' > 1.
	if fo.GetReturnNoupdateStarting() == 1 {
		q.Set("no_update", "True")
	}
	exposedURL := url.URL{Scheme: "http", Host: u, Path: "/update", RawQuery: q.Encode()}
	fo.OmahaUrl = exposedURL.String()
	log.Printf("CreateFakeOmaha: %q update URL: %s", fo.Name, fo.OmahaUrl)
	return fo, nil
}

// DeleteFakeOmaha implements TLS DeleteFakeOmaha API.
func (s *Server) DeleteFakeOmaha(ctx context.Context, req *tls.DeleteFakeOmahaRequest) (*empty.Empty, error) {
	r, err := s.resMgr.DeleteResource(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteFakeOmaha: delete resource: %s", err)
	}
	err = r.Close()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteFakeOmaha: close fake Omaha: %s", err)
	}
	return nil, nil
}

func (s *Server) exposePort(ctx context.Context, dutName string, localPort int, requireProxy bool) (string, error) {
	c := tls.NewWiringClient(s.wiringConn)

	rsp, err := c.ExposePortToDut(ctx, &tls.ExposePortToDutRequest{
		DutName:            dutName,
		LocalPort:          int32(localPort),
		RequireRemoteProxy: requireProxy,
	})
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(rsp.GetExposedAddress(), fmt.Sprintf("%d", rsp.GetExposedPort())), nil
}
