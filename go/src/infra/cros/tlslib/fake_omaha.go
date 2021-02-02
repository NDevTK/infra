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
	"strconv"
	"strings"

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
		return nil, status.Errorf(codes.InvalidArgument, "empty GS path in the target build")
	}
	payloads := fo.GetPayloads()
	if len(payloads) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "no payloads specified")
	}
	for _, p := range payloads {
		if p.GetType() == tls.FakeOmaha_Payload_TYPE_UNSPECIFIED {
			return nil, status.Errorf(codes.InvalidArgument, "payload %q has unspecified type", p.GetId())
		}
	}

	dutName := fo.GetDut()
	if len(dutName) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "empty DUT name")
	}
	payloadsServerURL, err := s.cacheForDut(ctx, gsPathPrefix, dutName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cache payload URL: %s", err)
	}
	n, err := nebraska.NewServer(ctx, nebraska.NewEnvironment(), gsPathPrefix, payloads, payloadsServerURL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start fake Omaha: %s", err)
	}
	u, err := s.exposePort(ctx, dutName, n.Port(), fo.GetExposedViaProxy())
	if err != nil {
		msg := []string{fmt.Sprintf("expose fake Omaha: %s", err)}
		if err := n.Close(); err != nil {
			msg = append(msg, fmt.Sprintf("close Nebraska: %s", err))
		}
		return nil, status.Errorf(codes.Internal, strings.Join(msg, ", "))
	}

	name := fmt.Sprintf("fakeOmaha/%s", uuid.New().String())
	if err := s.resMgr.Add(name, n); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create resource: %s", err)
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
	r, err := s.resMgr.Remove(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete resource: %s", err)
	}
	if err = r.Close(); err != nil {
		return nil, status.Errorf(codes.Internal, "close fake Omaha: %s", err)
	}
	return nil, nil
}

func (s *Server) exposePort(ctx context.Context, dutName string, localPort int, requireProxy bool) (string, error) {
	c := s.wiringClient()

	rsp, err := c.ExposePortToDut(ctx, &tls.ExposePortToDutRequest{
		DutName:            dutName,
		LocalPort:          int32(localPort),
		RequireRemoteProxy: requireProxy,
	})
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(rsp.GetExposedAddress(), strconv.Itoa(int(rsp.GetExposedPort()))), nil
}
