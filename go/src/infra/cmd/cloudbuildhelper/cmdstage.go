// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/cmd/cloudbuildhelper/builder"
	"infra/cmd/cloudbuildhelper/dockerfile"
	"infra/cmd/cloudbuildhelper/fileset"
	"infra/cmd/cloudbuildhelper/manifest"
)

var cmdStage = &subcommands.Command{
	UsageLine: "stage <target-manifest-path> -output-tarball <path> [...]",
	ShortDesc: "prepares the tarball with the context directory",
	LongDesc: `Prepares the tarball with the context directory.

Evaluates input YAML manifest specified via the positional argument, executes
all local build steps there, and rewrites Dockerfile to use pinned digests
instead of tags. Writes the resulting context dir to a *.tar.gz file specified
via "-output-tarball". The contents of this tarball is exactly what will be sent
to the docker daemon or to a Cloud Build worker.

If given "-output-directory" flag, copies the context directory content into
the given output directory instead of writing it into a tarball. Creates the
directory if it doesn't exist.
`,

	CommandRun: func() subcommands.CommandRun {
		c := &cmdStageRun{}
		c.init()
		return c
	},
}

type cmdStageRun struct {
	commandBase

	targetManifest  string
	outputTarball   string
	outputDirectory string
}

func (c *cmdStageRun) init() {
	c.commandBase.init(c.exec, extraFlags{restrictions: true}, []*string{
		&c.targetManifest,
	})
	c.Flags.StringVar(&c.outputTarball, "output-tarball", "", "Where to write the tarball with the context directory.")
	c.Flags.StringVar(&c.outputDirectory, "output-directory", "", "Where to copy the context directory to.")
}

func (c *cmdStageRun) exec(ctx context.Context) error {
	var outputWriter func(out *fileset.Set) error

	switch {
	case c.outputTarball == "" && c.outputDirectory == "":
		return errors.Reason("either -output-tarball or -output-directory flags are required").Tag(isCLIError).Err()

	case c.outputTarball != "" && c.outputDirectory != "":
		return errors.Reason("-output-tarball and -output-directory flags can't be used together").Tag(isCLIError).Err()

	case c.outputTarball != "":
		outputWriter = func(out *fileset.Set) error {
			logging.Infof(ctx, "Writing %d files to %q...", out.Len(), c.outputTarball)
			hash, err := out.ToTarGzFile(c.outputTarball)
			if err != nil {
				return errors.Annotate(err, "failed to save the output").Err()
			}
			logging.Infof(ctx, "Resulting tarball SHA256 is %q", hash)
			return nil
		}

	case c.outputDirectory != "":
		outputWriter = func(out *fileset.Set) error {
			logging.Infof(ctx, "Writing %d files to %q...", out.Len(), c.outputDirectory)
			if err := out.Materialize(c.outputDirectory); err != nil {
				return errors.Annotate(err, "failed to save the output").Err()
			}
			return nil
		}

	default:
		panic("impossible")
	}

	m, _, err := c.loadManifest(ctx, c.targetManifest, false, false)
	if err != nil {
		return err
	}
	return stage(ctx, m.Manifest, outputWriter)
}

// stage executes logic of 'stage' subcommand, calling the callback in the
// end to handle the resulting fileset.
func stage(ctx context.Context, m *manifest.Manifest, cb func(*fileset.Set) error) error {
	// If the manifest had only `contextdir` field without `dockerfile` one, try
	// to find the default Dockerfile (just like Docker itself does), but don't
	// fail if it is missing (this is often the case for e.g GAE tarballs).
	var dockerFilePath string
	var dockerFileRequired bool
	if m.Dockerfile != "" {
		dockerFilePath = m.Dockerfile
		dockerFileRequired = true
	} else if m.ContextDir != "" {
		dockerFilePath = filepath.Join(m.ContextDir, "Dockerfile")
	}

	// Load Dockerfile and resolve image tags there into digests using pins.yaml.
	var dockerFileBody []byte
	if dockerFilePath != "" {
		var err error
		switch dockerFileBody, err = dockerfile.LoadAndResolve(dockerFilePath, m.ImagePins); {
		case errors.Is(err, fs.ErrNotExist) && !dockerFileRequired:
			dockerFilePath = "" // it is missing and it is OK
		case err != nil:
			if pin := dockerfile.IsMissingPinErr(err); pin != nil {
				logging.Errorf(ctx, "------------------------------------------------------------------------")
				logging.Errorf(ctx, "Dockerfile refers to %q which is not pinned in %q", pin.ImageRef(), m.ImagePins)
				logging.Errorf(ctx, "Add a pin there first by running:")
				logging.Errorf(ctx, "  $ cloudbuildhelper pins-add %q %q", m.ImagePins, pin.ImageRef())
				logging.Errorf(ctx, "------------------------------------------------------------------------")
				return isCLIError.Apply(err)
			}
			return errors.Annotate(err, "resolving Dockerfile").Err()
		}
	}

	// Execute all build steps to get the resulting fileset.Set.
	b, err := builder.New()
	if err != nil {
		return errors.Annotate(err, "failed to initialize Builder").Err()
	}
	defer b.Close()
	out, err := b.Build(ctx, m)
	if err != nil {
		return errors.Annotate(err, "local build failed").Err()
	}

	// Append resolved Dockerfile to outputs (perhaps overwriting an existing
	// unresolved one). In tarballs produced by cloudbuildhelper the Dockerfile
	// *always* lives in the root of the context directory.
	if dockerFilePath != "" {
		if err := out.AddFromMemory("Dockerfile", dockerFileBody, nil); err != nil {
			return errors.Annotate(err, "failed to add Dockerfile to output").Err()
		}
	}

	// Let the callback do the rest.
	if err := cb(out); err != nil {
		return err
	}
	return b.Close()
}
