// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package git

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"

	"go.chromium.org/luci/common/api/gerrit"
)

var (
	runnerImpl runner = realRunner{}
)

type runner interface {
	run(ctx context.Context, dir string, stdoutBuf, stderrBuf *bytes.Buffer, name string, args ...string) error
}

type realRunner struct{}

func (c realRunner) run(ctx context.Context, dir string, stdoutBuf, stderrBuf *bytes.Buffer, name string, args ...string) error {
	stdoutBuf.Reset()
	stderrBuf.Reset()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	cmd.Dir = dir
	log.Printf("running %s %s", name, strings.Trim(fmt.Sprint(args), "[]"))
	err := cmd.Run()
	log.Printf("stdout\n%s", stdoutBuf.String())
	log.Printf("stderr\n%s", stderrBuf.String())
	return err
}

// Clone does a `git clone` on the provided repo URL into a subdirectory of the supplied dir, and
// returns the path to the folder of the checkout.
func Clone(ctx context.Context, url string, branch string, parentDir string) (string, error) {
	dir, err := ioutil.TempDir(parentDir, "gitclone")
	if err != nil {
		return dir, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	var stdoutBuf, stderrBuf bytes.Buffer
	cloneCmd := []string{"clone", "--depth=1", url, "-b", branch, "."}
	if err := runnerImpl.run(ctx, dir, &stdoutBuf, &stderrBuf, "git", cloneCmd...); err != nil {
		return dir, errors.New(stderrBuf.String())
	}
	return dir, nil
}

// FetchAndCherryPick attempts to cherry-pick a provided Gerrit revision into
// the provided local Git repo. It returns a bool indicating whether further
// cherry picks can be performed. It also returns an error if this fails (e.g.
// if the cherry-pick won't merge successfully) or nil if the cherry-pick works
// alright.
//
// Invocations of this method alter the supplied Git repo, so the order of
// invocations is important.
func FetchAndCherryPick(ctx context.Context, revision *gerrit.RevisionInfo, url string, repoDir string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	var stdoutBuf, stderrBuf bytes.Buffer
	if err := runnerImpl.run(
		ctx, repoDir, &stdoutBuf, &stderrBuf, "git", "fetch", "--depth=2", url, revision.Ref); err != nil {
		return false, errors.New(stderrBuf.String())
	}
	if err := runnerImpl.run(
		ctx, repoDir, &stdoutBuf, &stderrBuf, "git", "show", "-s", "--pretty=%p", "FETCH_HEAD"); err != nil {
		return false, errors.New(stderrBuf.String())
	}
	parentCommits := strings.Split(strings.Trim(stdoutBuf.String(), "\n"), " ")
	if len(parentCommits) > 1 {
		log.Printf("Found multiple parent commits, indicating a merge commit. "+
			"We are currently unable to validate this sort of situation. Aborting... %v", parentCommits)
		return false, nil
	}

	if err := runnerImpl.run(
		ctx, repoDir, &stdoutBuf, &stderrBuf, "git", "log", "--format=%B", "-n", "1", "FETCH_HEAD"); err != nil {
		return false, errors.New(stderrBuf.String())
	}
	commitMsg := stdoutBuf.String()
	if strings.Contains(commitMsg, "---") || strings.Contains(commitMsg, "+++") {
		log.Printf("It looks like this commit message contains a diff. That would break the next part of this program. See https://crbug.com/1031306. Aborting...")
		return false, nil
	}

	log.Printf("creating patch of %s in %s", revision.Ref, repoDir)
	// Use a big --unified value to make incorrect cherry-picks less likely.
	// This effectively means the patch will contain the entirety of each
	// changed file.
	formatPatchCmd := []string{"format-patch", "--unified=100000000", "FETCH_HEAD^1..FETCH_HEAD"}
	if err := runnerImpl.run(ctx, repoDir, &stdoutBuf, &stderrBuf, "git", formatPatchCmd...); err != nil {
		return false, errors.New(stderrBuf.String())
	}
	patchFile := strings.Trim(stdoutBuf.String(), "\n")
	log.Printf("patching branch")
	// We exclude binary files from the apply because of b/198542075.
	// This tool should probably be rewritten anyways.
	args := []string{"apply", "--3way", "--ignore-whitespace", patchFile}
	ignore_filetypes := []string{
		"7z", "aiqb", "a", "bin", "binaryproto", "bmp", "bz2", "csbin",
		"db", "ddc", "dll", "docx", "dv", "efi", "elf", "exe", "fw",
		"gif", "gz", "hex", "jar", "jpeg", "jpg", "jsonproto", "lib", "mp3",
		"mp4", "ogg", "opus", "pcap", "pdf", "pkg", "png", "pnvm",
		"rar", "raw", "rtf", "sbin", "sfi", "so", "spkg", "svg", "sys",
		"tlv", "ucode", "vbt", "wav", "webp", "whl", "xlsx", "xz", "zip",
	}
	for _, filetype := range ignore_filetypes {
		args = append(args, "--exclude", "*."+filetype)
	}
	// Allow list certain extension-less files: b/205959262
	// Git 2.34 was released recently (around 11/15/21) but has not yet fully
	// propagated across the fleet.
	// TODO(b/198542075): Delete this.
	dirs := []string{"apl", "cml", "tgl", "whl", "glk", "skl", "aml", "kbl", "jsl"}
	for _, dir := range dirs {
		args = append(args, "--exclude", fmt.Sprintf("chromeos-base/%s-ucode-firmware-private/files/*", dir))
	}
	if err := runnerImpl.run(ctx, repoDir, &stdoutBuf, &stderrBuf, "git", args...); err != nil {
		errStr := stderrBuf.String()
		if strings.Contains(errStr, "information is lacking or useless") {
			log.Printf("This looks like a case of https://crbug.com/1031306, in which something in the diff represents a non-existent file. Aborting...")
			return false, nil
		}
		stdoutLineScanner := bufio.NewScanner(strings.NewReader(stdoutBuf.String()))
		for stdoutLineScanner.Scan() {
			line := stdoutLineScanner.Text()
			if strings.HasPrefix(line, "CONFLICT") {
				errStr += line
				errStr += "\n"
			}
		}
		return false, errors.New(errStr)
	}
	return true, nil
}
