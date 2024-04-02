// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package secretmanager creates an interface for working with the GCP
// SecretManager API.
package secretmanager

import (
	"context"
	"errors"
	"fmt"

	cloudsm "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// GetSecret requests and returns a secret from GCP SecretManager.
//
// NOTE: secretName is in the format "projects/<projectNumber>/secrets/<secretName>/versions/<versionNumber>"
func GetSecret(ctx context.Context, secretName string, projectID, versionNumber int) (secret string, err error) {
	client, err := cloudsm.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		// Close the channel and catch the error if one arises during the call.
		err = errors.Join(err, client.Close())
	}()

	name := fmt.Sprintf("projects/%d/secrets/%s/versions/%d", projectID, secretName, versionNumber)

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", err
	}
	return string(result.Payload.Data), nil
}
