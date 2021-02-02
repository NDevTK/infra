// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package tlslib

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"testing"

	"go.chromium.org/chromiumos/config/go/api/test/tls"
	"google.golang.org/grpc"
)

// Flags needed for integration tests which depend on real DUTs and networking.
var (
	wiringPort = flag.Int("wiring-port", 0, "run integration test with a TLW server listening on this port")
	dutName    = flag.String("dut", "", "the dut name for the integration test")
)

func TestFakeOmahaIntegration(t *testing.T) {
	if *wiringPort == 0 {
		t.Skip("skip integration test due to no TLW service")
	}
	if *dutName == "" {
		t.Skip("skip integration test due to no DUT specified")
	}

	const fakeAURequest = `<?xml version="1.0" encoding="UTF-8"?>
<request requestid="1bcea19b-8ecf-4599-b37a-47018b7b8ecb" sessionid="710055d0-f9ec-4efd-aa8b-3ab153e4e0e9" protocol="3.0" updater="ChromeOSUpdateEngine" updaterversion="0.1.0.0" installsource="ondemandupdate" ismachine="1">
    <os version="Indy" platform="Chrome OS" sp="13336.0.0_x86_64"></os>
    <app appid="{3A837630-D749-4B7A-86C1-DB0ECC07A08B}" version="13336.0.0" track="stable-channel" board="banjo" hardware_class="BANJO C7A-C6I-A4O" delta_okay="true" installdate="4935" lang="en-US" fw_version="" ec_version="" >
        <updatecheck></updatecheck>
    </app>
</request>
`
	connTlw, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", *wiringPort), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("connect to TLW service on port %d", *wiringPort)
	}
	defer connTlw.Close()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("NewServer: %s", err)
	}
	s, err := NewServer(context.Background(), connTlw)
	if err != nil {
		t.Fatalf("NewServer: %s", err)
	}
	go s.Serve(l)
	defer s.GracefulStop()

	conn, err := grpc.Dial(fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("connect to TLS server: %s", err)
	}
	defer conn.Close()
	c := tls.NewCommonClient(conn)

	rsp, err := c.CreateFakeOmaha(context.Background(), &tls.CreateFakeOmahaRequest{
		FakeOmaha: &tls.FakeOmaha{
			Dut: *dutName,
			TargetBuild: &tls.ChromeOsImage{
				PathOneof: &tls.ChromeOsImage_GsPathPrefix{
					// It doesn't matter whether the board in below URL match
					// with the DUT board.
					GsPathPrefix: "gs://chromeos-image-archive/banjo-release/R90-13809.0.0",
				},
			},
			Payloads: []*tls.FakeOmaha_Payload{{Type: tls.FakeOmaha_Payload_FULL}},
		},
	})
	if err != nil {
		t.Errorf("CreateFakeOmaha() error: %s", err)
	}
	prefix := "fakeOmaha/"
	if !strings.HasPrefix(rsp.Name, prefix) {
		t.Errorf("CreateFakeOmaha() error: resource name %q not start with %q", rsp.Name, prefix)
	}
	log.Printf("The Omaha URL is %q", rsp.OmahaUrl)
	stream, err := c.ExecDutCommand(context.Background(), &tls.ExecDutCommandRequest{
		Name:    *dutName,
		Command: "curl",
		Args:    []string{"-X", "POST", "-d", "@-", "-H", "content-type:application/xml", rsp.OmahaUrl},
		Stdin:   []byte(fakeAURequest),
	})
	if err != nil {
		t.Fatalf("exec dut command error: %s", err)
	}
	var stdout bytes.Buffer
readStream:
	for {
		rsp, err := stream.Recv()
		switch err {
		case nil:
			stdout.Write(rsp.Stdout)
		case io.EOF:
			break readStream
		default:
			t.Fatalf("ExecDutCommand RPC error")
		}
	}
	output := stdout.String()
	// We think the test is good as long as receiving a valid response.
	if !strings.HasPrefix(output, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("fake Omaha didn't respond a valid xml payload, got %q", output)
	}

	log.Println("Delete the fake Omaha created")
	_, err = s.DeleteFakeOmaha(context.Background(), &tls.DeleteFakeOmahaRequest{Name: rsp.GetName()})
	if err != nil {
		t.Errorf("DeleteFakeOmaha(%q) error: %q", rsp.GetName(), err)
	}
}

func TestCreateFakeOmahaErrors(t *testing.T) {
	tests := []struct {
		name string
		req  *tls.CreateFakeOmahaRequest
	}{
		{"nil", nil},
		{
			"just dut name",
			&tls.CreateFakeOmahaRequest{
				FakeOmaha: &tls.FakeOmaha{Dut: "dutname"},
			},
		},
		{
			"no payload type",
			&tls.CreateFakeOmahaRequest{
				FakeOmaha: &tls.FakeOmaha{
					Dut: "dutname",
					TargetBuild: &tls.ChromeOsImage{
						PathOneof: &tls.ChromeOsImage_GsPathPrefix{
							GsPathPrefix: "gs://chromeos-image-archive/eve-release/R90-13809.0.0",
						},
					},
				},
			},
		},
	}
	s, err := NewServer(context.Background(), nil)
	if err != nil {
		t.Fatalf("New TLS server: %s", err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.CreateFakeOmaha(context.Background(), tc.req)
			if err == nil {
				t.Errorf("CreateFakeOmaha(%q) succeeded with empty input, want error", tc.req)
			}
		})
	}
}
