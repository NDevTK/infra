// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package templates

import (
	"fmt"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TemplateProcessor converts a container-specific template into a valid generic
// StartContainerRequest. Besides request conversions, a TemplateProcessor is
// also aware of a container's dependencies of other containers, whose addresses
// are determined at runtime. The addresses are provided as IpEndpoint
// placeholders in a template, and TemplateProcessor use placeholderPopulators
// to populate actual values.
type TemplateProcessor interface {
	Process(*api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error)
}

// RequestRouter is the entry point to template processing.
type RequestRouter struct{ TemplateProcessor }

func (*RequestRouter) Process(request *api.StartTemplatedContainerRequest) (*api.StartContainerRequest, error) {
	switch t := request.Template.Container.(type) {
	case *api.Template_CrosDut:
		actualProcessor := newCrosDutProcessor()
		return actualProcessor.Process(request)
	case *api.Template_CrosProvision:
		actualProcessor := newCrosProvisionProcessor()
		return actualProcessor.Process(request)
	case *api.Template_CrosTest:
		actualProcessor := newCrosTestProcessor()
		return actualProcessor.Process(request)
	default:
		return nil, status.Error(codes.Unimplemented, fmt.Sprintf("%v to be implemented", t))
	}
}
