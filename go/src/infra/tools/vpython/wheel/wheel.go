// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package wheel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/luci/luci-go/common/errors"
)

// Name is a parsed Python wheel name, defined here:
// https://www.python.org/dev/peps/pep-0427/#file-name-convention
//
// {distribution}-{version}(-{build tag})?-{python tag}-{abi tag}-\
// {platform tag}.whl .
type Name struct {
	Distribution string
	Version      string
	BuildTag     string
	PythonTag    string
	ABITag       string
	PlatformTag  string
}

func (wn *Name) String() string {
	return strings.Join([]string{
		wn.Distribution,
		wn.Version,
		wn.BuildTag,
		wn.PythonTag,
		wn.ABITag,
		wn.PlatformTag,
	}, "-") + ".whl"
}

// ParseName parses a wheel Name from its filename.
func ParseName(v string) (wn Name, err error) {
	base := strings.TrimSuffix(v, ".whl")
	if len(base) == len(v) {
		err = errors.Reason("missing .whl suffix").Err()
		return
	}

	skip := 0
	switch parts := strings.Split(base, "-"); len(parts) {
	case 6:
		// Extra part: build tag.
		wn.BuildTag = parts[2]
		skip = 1
		fallthrough

	case 5:
		wn.Distribution = parts[0]
		wn.Version = parts[1]
		wn.PythonTag = parts[2+skip]
		wn.ABITag = parts[3+skip]
		wn.PlatformTag = parts[4+skip]

	default:
		err = errors.Reason("unknown number of segments (%(segments)d)").
			D("segments", len(parts)).
			Err()
		return
	}
	return
}

// GlobFrom identifies all wheel files in the directory dir and returns their
// parsed wheel names.
func GlobFrom(dir string) ([]Name, error) {
	globPattern := filepath.Join(dir, "*.whl")
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, errors.Annotate(err).Reason("failed to list wheel directory: %(dir)s").
			D("dir", dir).
			D("pattern", globPattern).
			Err()
	}

	names := make([]Name, 0, len(matches))
	for _, match := range matches {
		switch st, err := os.Stat(match); {
		case err != nil:
			return nil, errors.Annotate(err).Reason("failed to stat wheel: %(path)s").
				D("path", match).
				Err()

		case st.IsDir():
			// Ignore directories.
			continue

		default:
			// A ".whl" file.
			name := filepath.Base(match)
			wheelName, err := ParseName(name)
			if err != nil {
				return nil, errors.Annotate(err).Reason("failed to parse wheel from: %(name)s").
					D("name", name).
					D("dir", dir).
					Err()
			}
			names = append(names, wheelName)
		}
	}
	return names, nil
}

// WriteRequirementsFile writes a valid "requirements.txt"-style pip reuirements
// file containing the supplied wheels.
//
// The generated requirements will request the exact wheel version.
func WriteRequirementsFile(path string, wheels []Name) error {
	fd, err := os.Create(path)
	if err != nil {
		return errors.Annotate(err).Reason("failed to create requirements file").Err()
	}

	// Emit a series of "Distribution==Version" strings.
	seen := make(map[Name]struct{}, len(wheels))
	for _, wheel := range wheels {
		// Only mention a given Distribution/Version once.
		archetype := Name{
			Distribution: wheel.Distribution,
			Version:      wheel.Version,
		}
		if _, ok := seen[archetype]; ok {
			// Already seen a package for this archetype, skip it.
			continue
		}
		seen[archetype] = struct{}{}

		if _, err := fmt.Fprintf(fd, "%s==%s\n", archetype.Distribution, archetype.Version); err != nil {
			fd.Close()
			return errors.Annotate(err).Reason("failed to write to requirements file").Err()
		}
	}

	if err := fd.Close(); err != nil {
		return errors.Annotate(err).Reason("failed to Close").Err()
	}
	return nil
}
