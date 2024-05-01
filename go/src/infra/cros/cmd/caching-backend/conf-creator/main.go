// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This package creates the configuration files for nginx and keepalived used
// in the caching backend in Chrome OS fleet labs.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/grpc/prpc"

	models "infra/unifiedfleet/api/v1/models"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
)

var (
	keepalivedConfigFilePath = flag.String("keepalived-conf", "/mnt/conf/keepalived/keepalived.conf", "Path to where keepalived.conf should be created.")
	nginxConfigFilePath      = flag.String("nginx-conf", "/mnt/conf/nginx/nginx.conf", "Path to where nginx.conf should be created.")
	serviceAccountJSONPath   = flag.String("service-account", "/creds/service_accounts/artifacts-downloader-service-account.json", "Path to the service account JSON key file.")
	ufsService               = flag.String("ufs-host", "ufs.api.cr.dev", "Host of the UFS service.")
	cacheSizeInGB            = flag.Uint("nginx-cache-size", 750, "The size of nginx cache in GB.")
	nginxWorkerCount         = flag.Uint("nginx-worker-count", 0, "The nginx worker processes configuration (use '0' for 'auto'). (default 0)")
	gsaServerCount           = flag.Uint("gsa-server-count", 1, "The number of upstream gs_archive_server instances to be added in nginx-conf.")
	gsaInitialPort           = flag.Uint("gsa-initial-port", 18000, "The port number for the first instance of the gs_archive_server in nginx.conf. Port number will increase by 1 for all subsequent entries.")
	keepalivedInterface      = flag.String("keepalived-interface", "bond0", "The interface keepalived listens on.")
	otelTraceEndpoint        = flag.String("otel-trace-endpoint", "", "The OTel collector endpoint. e.g. http://127.0.0.1:4317")
)

var (
	nodeIP   = os.Getenv("NODE_IP")
	nodeName = os.Getenv("NODE_NAME")
)

func main() {
	if err := innerMain(); err != nil {
		log.Fatalf("Exiting due to an error: %s", err)
	}
	log.Printf("Exiting successfully")
}

func innerMain() error {
	flag.Parse()
	if nodeIP == "" {
		return fmt.Errorf("environment variable NODE_IP missing,")
	}
	if nodeName == "" {
		return fmt.Errorf("environment variable NODE_NAME missing,")
	}
	log.Println("Getting caching service information from UFS...")
	services, err := getCachingServices()
	if err != nil {
		return err
	}
	service, ok := findService(services, nodeIP, nodeName)
	if !ok {
		log.Println("Could not find caching service for this node in UFS")
		log.Println("Creating non-operational nginx.conf...")
		if err := ioutil.WriteFile(*nginxConfigFilePath, []byte(noOpNginxTemplate), 0644); err != nil {
			return err
		}
		log.Println("Creating non-operational keepalived.conf...")
		if err := ioutil.WriteFile(*keepalivedConfigFilePath, []byte(noOpKeepalivedTemplate), 0644); err != nil {
			return err
		}
		return nil
	}
	vip, err := nodeVirtualIP(service)
	if err != nil {
		return err
	}
	n := nginxConfData{
		WorkerCount: *nginxWorkerCount,
		// TODO(sanikak): Define types to make the unit clearer.
		// E.g.  type gigabyte int.
		CacheSizeInGB:     int(*cacheSizeInGB),
		GSAServerCount:    int(*gsaServerCount),
		GSAInitialPort:    int(*gsaInitialPort),
		VirtualIP:         vip,
		OtelTraceEndpoint: *otelTraceEndpoint,
	}
	k := keepalivedConfData{
		VirtualIP: vip,
		Interface: *keepalivedInterface,
	}
	switch {
	case nodeIP == service.GetPrimaryNode() || nodeName == service.GetPrimaryNode():
		peerIP, err := lookupHost(service.GetSecondaryNode())
		if err != nil {
			return err
		}
		n.UpstreamHost = net.JoinHostPort(peerIP, strconv.Itoa(int(service.GetPort())))
		k.UnicastPeer = peerIP
		// Keepalived configuration uses the following non-inclusive language.
		k.State = "MASTER"
		k.Priority = 150
	case nodeIP == service.GetSecondaryNode() || nodeName == service.GetSecondaryNode():
		peerIP, err := lookupHost(service.GetPrimaryNode())
		if err != nil {
			return err
		}
		k.UnicastPeer = peerIP
		k.State = "BACKUP"
		k.Priority = 100
	default:
		return fmt.Errorf("node is neither the primary nor the secondary")
	}
	if err := buildAndWriteConfig("nginx", nginxTemplate, n, *nginxConfigFilePath); err != nil {
		return err
	}
	if s := service.GetState(); s != models.State_STATE_SERVING {
		log.Printf("Didn't config keepalived since the service state in UFS isn't STATE_SERVING (%s instead)", s)
		return ioutil.WriteFile(*keepalivedConfigFilePath, []byte(noOpKeepalivedTemplate), 0644)
	}
	return buildAndWriteConfig("keepalived", keepalivedTemplate, k, *keepalivedConfigFilePath)
}

func buildAndWriteConfig(name string, templ string, data interface{}, path string) error {
	log.Printf("Configuring %q and writing to %q ...", name, path)
	d, err := buildConfig(templ, data)
	if err != nil {
		return fmt.Errorf("build and write config of %q: %s", name, err)
	}
	if err := ioutil.WriteFile(path, []byte(d), 0644); err != nil {
		return fmt.Errorf("build and write config of %q: %s", name, err)
	}
	return nil
}

// getCachingServiceFromUFS gets all caching services listed in the UFS.
func getCachingServices() ([]*models.CachingService, error) {
	ctx := context.Background()
	md := metadata.Pairs("namespace", "os")
	ctx = metadata.NewOutgoingContext(ctx, md)
	o := auth.Options{
		Method:                 auth.ServiceAccountMethod,
		ServiceAccountJSONPath: *serviceAccountJSONPath,
	}
	a := auth.NewAuthenticator(ctx, auth.SilentLogin, o)
	hc, err := a.Client()
	if err != nil {
		return nil, fmt.Errorf("could not establish HTTP client: %s", err)
	}
	ic := ufsapi.NewFleetPRPCClient(&prpc.Client{
		C:    hc,
		Host: *ufsService,
		Options: &prpc.Options{
			UserAgent: "caching-backend/3.0.0", // Any UserAgent with version below 3.0.0 is unsupported by UFS.
		},
	})
	res, err := ic.ListCachingServices(ctx, &ufsapi.ListCachingServicesRequest{})
	if err != nil {
		return nil, fmt.Errorf("query to UFS failed: %s", err)
	}
	return res.GetCachingServices(), nil
}
