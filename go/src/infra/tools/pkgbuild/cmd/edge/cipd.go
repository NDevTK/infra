// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/filesystem"
)

type cipdPackage actions.Package

var (
	errPackgeNotExist     = errors.New("no such derivation tag")
	errAmbiguousPackgeTag = errors.New("ambiguity when resolving the derivation tag")
)

func (pkg cipdPackage) check(ctx context.Context, cipdService string) error {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return nil
	}

	step, ctx := startStep(ctx, fmt.Sprintf("describing cipd package %s:%s", cipd.Name, pkg.derivationTag()))
	return step.With(func() error {
		cmd := step.cipdCommand("describe", cipd.Name,
			"-service-url", cipdService,
			"-version", pkg.derivationTag(),
		)
		var b bytes.Buffer
		cmd.Stdout = io.MultiWriter(&b, cmd.Stdout)
		cmd.Stderr = io.MultiWriter(&b, cmd.Stderr)

		logging.Infof(ctx, "command: %+v", cmd)
		if err := cmd.Run(); err != nil {
			out := b.String()
			if strings.Contains(out, "no such tag") || strings.Contains(out, "no such package") {
				return errPackgeNotExist
			} else if strings.Contains(out, "ambiguity when resolving the tag") {
				return errAmbiguousPackgeTag
			}

			logging.Infof(ctx, "%s", out)
			return err
		}

		return nil
	})
}

func (pkg cipdPackage) upload(ctx context.Context, workdir, cipdService string, tags []string) (name string, iid string, err error) {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return "", "", nil
	}

	step, ctx := startStep(ctx, fmt.Sprintf("creating cipd package %s:%s with %s", cipd.Name, pkg.derivationTag(), cipd.Version))
	if err := step.With(func() error {
		// If .cipd file already exists, assume it has been uploaded.
		out := filepath.Join(workdir, pkg.DerivationID+".cipd")
		if _, err := os.Stat(out); err == nil || !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		iid, err = buildCIPD(ctx, step, name, pkg.Handler.OutputDirectory(), out)
		if err != nil {
			_ = filesystem.RemoveAll(out)
			return errors.Annotate(err, "failed to build cipd package").Err()
		}

		if err := registerCIPD(ctx, step, cipdService, out, append([]string{pkg.derivationTag()}, tags...)); err != nil {
			return errors.Annotate(err, "failed to register cipd package").Err()
		}

		return nil
	}); err != nil {
		return "", "", nil
	}

	return
}

func (pkg cipdPackage) download(ctx context.Context, cipdService string) error {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return nil
	}

	step, ctx := startStep(ctx, fmt.Sprintf("dowloading cipd package %s:%s", cipd.Name, pkg.derivationTag()))
	return step.With(func() error {
		if err := pkg.Handler.Build(func() error {
			cmd := step.cipdCommand("export",
				"-service-url", cipdService,
				"-root", pkg.Handler.OutputDirectory(),
				"-ensure-file", "-",
			)
			cmd.Stdin = strings.NewReader(fmt.Sprintf("%s %s", cipd.Name, pkg.derivationTag()))

			logging.Infof(ctx, "command: %+v", cmd)
			return cmd.Run()
		}); err != nil {
			logging.Infof(ctx, "failed to download package from cipd (possible cache miss): %s", err)
		}

		return nil
	})
}

func (pkg cipdPackage) setTags(ctx context.Context, cipdService string, tags []string) error {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return nil
	}

	if len(tags) == 0 {
		return nil
	}

	step, ctx := startStep(ctx, fmt.Sprintf("tagging cipd package %s:%s", cipd.Name, pkg.derivationTag()))
	return step.With(func() error {
		cmd := step.cipdCommand("set-tag", cipd.Name,
			"-service-url", cipdService,
			"-version", pkg.derivationTag(),
		)

		for _, tag := range tags {
			cmd.Args = append(cmd.Args, "-tag", tag)
		}

		logging.Infof(ctx, "command: %+v", cmd)
		if err := cmd.Run(); err != nil {
			return errors.Annotate(err, "failed to set-tag for cipd package %s:%s", cipd.Name, pkg.derivationTag()).Err()
		}

		return nil
	})
}

func (pkg cipdPackage) derivationTag() string {
	return "derivation:" + pkg.DerivationID
}

func buildCIPD(ctx context.Context, step *step, name, src, dst string) (Iid string, err error) {
	resultFile := dst + ".json"
	cmd := step.cipdCommand("pkg-build",
		"-name", name,
		"-in", src,
		"-out", dst,
		"-json-output", resultFile,
	)

	logging.Infof(ctx, "command: %+v", cmd)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	f, err := os.Open(resultFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var result struct {
		Result struct {
			Package    string
			InstanceID string `json:"instance_id"`
		}
	}
	if err := json.NewDecoder(f).Decode(&result); err != nil {
		return "", err
	}

	return result.Result.InstanceID, nil
}

func registerCIPD(ctx context.Context, step *step, cipdService, pkg string, tags []string) error {
	cmd := step.cipdCommand("pkg-register", pkg,
		"-service-url", cipdService,
	)

	for _, tag := range tags {
		cmd.Args = append(cmd.Args, "-tag", tag)
	}

	logging.Infof(ctx, "command: %+v", cmd)
	return cmd.Run()
}

func (s *step) cipdCommand(arg ...string) *exec.Cmd {
	cipd, err := exec.LookPath("cipd")
	if err != nil {
		cipd = "cipd"
	}

	var cmd *exec.Cmd
	// Use cmd to execute batch file on windows.
	if filepath.Ext(cipd) == ".bat" {
		cmd = exec.Command("cmd.exe", append([]string{"/C", cipd}, arg...)...)
	} else {
		cmd = exec.Command(cipd, arg...)
	}
	cmd.Stdout = s.Stdout()
	cmd.Stderr = s.Stderr()

	return cmd
}
