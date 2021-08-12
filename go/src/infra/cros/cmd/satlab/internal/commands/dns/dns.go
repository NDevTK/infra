// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dns

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/paths"

	"go.chromium.org/luci/common/errors"
)

// A classifier takes a line and determines whether to keep, remove, or modify it.
type classifier func(map[string]bool, string) commands.Decision

// A replacer takes a line that is selected to be modified and modifies it.
type replacer func(string) string

// readContents gets the content of a DNS file.
// TODO(gregorynisbet): inline this function.
func readContents() (string, error) {
	args := []string{
		paths.DockerPath,
		"exec",
		"dns",
		"/bin/cat",
		"/etc/dut_hosts/hosts",
	}
	out, err := exec.Command(args[0], args[1:]...).Output()
	return strings.TrimRight(string(out), "\n\t"), errors.Annotate(err, "get dns file content").Err()
}

// WriteBackup set the content of the backup DNS file.
func writeBackup(content string) error {
	name, err := commands.MakeTempFile(content)
	if err != nil {
		return errors.Annotate(err, "set backup dns file content").Err()
	}
	args := []string{
		paths.DockerPath,
		"cp",
		name,
		"dns:/etc/dut_hosts/hosts.BAK",
	}
	err = exec.Command(args[0], args[1:]...).Run()
	return errors.Annotate(err, fmt.Sprintf("set backup dns file content: running %s", strings.Join(args, " "))).Err()
}

// SetDNSFileContent set the content of the DNS file.
func SetDNSFileContent(content string) error {
	name, err := commands.MakeTempFile(content)
	if err != nil {
		return errors.Annotate(err, "set dns file content").Err()
	}
	args := []string{
		paths.DockerPath,
		"cp",
		name,
		"dns:/etc/dut_hosts/hosts",
	}
	err = exec.Command(args[0], args[1:]...).Run()
	return errors.Annotate(err, fmt.Sprintf("set backup dns file content: running %s", strings.Join(args, " "))).Err()
}

// ForceReloadDNSMasqProcess sends the hangup signal to the dnsmasq process inside the dns container
// and forces it to reload its config.
func ForceReloadDNSMasqProcess() error {
	args := []string{
		paths.DockerPath,
		"exec",
		"dns",
		"/bin/sh",
		"-c",
		"/usr/bin/killall -HUP dnsmasq",
	}
	err := exec.Command(args[0], args[1:]...).Run()
	return errors.Annotate(err, "hup dns process").Err()
}

// EnsureRecords ensures that the given DNS records in question are up to date with respect to
// a map mapping hostnames to addresses.
func ensureRecords(content string, newRecords map[string]string) error {
	// Set the backup DNS file so that the user can see the previous state.
	if err := writeBackup(content); err != nil {
		return errors.Annotate(err, "ensure dns records").Err()
	}

	seen := make(map[string]bool)

	classifier := makeClassifier(newRecords)
	replacer := func(line string) string {
		words := strings.Fields(line)
		if len(words) < 2 {
			return ""
		}
		return fmt.Sprintf("%s\t%s", newRecords[words[1]], words[1])
	}

	newContent, err := replaceLineContents(
		seen,
		strings.Split(content, "\n"),
		classifier,
		replacer,
	)

	for host, addr := range newRecords {
		if seen[host] {
			// Do nothing, line already added.
		} else {
			fmt.Fprintf(os.Stderr, "Adding new DNS entry for %s", host)
			newContent = append(newContent, fmt.Sprintf("%s\t%s\n", addr, host))
		}
	}

	if err != nil {
		return errors.Annotate(err, "ensure dns records").Err()
	}
	if err := SetDNSFileContent(strings.Join(newContent, "\n")); err != nil {
		return errors.Annotate(err, "ensure dns records").Err()
	}
	if err := ForceReloadDNSMasqProcess(); err != nil {
		return errors.Annotate(err, "ensure dns records").Err()
	}
	return nil
}

// MakeClassifier makes a classifier that determines whether to modify a given addr, host line or not.
func makeClassifier(newRecords map[string]string) classifier {
	// Nth takes the elements and the given index and safely accesses the string
	// at that index or returns "" if no such string exists.
	nth := func(els []string, idx int) string {
		if idx >= len(els) {
			return ""
		}
		return els[idx]
	}

	// Classifier takes a map of hostnames that have seen before and the current line
	// and determines how to transform it.
	classifier := func(seen map[string]bool, line string) commands.Decision {
		words := strings.Fields(line)
		// Keep blank lines.
		if len(words) == 0 {
			return commands.Keep
		}
		// Keep comments.
		if strings.HasPrefix(nth(words, 0), "#") {
			return commands.Keep
		}
		// Modify lines of the form: addr host.
		// Discard lines of this form after the first one has been
		// processed.
		if _, ok := newRecords[nth(words, 1)]; ok {
			host := nth(words, 1)
			if _, alreadySeen := seen[host]; !alreadySeen {
				seen[host] = true
				return commands.Modify
			}
			return commands.Reject
		}
		return commands.Keep
	}
	return classifier
}

// ReplaceLineContents walks a sequence of lines and keeps, modifies, or removes each line
// according to the classifier and replacer.
func replaceLineContents(seen map[string]bool, lines []string, classifier classifier, replacer replacer) ([]string, error) {
	if seen == nil {
		return nil, errors.New("replace line contents: map cannot be nil")
	}
	var out []string
	for _, line := range lines {
		decision := classifier(seen, line)
		switch decision {
		case commands.Unknown:
			return nil, errors.New("replace line contents: unexpected decision")
		case commands.Keep:
			out = append(out, line)
		case commands.Modify:
			out = append(out, replacer(line))
		case commands.Reject:
			continue
		default:
			return nil, errors.New("replace line contents: unrecognized decision")
		}
	}
	return out, nil
}

// UpdateRecord ensures that the contents of the /etc/hosts file in the dns container are up to date
// with a given host and address.
// UpdateRecord returns the original contents before modification, to allow its caller to undo the modification.
func UpdateRecord(host string, addr string) (string, error) {
	if host == "" {
		return "", errors.New("update record: no hostname")
	}
	if addr == "" {
		return "", errors.New("update record: no address")
	}
	content, err := readContents()
	if err != nil {
		return "", errors.Annotate(err, "update record").Err()
	}
	if err := ensureRecords(content, map[string]string{host: addr}); err != nil {
		return "", errors.Annotate(err, "update record").Err()
	}
	return content, nil
}
