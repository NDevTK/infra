package gs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	gcgs "go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"
)

// DirWriter Mockable interface for writing whole directories at once
type DirWriter interface {
	WriteDir(ctx context.Context) error
}

type prodDirWriter struct {
	// The directory to be written from
	localRootDir string
	// The directory to be written to
	gsRootDir gcgs.Path

	// Mockable means of carrying out file-level writes
	client AuthedClient
}

var _ DirWriter = &prodDirWriter{}

// AuthedClient Mockable wrapper around the core "spin up subWriter" flow
type AuthedClient interface {
	NewWriter(p Path) (io.Writer, error)
}

type realAuthedClient struct {
	client gcgs.Client
}

var _ AuthedClient = &realAuthedClient{}

func (c *realAuthedClient) NewWriter(p Path) (io.Writer, error) {
	return c.client.NewWriter(gcgs.Path(p))
}

// NewDirWriter creates an object which can write a directory and its subdirectories to the given Google Storage path
func NewDirWriter(localPath string, gsPath Path, client AuthedClient) (DirWriter, error) {
	if err := verifyPaths(localPath, string(gsPath)); err != nil {
		return nil, err
	}
	return &prodDirWriter{
		localRootDir: localPath,
		gsRootDir:    gcgs.Path(gsPath),
		client:       client,
	}, nil
}

func verifyPaths(localPath string, gsPath string) error {
	problems := []string{}
	if _, err := os.Stat(localPath); err != nil {
		problems = append(problems, fmt.Sprintf("invalid local path (%s)", localPath))
	} else if _, err := os.Open(localPath); err != nil {
		problems = append(problems, fmt.Sprintf("unreadable local path (%s)", localPath))
	}
	if _, err := url.Parse(gsPath); err != nil {
		problems = append(problems, fmt.Sprintf("invalid GS path (%s)", gsPath))
	}
	if len(problems) > 0 {
		return errors.Reason("path errors: %s", strings.Join(problems, ", ")).Err()
	}
	return nil
}

// Path Google Storage path, to file or directory
type Path gcgs.Path

// WriteDir Write all files in the subtree of the local path to the corresponding places
//  in the subtree of the GS path
func (w *prodDirWriter) WriteDir(ctx context.Context) error {
	logging.Debugf(ctx, "writing %s", w.localRootDir)
	writeOne := func(src string, info os.FileInfo, err error) error {
		if info.IsDir() {
			logging.Debugf(ctx, "file %s is a directory", src)
			return nil
		}
		relPath, err := filepath.Rel(w.localRootDir, src)
		if err != nil {
			return errors.Annotate(err, "writing from %s to %s/<?>", src, w.gsRootDir).Err()
		}
		gsDest := w.gsRootDir.Concat(relPath)
		dest := Path(gsDest)
		logging.Debugf(ctx, "writing from %s to %s", src, dest)
		f, err := os.Open(src)
		if err != nil {
			return errors.Annotate(err, "writing from %s to %s", src, dest).Err()
		}
		if info.Size() <= 0 {
			logging.Warningf(ctx, "file %s empty", src)
			return nil
		}
		writer, err := w.client.NewWriter(dest)
		if err != nil {
			return errors.Annotate(err, "writing from %s to %s", src, dest).Err()
		}
		bs := make([]byte, info.Size())
		if _, err = f.Read(bs); err != nil {
			return errors.Annotate(err, "writing from %s to %s", src, dest).Err()
		}
		if n, err := writer.Write(bs); err != nil || n <= 0 {
			if err != nil {
				return errors.Annotate(err, "writing from %s to %s", src, dest).Err()
			}
			return errors.Reason("nothing written to %s", dest).Err()
		} else {
			logging.Debugf(ctx, "wrote %d bytes to %s", n, dest)
		}
		cli := (w.client).(*realAuthedClient).client
		logging.Debugf(ctx, "now checking contents of  %s", dest)
		if a, err := cli.Attrs(gsDest); err != nil {
			return errors.Annotate(err, "checking contents of file %s", gsDest).Err()
		} else {
			logging.Debugf(ctx, "%s successfully written to\nsize: %d, bucket: %s, name: %s, owner: %s\n",
				dest, a.Size, a.Bucket, a.Name, a.Owner,
			)
		}
		logging.Debugf(ctx, "checked contents of  %s", dest)
		return nil
	}
	if err := filepath.Walk(w.localRootDir, writeOne); err != nil {
		return errors.Annotate(err, "writing dir %s to %s", w.localRootDir, w.gsRootDir).Err()
	}
	cli := (w.client).(*realAuthedClient).client
	if _, err := cli.Attrs(w.gsRootDir); err != nil {
		return errors.Annotate(err, "checking contents of root dir %s", w.gsRootDir).Err()
	}
	return nil
}

func newAuthenticatedTransport(ctx context.Context, f *authcli.Flags) (http.RoundTripper, error) {
	o, err := f.Options()
	if err != nil {
		return nil, errors.Annotate(err, "creating authenticated transport").Err()
	}
	a := auth.NewAuthenticator(ctx, auth.SilentLogin, o)
	rt, err := a.Transport()
	if err != nil {
		return nil, errors.Annotate(err, "creating authenticated transport").Err()
	}
	return rt, nil
}

// NewAuthedClient Create a client with the given auth flags
func NewAuthedClient(ctx context.Context, f *authcli.Flags) (AuthedClient, error) {
	t, err := newAuthenticatedTransport(ctx, f)
	if err != nil {
		return nil, errors.Annotate(err, "creating authenticated GS client").Err()
	}
	cli, err := gcgs.NewProdClient(ctx, t)
	if err != nil {
		return nil, errors.Annotate(err, "creating authenticated GS client").Err()
	}
	return &realAuthedClient{client: cli}, nil
}
