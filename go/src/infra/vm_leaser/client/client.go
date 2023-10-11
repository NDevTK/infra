// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"infra/vm_leaser/internal/site"
)

// Config stores options needed for the VM Leaser service.
type Config struct {
	// Endpoint of the VM Leaser service.
	vmLeaserServiceEndpoint string
	// Port of the VM Leaser service.
	vmLeaserServicePort int
	// Transport credentials
	creds credentials.TransportCredentials
}

// Client is a VM Leaser client.
type Client struct {
	conn           *grpc.ClientConn
	VMLeaserClient api.VMLeaserServiceClient
	Email          string
}

// Close closes the client.
func (c *Client) Close() {
	if c != nil && c.conn != nil {
		c.conn.Close()
	}
}

// LocalConfig returns the local configuration for the VM Leaser client.
func LocalConfig() *Config {
	return &Config{
		vmLeaserServiceEndpoint: site.LocalVMLeaserServiceEndpoint,
		vmLeaserServicePort:     site.LocalVMLeaserServicePort,
		creds:                   insecure.NewCredentials(),
	}
}

// StagingConfig returns the staging configuration for the VM Leaser client.
//
// The staging instance of the VM Leaser service is a GCP cloud project.
func StagingConfig() *Config {
	return &Config{
		vmLeaserServiceEndpoint: site.StagingVMLeaserServiceEndpoint,
		vmLeaserServicePort:     site.StagingVMLeaserServicePort,
		creds:                   credentials.NewTLS(nil),
	}
}

// ProdConfig returns the production configuration for the VM Leaser client.
//
// The prod instance of the VM Leaser service is a GCP cloud project.
func ProdConfig() *Config {
	return &Config{
		vmLeaserServiceEndpoint: site.ProdVMLeaserServiceEndpoint,
		vmLeaserServicePort:     site.ProdVMLeaserServicePort,
		creds:                   credentials.NewTLS(nil),
	}
}

// NewClient creates a new client for the VM Leaser service.
func NewClient(ctx context.Context, c *Config) (*Client, error) {
	if c == nil {
		return nil, errors.New("vm leaser client: cannot create new client from empty base config")
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(c.creds),
		grpc.WithBlock(),
	}

	auth := auth.NewAuthenticator(ctx, auth.SilentLogin, chromeinfra.SetDefaultAuthOptions(auth.Options{
		UseIDTokens: true,
		Audience:    "https://" + c.vmLeaserServiceEndpoint,
	}))
	creds, err := auth.PerRPCCredentials()
	if err != nil {
		return nil, err
	}
	dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(creds))

	// Fail fast if dial is not available
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	target := fmt.Sprintf("%s:%d", c.vmLeaserServiceEndpoint, c.vmLeaserServicePort)
	conn, err := grpc.DialContext(dialCtx, target, dialOpts...)
	if err != nil {
		return nil, err
	}
	email, _ := auth.GetEmail()
	return &Client{
		conn:           conn,
		VMLeaserClient: api.NewVMLeaserServiceClient(conn),
		Email:          email,
	}, nil
}
