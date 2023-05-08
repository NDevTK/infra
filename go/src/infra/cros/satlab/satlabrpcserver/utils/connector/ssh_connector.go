// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package connector

import (
	"context"

	"golang.org/x/crypto/ssh"
)

type ISSHClientConnector interface {
	Connect(ctx context.Context, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
}
