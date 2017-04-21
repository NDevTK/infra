// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"strconv"
	"strings"

	"github.com/luci/luci-go/vpython/api/vpython"
)

type pep425MacArch struct {
	major int
	minor int
	arch  string
}

func parsePEP425MacArch(v string) (a pep425MacArch, is bool) {
	parts := strings.SplitN(v, "_", 4)
	if len(parts) != 4 {
		return
	}
	if parts[0] != "macosx" {
		return
	}

	var err error
	if a.major, err = strconv.Atoi(parts[1]); err != nil {
		return
	}
	if a.minor, err = strconv.Atoi(parts[2]); err != nil {
		return
	}

	a.arch = parts[3]
	is = true
	return
}

func (a *pep425MacArch) less(other pep425MacArch) bool {
	switch {
	case a.major < other.major:
		return true
	case a.minor < other.minor:
		return true
	default:
		return false
	}
}

func pep425IsBetterMacArch(cur, other string) bool {
	// Parse a Mac architecture string
	curArch, curIs := parsePEP425MacArch(cur)
	otherArch, otherIs := parsePEP425MacArch(other)
	switch {
	case !curIs:
		return otherIs
	case !otherIs:
		return false
	case curArch.arch != "intel" && otherArch.arch == "intel":
		// Prefer "intel" architecture over others, since it's more modern and
		// generic.
		return true
	case curArch.arch == "intel" && otherArch.arch != "intel":
		return false
	case otherArch.less(curArch):
		// We prefer the lowest Mac architecture available.
		return true
	default:
		return false
	}
}

func pep425IsBetterLinuxArch(cur, other string) bool {
	// Determies if the specified architecture is a Linux architecture and, if so,
	// is it a "manylinux1_" Linux architecture.
	isLinuxArch := func(arch string) (is bool, many bool) {
		switch {
		case strings.HasPrefix(arch, "linux_"):
			is = true
		case strings.HasPrefix(arch, "manylinux1_"):
			is, many = true, true
		}
		return
	}

	// We prefer "manylinux1_" architectures over "linux_" architectures.
	curIs, curMany := isLinuxArch(cur)
	otherIs, otherMany := isLinuxArch(other)
	switch {
	case !curIs:
		return otherIs
	case !otherIs:
		return false
	case curMany:
		return false
	default:
		return otherMany
	}
}

// pep425TagSelector chooses the "best" PEP425 tag from a set of potential tags.
// This "best" tag will be used to resolve our CIPD templates and allow for
// Python implementation-specific CIPD template parameters.
func pep425TagSelector(goOS string, tags []*vpython.Environment_Pep425Tag) *vpython.Environment_Pep425Tag {
	var best *vpython.Environment_Pep425Tag

	// isPreferredOSArch is an OS-specific architecture preference function.
	isPreferredOSArch := func(best, candidate string) bool { return false }
	switch goOS {
	case "linux":
		isPreferredOSArch = pep425IsBetterLinuxArch
	case "darwin":
		isPreferredOSArch = pep425IsBetterMacArch
	}

	isBetter := func(t *vpython.Environment_Pep425Tag) bool {
		switch {
		case best == nil:
			return true
		case t.Count() > best.Count():
			// More populated fields is more specificity.
			return true
		case best.AnyArch() && !t.AnyArch():
			// More specific architecture is preferred.
			return true
		case !best.HasABI() && t.HasABI():
			// More specific ABI is preferred.
			return true
		case isPreferredOSArch(best.Arch, t.Arch):
			return true
		case strings.HasPrefix(best.Version, "py") && !strings.HasPrefix(t.Version, "py"):
			// Prefer specific Python (e.g., cp27) version over generic (e.g., py27).
			return true

		default:
			return false
		}
	}

	for _, t := range tags {
		if isBetter(t) {
			best = t
		}
	}
	return best
}

// getPEP425CIPDTemplates returns the set of CIPD template strings for a
// given PEP425 tag.
//
// Template parameters are derived from the most representative PEP425 tag.
// Any missing tag parameters will result in their associated template
// parameters not getting exported.
//
// The full set of exported tag parameters is:
// - py_version: The PEP425 Python "version" (e.g., "cp27").
// - py_abi: The PEP425 Python ABI (e.g., "cp27mu").
// - py_arch: The PEP425 Python architecture (e.g., "manylinux1_x86_64").
// - py_tag: The full PEP425 tag (e.g., "cp27-cp27mu-manylinux1_x86_64").
//
// Infra CIPD packages tend to use "${platform}" (generic) combined with
// "${py_abi}" and "${py_arch}" to identify its packages.
func getPEP425CIPDTemplates(goOS string, tags []*vpython.Environment_Pep425Tag) map[string]string {
	tag := pep425TagSelector(goOS, tags)
	if tag == nil {
		return nil
	}

	template := make(map[string]string, 4)
	if tag.Version != "" {
		template["py_version"] = tag.Version
	}
	if tag.Abi != "" {
		template["py_abi"] = tag.Abi
	}
	if tag.Arch != "" {
		template["py_arch"] = tag.Arch
	}
	if tag.Version != "" && tag.Abi != "" && tag.Arch != "" {
		template["py_tag"] = tag.TagString()
	}
	return template
}
