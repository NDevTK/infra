package builtins

import (
	"bytes"
	"context"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"infra/libs/cipkg"
)

// Fetch url(s) in a single derivation. Usually FetchURLsBuilder shouldn't be
// used directly. Use FetchURLs for downloading multiple URLs and FetchURL for
// single.
const FetchURLsBuilder = BuiltinBuilderPrefix + "fetchURLs"

type FetchURLs struct {
	Name string
	URLs []FetchURL
}

func (f *FetchURLs) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	arg, err := json.Marshal(f)
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("encode json failed: %v: %w", arg, err)
	}
	return cipkg.Derivation{
		Name:    f.Name,
		Builder: FetchURLsBuilder,
		Args:    []string{string(arg)},
	}, cipkg.PackageMetadata{}, nil
}

type FetchURL struct {
	Name          string
	URL           string
	Filename      string
	Executable    bool
	HashAlgorithm crypto.Hash
	HashString    string
}

func (f *FetchURL) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	return (&FetchURLs{
		Name: f.Name,
		URLs: []FetchURL{*f},
	}).Generate(ctx)
}

func fetchURLs(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:fetchURLs", FetchURLs{...}]
	if len(cmd.Args) != 2 {
		return fmt.Errorf("invalid arguments: %v", cmd.Args)
	}
	out := GetEnv("out", cmd.Env)

	var arg FetchURLs
	if err := json.Unmarshal([]byte(cmd.Args[1]), &arg); err != nil {
		return fmt.Errorf("parse argument failed: %s: %w", cmd.Args, err)
	}

	for _, a := range arg.URLs {
		if err := fetchURL(ctx, out, a); err != nil {
			return fmt.Errorf("fetch url failed: %v: %w", arg, err)
		}
	}

	return nil
}

func fetchURL(ctx context.Context, out string, arg FetchURL) error {
	var h hash.Hash
	switch {
	case arg.HashAlgorithm == HashIgnore:
		h = &emptyHash{}
	case arg.HashAlgorithm.Available():
		h = arg.HashAlgorithm.New()
	default:
		return fmt.Errorf("unavailable hash algorithm: %s", arg.HashAlgorithm)
	}
	hv, err := hex.DecodeString(arg.HashString)
	if err != nil {
		return fmt.Errorf("decode hash string failed: %s: %w", arg.HashString, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, arg.URL, nil)
	if err != nil {
		return fmt.Errorf("new request failed: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("get url failed: %w", err)
	}
	defer resp.Body.Close()

	if arg.Filename == "" {
		arg.Filename = path.Base(req.URL.Path)
	}
	filename := filepath.Join(out, arg.Filename)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file failed: %s: %w", filename, err)
	}
	defer f.Close()

	r := io.TeeReader(resp.Body, h)
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("download file failed: %s: %w", filename, err)
	}
	hresp := h.Sum(nil)
	if !bytes.Equal(hv, hresp) {
		return fmt.Errorf("hash value mismatch: expected: %x, result: %x", hv, hresp)
	}

	if arg.Executable {
		finfo, err := f.Stat()
		if err != nil {
			return fmt.Errorf("get file info failed: %s: %w", filename, err)
		}
		if err := os.Chmod(filename, finfo.Mode()|0111); err != nil {
			return fmt.Errorf("chmod file failed: %s: %w", filename, err)
		}
	}
	return nil
}

const HashIgnore crypto.Hash = 255

// If no hash algorithm is provided, we will use an empty hash to skip the
// check and assume the artifact from the url is always same. This is only
// used for legacy content.
type emptyHash struct{}

func (h *emptyHash) Write(p []byte) (int, error) { return len(p), nil }
func (h *emptyHash) Sum(b []byte) []byte         { return []byte{} }
func (h *emptyHash) Reset()                      {}
func (h *emptyHash) Size() int                   { return 0 }
func (h *emptyHash) BlockSize() int              { return 0 }
