// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"context"
	"crypto"
	"fmt"
	"hash"
	"infra/libs/cipkg_new/core"
	"io"
	"net/http"

	"github.com/spf13/afero"
)

// ActionURLFetchTransformer is the default transformer for url action.
func ActionURLFetchTransformer(a *core.ActionURLFetch, deps []PackageDependency) (*core.Derivation, error) {
	return ReexecDerivation(a, deps, false)
}

// ActionURLFetchExecutor is the default executor for url action.
func ActionURLFetchExecutor(ctx context.Context, a *core.ActionURLFetch, dstFS afero.Fs) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.Url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := dstFS.Create("file")
	if err != nil {
		return err
	}
	defer f.Close()

	h, err := hashFromEnum(a.HashAlgorithm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, io.TeeReader(resp.Body, h)); err != nil {
		return err
	}

	if actual := fmt.Sprintf("%x", h.Sum(nil)); actual != a.HashValue {
		return fmt.Errorf("hash mismatch: expected: %s, actual: %s", a.HashValue, actual)
	}
	return nil
}

func hashFromEnum(h core.HashAlgorithm) (hash.Hash, error) {
	if h == core.HashAlgorithm_HASH_UNSPECIFIED {
		return &emptyHash{}, nil
	}

	switch h {
	case core.HashAlgorithm_HASH_MD5:
		return crypto.MD5.New(), nil
	case core.HashAlgorithm_HASH_SHA256:
		return crypto.SHA256.New(), nil
	}

	return nil, fmt.Errorf("unknown hash algorithm: %s", h)
}

// If no hash algorithm is provided, we will use an empty hash to skip the
// check and assume the artifact from the url is always same. This is only
// used for legacy content.
type emptyHash struct{}

func (h *emptyHash) Write(p []byte) (int, error) { return len(p), nil }
func (h *emptyHash) Sum(b []byte) []byte         { return []byte{} }
func (h *emptyHash) Reset()                      {}
func (h *emptyHash) Size() int                   { return 0 }
func (h *emptyHash) BlockSize() int              { return 0 }
