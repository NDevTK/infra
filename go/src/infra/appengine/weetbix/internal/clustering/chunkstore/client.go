// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chunkstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"regexp"

	"cloud.google.com/go/storage"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/grpcmon"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"infra/appengine/weetbix/internal/clustering"
	cpb "infra/appengine/weetbix/internal/clustering/proto"
)

// objectRe matches validly formed object IDs.
var objectRe = regexp.MustCompile(`^[0-9a-f]{32}$`)

// Client provides methods to store and retrieve chunks of test failures.
type Client struct {
	// client is the GCS client used to access chunks.
	client *storage.Client
	// bucket is the GCS bucket in which chunks are stored.
	bucket string
}

// NewClient initialises a new chunk storage client, that uses the specified
// GCS bucket as the backing store.
func NewClient(ctx context.Context, bucket string) (*Client, error) {
	// Credentials with Cloud scope.
	creds, err := auth.GetPerRPCCredentials(ctx, auth.AsSelf, auth.WithScopes(auth.CloudOAuthScopes...))
	if err != nil {
		return nil, errors.Annotate(err, "failed to get PerRPCCredentials").Err()
	}

	// Initialize the client.
	options := []option.ClientOption{
		option.WithGRPCDialOption(grpc.WithPerRPCCredentials(creds)),
		option.WithGRPCDialOption(grpcmon.WithClientRPCStatsMonitor()),
		option.WithScopes(storage.ScopeReadWrite),
	}
	cl, err := storage.NewClient(ctx, options...)

	if err != nil {
		return nil, errors.Annotate(err, "failed to instantiate Cloud Storage client").Err()
	}
	return &Client{
		client: cl,
		bucket: bucket,
	}, nil
}

// Close releases resources associated with the client.
func (c *Client) Close() {
	c.client.Close()
}

// Put saves the given chunk to storage. If successful, it returns
// the randomly-assigned ID of the created object.
func (c *Client) Put(ctx context.Context, project string, content *cpb.Chunk) (string, error) {
	if err := validateProject(project); err != nil {
		return "", err
	}
	b, err := proto.Marshal(content)
	if err != nil {
		return "", errors.Annotate(err, "marhsalling chunk").Err()
	}
	objID, err := generateObjectID()
	if err != nil {
		return "", err
	}

	name := fileName(project, objID)
	obj := c.client.Bucket(c.bucket).Object(name)
	w := obj.NewWriter(ctx)

	// As the file is small (<8MB), set ChunkSize to object size to avoid
	// excessive memory usage.
	w.ChunkSize = len(b)
	w.ContentType = "application/x-protobuf"
	_, err = w.Write(b)
	if err != nil {
		return "", errors.Annotate(err, "writing object %q", name).Err()
	}
	err = w.Close()
	if err != nil {
		return "", errors.Annotate(err, "closing object %q", name).Err()
	}
	return objID, nil
}

// Get retrieves the chunk with the specified object ID and returns it.
func (c *Client) Get(ctx context.Context, project, objectID string) (*cpb.Chunk, error) {
	if err := validateProject(project); err != nil {
		return nil, err
	}
	if err := validateObjectID(objectID); err != nil {
		return nil, err
	}
	name := fileName(project, objectID)
	obj := c.client.Bucket(c.bucket).Object(name)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "creating reader %q", name).Err()
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Annotate(err, "read object %q", name).Err()
	}
	content := &cpb.Chunk{}
	if err := proto.Unmarshal(b, content); err != nil {
		return nil, errors.Annotate(err, "unmarshal chunk").Err()
	}
	return content, nil
}

func validateProject(project string) error {
	if !clustering.ProjectRe.MatchString(project) {
		return fmt.Errorf("project %q is not a valid", project)
	}
	return nil
}

func validateObjectID(id string) error {
	if objectRe.MatchString(id) {
		return fmt.Errorf("object ID %q is not a valid", id)
	}
	return nil
}

// generateObjectID returns a random 128-bit object ID, encoded as
// 32 lowercase hexadecimal characters.
func generateObjectID() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func fileName(project, objectID string) string {
	return fmt.Sprintf("/projects/%s/chunks/%s.binarypb", project, objectID)
}
