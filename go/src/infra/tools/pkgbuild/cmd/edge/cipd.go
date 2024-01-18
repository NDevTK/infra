// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/filesystem"
)

type cipdPackage actions.Package

func (pkg cipdPackage) check(ctx context.Context, cipdService string) bool {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return false
	}

	logging.Infof(ctx, "describing cipd package %s:%s", cipd.Name, pkg.derivationTag())

	cmd := cipdCommand("describe", cipd.Name,
		"-service-url", cipdService,
		"-version", pkg.derivationTag(),
	)

	if err := cmd.Run(); err != nil {
		logging.Infof(ctx, "failed to describe cipd package (possible new package) %s:%s", cipd.Name, pkg.derivationTag())
		return false
	}

	return true
}

func (pkg cipdPackage) upload(ctx context.Context, workdir, cipdService, prefix string, tags []string) (name string, iid string, err error) {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return "", "", nil
	}

	name = path.Join(prefix, cipd.Name)
	logging.Infof(ctx, "creating cipd package %s:%s with %s", name, pkg.derivationTag(), cipd.Version)

	out := filepath.Join(workdir, pkg.DerivationID+".cipd")
	if _, err := os.Stat(out); err == nil || !errors.Is(err, fs.ErrNotExist) {
		return "", "", err
	}

	iid, err = buildCIPD(name, pkg.Handler.OutputDirectory(), out)
	if err != nil {
		_ = filesystem.RemoveAll(out)
		return "", "", errors.Annotate(err, "failed to build cipd package").Err()
	}

	if err := registerCIPD(cipdService, out, append([]string{pkg.derivationTag()}, tags...)); err != nil {
		return "", "", errors.Annotate(err, "failed to register cipd package").Err()
	}

	return
}

func (pkg cipdPackage) download(ctx context.Context, cipdService string) error {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return nil
	}

	if err := pkg.Handler.Build(func() error {
		logging.Infof(ctx, "dowloading cipd package %s:%s", cipd.Name, pkg.derivationTag())

		cmd := cipdCommand("export",
			"-service-url", cipdService,
			"-root", pkg.Handler.OutputDirectory(),
			"-ensure-file", "-",
		)
		cmd.Stdin = strings.NewReader(fmt.Sprintf("%s %s", cipd.Name, pkg.derivationTag()))

		return cmd.Run()
	}); err != nil {
		logging.Infof(ctx, "failed to download package from cipd (possible cache miss): %s", err)
	}

	return nil
}

func (pkg cipdPackage) setTags(ctx context.Context, cipdService string, tags []string) error {
	cipd := pkg.Action.Metadata.GetCipd()
	if cipd == nil || cipd.Name == "" {
		return nil
	}

	if len(tags) == 0 {
		return nil
	}

	logging.Infof(ctx, "tagging cipd package %s:%s", cipd.Name, pkg.derivationTag())

	cmd := cipdCommand("set-tag", cipd.Name,
		"-service-url", cipdService,
		"-version", pkg.derivationTag(),
	)

	for _, tag := range tags {
		cmd.Args = append(cmd.Args, "-tag", tag)
	}

	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "failed to set-tag for cipd package %s:%s", cipd.Name, pkg.derivationTag()).Err()
	}

	return nil
}

func (pkg cipdPackage) derivationTag() string {
	return "derivation:" + pkg.DerivationID
}

func buildCIPD(name, src, dst string) (Iid string, err error) {
	resultFile := dst + ".json"
	cmd := cipdCommand("pkg-build",
		"-name", name,
		"-in", src,
		"-out", dst,
		"-json-output", resultFile,
	)

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

func registerCIPD(cipdService, pkg string, tags []string) error {
	cmd := cipdCommand("pkg-register", pkg,
		"-service-url", cipdService,
	)

	for _, tag := range tags {
		cmd.Args = append(cmd.Args, "-tag", tag)
	}

	return cmd.Run()
}

func cipdCommand(arg ...string) *exec.Cmd {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
