// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/maruel/subcommands"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/flag/stringlistflag"
	"go.chromium.org/luci/common/flag/stringmapflag"
	"go.chromium.org/luci/common/logging"

	"infra/cmd/cloudbuildhelper/cloudbuild"
	"infra/cmd/cloudbuildhelper/docker"
	"infra/cmd/cloudbuildhelper/fileset"
	"infra/cmd/cloudbuildhelper/manifest"
	"infra/cmd/cloudbuildhelper/registry"
	"infra/cmd/cloudbuildhelper/storage"
)

// See cmdBuild help string below.
const inputsHashCanonicalTag = ":inputs-hash"

var cmdBuild = &subcommands.Command{
	UsageLine: "build <target-manifest-path> [...]",
	ShortDesc: "builds a docker image using Google Cloud Build",
	LongDesc: `Builds a docker image using Google Cloud Build.

Either reuses an existing image or builds a new one (see below for details). If
builds a new one, tags it with -canonical-tag.

The canonical tag should identify the exact version of inputs (e.g. it usually
includes git revision or other unique version identifier). It is used as
immutable alias of sources and the resulting image.

If -canonical-tag is set to a literal constant ":inputs-hash", it is calculated
from SHA256 of the tarball with the context directory. This is useful to skip
rebuilding the image if inputs do not change, without imposing any specific
schema of canonical tags.

The "build" command works in multiple steps:
  1. Searches for an existing image with the given -canonical-tag. If it exists,
     assumes the build has already been done and skips the rest of the steps.
     This applies to both deterministic and non-deterministic targets.
  2. Prepares a context directory by evaluating the target manifest YAML,
     resolving tags in Dockerfile and executing local build steps. The result
     of this process is a *.tar.gz tarball that will be sent to Docker daemon.
     See "stage" subcommand for more details.
  3. Calculates SHA256 of the tarball and uses it to construct a Google Storage
     path. If the tarball at that path already exists in Google Storage and
     the target is marked as deterministic in the manifest YAML, examines
     tarball's metadata to find the canonical tag of some previous image built
     from this tarball. If it exists, returns this canonical tag as the result.
  4. If the target is not marked as deterministic, or there's no existing images
     that can be reused, triggers "docker build" via Cloud Build and feeds it
     the uploaded tarball as the context. The result of this process is a new
     docker image.
  5. Pushes this image to the registry under -canonical-tag tag.
  6. Updates metadata of the tarball in Google Storage with the reference to the
     produced image (its SHA256 digest and its canonical tag), so that future
     builds can discover and reuse it, if necessary.

In the very end, regardless of whether a new image was built or some existing
one was reused, pushes the image to the registry under given -tag (or tags), if
any. The is primary used to update "latest" tag.
`,

	CommandRun: func() subcommands.CommandRun {
		c := &cmdBuildRun{}
		c.init()
		return c
	},
}

type cmdBuildRun struct {
	commandBase

	targetManifest string
	infra          string
	canonicalTag   string
	buildID        string
	force          bool
	labels         stringmapflag.Value
	tags           stringlistflag.Flag
}

func (c *cmdBuildRun) init() {
	c.commandBase.init(c.exec, true, true, []*string{
		&c.targetManifest,
	})
	c.Flags.StringVar(&c.infra, "infra", "dev", "What section to pick from 'infra' field in the YAML.")
	c.Flags.StringVar(&c.canonicalTag, "canonical-tag", "", "Tag to push the image to if we built a new image.")
	c.Flags.StringVar(&c.buildID, "build-id", "", "Identifier of the CI build that calls this tool (used in various metadata).")
	c.Flags.BoolVar(&c.force, "force", false, "Rebuild and reupload the image, ignoring existing artifacts.")
	c.Flags.Var(&c.labels, "label", "Labels to attach to the docker image, in k=v form.")
	c.Flags.Var(&c.tags, "tag", "Additional tag(s) to unconditionally push the image to (e.g. \"latest\").")
}

func (c *cmdBuildRun) exec(ctx context.Context) error {
	m, err := manifest.Load(c.targetManifest)
	if err != nil {
		return errors.Annotate(err, "when loading manifest").Tag(isCLIError).Err()
	}

	infra, ok := m.Infra[c.infra]
	switch {
	case !ok:
		return errBadFlag("-infra", fmt.Sprintf("no %q infra specified in the manifest", c.infra))
	case infra.Storage == "":
		return errors.Reason("in %q: infra[...].storage is required when using remote build", c.targetManifest).Tag(isCLIError).Err()
	case infra.CloudBuild.Project == "":
		return errors.Reason("in %q: infra[...].cloudbuild.project is required when using remote build", c.targetManifest).Tag(isCLIError).Err()
	}

	// Tags use allowed alphabet.
	if c.canonicalTag != "" && c.canonicalTag != inputsHashCanonicalTag {
		if err := registry.ValidateTag(c.canonicalTag); err != nil {
			return errBadFlag("-canonical-tag", err.Error())
		}
	}
	for _, t := range c.tags {
		if err := registry.ValidateTag(t); err != nil {
			return errBadFlag("-tag", err.Error())
		}
	}

	// If not pushing to a registry, just build and then discard the image. This
	// is accomplished by NOT passing the image name to runBuild.
	image := ""
	if infra.Registry != "" {
		image = path.Join(infra.Registry, m.Name)
	} else {
		// If not using a registry, can't push any tags.
		switch {
		case c.canonicalTag != "":
			return errBadFlag("-canonical-tag", "can't be used if a registry is not specified in the manifest")
		case len(c.tags) != 0:
			return errBadFlag("-tag", "can't be used if a registry is not specified in the manifest")
		}
	}

	// Need a token source to talk to Google Storage and Cloud Build.
	ts, err := c.tokenSource(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to setup auth").Err()
	}

	// Instantiate infra services based on what's in the manifest.
	store, err := storage.New(ctx, ts, infra.Storage)
	if err != nil {
		return errors.Annotate(err, "failed to initialize Storage").Err()
	}
	builder, err := cloudbuild.New(ctx, ts, infra.CloudBuild)
	if err != nil {
		return errors.Annotate(err, "failed to initialize Builder").Err()
	}
	registry := &registry.Client{TokenSource: ts} // can talk to any registry

	res, err := runBuild(ctx, buildParams{
		Manifest:     m,
		Force:        c.force,
		Image:        image,
		Labels:       c.labels,
		BuildID:      c.buildID,
		CanonicalTag: c.canonicalTag,
		Tags:         c.tags,
		Stage:        stage,
		Store:        store,
		Builder:      builder,
		Registry:     registry,
	})
	return c.reportResult(ctx, res, err)
}

// reportResult is called to report the result of the build (successful or not).
func (c *cmdBuildRun) reportResult(ctx context.Context, r buildResult, err error) error {
	if err != nil {
		r.Error = err.Error()
	}

	img := r.Image
	if img == nil && err == nil {
		logging.Infof(ctx, "Image builds successfully") // not using a registry at all
	}
	if img != nil {
		img.Log(ctx, "The final image:")
		r.ViewImageURL = img.ViewURL()
	}

	if jerr := c.writeJSONOutput(&r); jerr != nil {
		return errors.Annotate(jerr, "failed to write JSON output").Err()
	}
	return err
}

// storageImpl is implemented by *storage.Storage.
type storageImpl interface {
	Check(ctx context.Context, name string) (*storage.Object, error)
	Upload(ctx context.Context, name, digest string, r io.Reader) (*storage.Object, error)
	UpdateMetadata(ctx context.Context, obj *storage.Object, cb func(m *storage.Metadata) error) error
}

// builderImpl is implemented by *cloudbuild.Builder.
type builderImpl interface {
	Trigger(ctx context.Context, r cloudbuild.Request) (*cloudbuild.Build, error)
	Check(ctx context.Context, bid string) (*cloudbuild.Build, error)
}

// registryImpl is implemented by *registry.Client.
type registryImpl interface {
	GetImage(ctx context.Context, image string) (*registry.Image, error)
	TagImage(ctx context.Context, img *registry.Image, tag string) error
}

// stageCallback prepares local files and calls 'cb'.
//
// Nominally implemented by 'stage' function.
type stageCallback func(c context.Context, m *manifest.Manifest, cb func(*fileset.Set) error) error

// buildParams are passed to runBuild.
type buildParams struct {
	// Inputs.
	Manifest     *manifest.Manifest // original manifest
	Force        bool               // true to always build an image, ignoring any caches
	Image        string             // full image name to upload (or "" to skip uploads)
	Labels       map[string]string  // extra labels to put into the image
	BuildID      string             // identifier of a CI build that called us
	CanonicalTag string             // a tag to apply to the image if we really built it
	Tags         []string           // extra tags to advance

	// Local build (usually 'stage', mocked in tests).
	Stage stageCallback

	// Infra.
	Store    storageImpl  // where to upload the tarball, mocked in tests
	Builder  builderImpl  // where to build images, mocked in tests
	Registry registryImpl // how to talk to docker registry, mocked in tests
}

// buildResult is returned by runBuild and put into -json-output.
//
// Some fields are populated in reportResult right prior writing to the output.
type buildResult struct {
	Error        string    `json:"error,omitempty"`          // non-empty if the build failed
	Image        *imageRef `json:"image,omitempty"`          // built or reused image (if any)
	ViewImageURL string    `json:"view_image_url,omitempty"` // URL for humans to look at the image (if any)
	ViewBuildURL string    `json:"view_build_url,omitempty"` // URL for humans to look at the Cloud Build log
}

// imageRef is stored as metadata of context tarballs in Google Storage.
//
// It refers to some image built from the tarball.
type imageRef struct {
	Image        string `json:"image"`  // name of the uploaded image "<registry>/<name>"
	Digest       string `json:"digest"` // docker digest of the uploaded image "sha256:..."
	CanonicalTag string `json:"tag"`    // its canonical tag

	BuildID string `json:"build_id,omitempty"` // parent CI build that produced this image (FYI)
}

// buildRef is stored as metadata of context tarballs in Google Storage.
//
// If refers to some CI build that reused the tarball or image built from it.
// This information is retained for debugging.
type buildRef struct {
	BuildID      string `json:"build_id"`      // value of -build-id flag
	CanonicalTag string `json:"tag,omitempty"` // value of -canonical-tag flag
}

// Log dumps information about the image to the log.
func (r *imageRef) Log(ctx context.Context, preamble string) {
	logging.Infof(ctx, "%s", preamble)
	if r.CanonicalTag == "" {
		logging.Infof(ctx, "    Name:   %s", r.Image)
	} else {
		logging.Infof(ctx, "    Name:   %s:%s", r.Image, r.CanonicalTag)
	}
	logging.Infof(ctx, "    Digest: %s", r.Digest)
	logging.Infof(ctx, "    View:   %s", r.ViewURL())
}

// ViewURL returns an URL of the image, for humans.
func (r *imageRef) ViewURL() string {
	if r.CanonicalTag != "" {
		return fmt.Sprintf("https://%s:%s", r.Image, r.CanonicalTag)
	}
	return fmt.Sprintf("https://%s@%s", r.Image, r.Digest)
}

// runBuild is top-level logic of "build" command.
//
// On errors may return partially populated buildResult.
func runBuild(ctx context.Context, p buildParams) (res buildResult, err error) {
	// Skip the build completely if there's already an image with the requested
	// canonical tag. This check is delayed until later if ":inputs-hash" is used
	// as a canonical tag, since we don't know it yet (need to build the tarball
	// in p.Stage first).
	if p.Image != "" && p.CanonicalTag != "" && p.CanonicalTag != inputsHashCanonicalTag {
		if res.Image, err = maybeReuseExistingImage(ctx, p); err != nil {
			return
		}
	}

	// Build the image if haven't found an existing one.
	if res.Image == nil {
		err = p.Stage(ctx, p.Manifest, func(out *fileset.Set) error {
			var err error
			res, err = remoteBuild(ctx, p, out)
			return err
		})
		if err != nil {
			return
		}
	}

	// Attach all requested tags (even if we reused an existing image).
	//
	// Note that res.Image may be nil if we are building the image but not
	// uploading it anywhere (if "registry" is not set in the manifest).
	if res.Image != nil {
		if err := tagImage(ctx, p.Registry, res.Image, p.Tags); err != nil {
			return res, errors.Annotate(err, "tagging the image with -tag(s)").Err()
		}
	}

	return
}

// maybeReuseExistingImage searches for an image with the canonical tag.
//
// Returns:
//   (img, nil) if there's an image we can reuse.
//   (nil, nil) if we need to build a new image.
//   (nil, err) if failed to check.
func maybeReuseExistingImage(ctx context.Context, p buildParams) (*imageRef, error) {
	fullName := fmt.Sprintf("%s:%s", p.Image, p.CanonicalTag)
	switch img, err := getImage(ctx, p.Registry, fullName); {
	case err != nil:
		return nil, err // already annotated
	case img != nil && p.Force:
		logging.Warningf(ctx, "Using -force, will overwrite existing canonical tag %s => %s", p.CanonicalTag, img.Digest)
	case img != nil:
		logging.Infof(ctx, "The canonical tag already exists, skipping the build")
		return &imageRef{
			Image:        p.Image,
			Digest:       img.Digest,
			CanonicalTag: p.CanonicalTag,
		}, nil
	default:
		logging.Infof(ctx, "No such image, will have to build it")
	}
	return nil, nil
}

// remoteBuild executes high level remote build logic.
//
// It takes locally built fileset, uploads it to the storage (if necessary)
// and invokes Cloud Build builder (if necessary).
//
// On errors may return partially populated buildResult.
func remoteBuild(ctx context.Context, p buildParams, out *fileset.Set) (res buildResult, err error) {
	logging.Infof(ctx, "Writing tarball with %d files to a temp file to calculate its hash...", out.Len())
	f, digest, err := writeToTemp(out)
	if err != nil {
		err = errors.Annotate(err, "failed to write the tarball with context dir").Err()
		return
	}

	// Cleanup no matter what. Note that we don't care about IO flush errors in
	// f.Close() as long as uploadToStorage sent everything successfully (as
	// verified by checking the hash there).
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	size, err := f.Seek(0, 1)
	if err != nil {
		err = errors.Annotate(err, "failed to query the size of the temp file").Err()
		return
	}

	logging.Infof(ctx, "Tarball digest: %s", digest)
	logging.Infof(ctx, "Tarball length: %s", humanize.Bytes(uint64(size)))

	// Now that we know the inputs, we can resolve "-canonical-tag :inputs-hash"
	// and do maybeReuseExistingImage check we skipped in `runBuild`.
	if p.CanonicalTag == inputsHashCanonicalTag {
		p.CanonicalTag = "cbh-inputs-" + digest[:24]
		logging.Infof(ctx, "Canonical tag:  %s", p.CanonicalTag)
		res.Image, err = maybeReuseExistingImage(ctx, p)
		if err != nil || res.Image != nil {
			return
		}
	}

	// Upload the tarball (or grab metadata of existing object).
	obj, err := uploadToStorage(ctx, p.Store,
		fmt.Sprintf("%s/%s.tar.gz", p.Manifest.Name, digest),
		digest, f)
	if err != nil {
		return // err is annotated already
	}

	// Metadata about *this* build to associate with the tarball in the storage,
	// even if we reuse an existing tarball or image. This information is retained
	// to simplify debugging.
	buildRef := &buildRef{
		BuildID:      p.BuildID,
		CanonicalTag: p.CanonicalTag,
	}

	// Dump metadata into the log, just FYI. In particular this logs all previous
	// buildRef's that reused this tarball.
	obj.Log(ctx)

	// If the target is marked as deterministic, it means the image is a pure
	// function of the tarball and we can reuse an existing image if we already
	// built something from this tarball.
	determ := p.Manifest.Deterministic != nil && *p.Manifest.Deterministic
	if determ && p.Image != "" && p.CanonicalTag != "" {
		logging.Infof(ctx, "The target is marked as deterministic: looking for existing images built from this tarball...")
		switch imgRef, ts, err := reuseExistingImage(ctx, obj, p.Image, p.Registry); {
		case err != nil:
			return res, err // annotated already
		case imgRef != nil:
			if !p.Force {
				logging.Infof(ctx,
					"Returning an image with canonical tag %q, it was built from this exact tarball %s",
					imgRef.CanonicalTag, humanize.Time(ts))
				res.Image = imgRef
				// Let it be known that we reused the image produced from this tarball.
				return res, updateMetadata(ctx, obj, p.Store, nil, buildRef)
			}
			logging.Warningf(ctx,
				"Using -force, ignoring existing image built from this tarball (%s => %s)",
				imgRef.CanonicalTag, imgRef.Digest)
		default:
			logging.Infof(ctx, "Have no previous images built from this tarball")
		}
	}

	// Trigger Cloud Build build to "transform" the tarball into a docker image.
	imageDigest, build, err := doCloudBuild(ctx, obj, digest, p)
	if build != nil {
		res.ViewBuildURL = build.LogURL
	}
	if err != nil {
		return // err is annotated already
	}
	if p.Image == "" {
		logging.Warningf(ctx, "The registry is not configured, the image wasn't pushed")
		return
	}

	// Our new image.
	res.Image = &imageRef{
		Image:        p.Image,
		Digest:       imageDigest,
		CanonicalTag: p.CanonicalTag,
		BuildID:      p.BuildID,
	}

	if p.CanonicalTag != "" {
		// Apply the canonical tag to the image since we built a new image and need
		// to give it a canonical name.
		if err := tagImage(ctx, p.Registry, res.Image, []string{p.CanonicalTag}); err != nil {
			return res, errors.Annotate(err, "tagging the image with the canonical tag").Err()
		}
		// Modify tarball's metadata to let the future builds know they can reuse
		// the image we've just built. We do it only when using canonical tags,
		// since we want all such "reusable" images to have a readable tag that
		// identifies them.
		if err := updateMetadata(ctx, obj, p.Store, res.Image, buildRef); err != nil {
			return res, err // already annotated
		}
	}

	return
}

////////////////////////////////////////////////////////////////////////////////
// Dealing with the registry.

// getImage asks the registry to resolve "<image>:<tag>" reference.
//
// Returns:
//   (img, nil) if there's such image.
//   (nil, nil) if there's no such image.
//   (nil, err) on errors communicating with the registry.
func getImage(ctx context.Context, r registryImpl, imageRef string) (*registry.Image, error) {
	logging.Infof(ctx, "Checking whether %s already exists...", imageRef)
	switch img, err := r.GetImage(ctx, imageRef); {
	case err == nil:
		return img, nil
	case registry.IsManifestUnknown(err):
		return nil, nil
	default:
		return nil, errors.Annotate(err, "checking existence of %q", imageRef).Err()
	}
}

// tagImage pushes the given image to all given tags (sequentially).
//
// This involves fetching the image manifest first (via its digest) and then
// uploading it back under a new name.
func tagImage(ctx context.Context, r registryImpl, imgRef *imageRef, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	logging.Debugf(ctx, "Fetching the image manifest...")
	img, err := r.GetImage(ctx, fmt.Sprintf("%s@%s", imgRef.Image, imgRef.Digest))
	if err != nil {
		return errors.Annotate(err, "fetching the image manifest").Err()
	}

	for _, t := range tags {
		logging.Infof(ctx, "Tagging %s => %s", t, imgRef.Digest)
		if r.TagImage(ctx, img, t); err != nil {
			return errors.Annotate(err, "pushing tag %q", t).Err()
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Upload to the storage.

// writeToTemp saves the fileset.Set as a temporary *.tar.gz file, returning it
// and its SHA256 hex digest.
//
// The file is opened in read/write mode. The caller is responsible for closing
// and deleting it when done.
func writeToTemp(out *fileset.Set) (*os.File, string, error) {
	f, err := ioutil.TempFile("", "cloudbuildhelper_*.tar.gz")
	if err != nil {
		return nil, "", err
	}
	h := sha256.New()
	if err := out.ToTarGz(io.MultiWriter(f, h)); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, "", err
	}
	return f, hex.EncodeToString(h.Sum(nil)), nil
}

// uploadToStorage uploads the given file to the storage if it's not there yet.
func uploadToStorage(ctx context.Context, s storageImpl, obj, digest string, f *os.File) (*storage.Object, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	switch uploaded, err := s.Check(ctx, obj); {
	case err != nil:
		return nil, errors.Annotate(err, "failed to query the storage for presence of uploaded tarball").Err()
	case uploaded != nil:
		return uploaded, nil
	}

	// Rewind the temp file we have open in read/write mode.
	if _, err := f.Seek(0, 0); err != nil {
		return nil, errors.Annotate(err, "failed to seek inside the temp file").Err()
	}

	uploaded, err := s.Upload(ctx, obj, digest, f)
	return uploaded, errors.Annotate(err, "failed to upload the tarball").Err()
}

////////////////////////////////////////////////////////////////////////////////
// Running Cloud Build.

// doCloudBuild builds and pushes (but not tags) a docker image via Cloud Build.
//
// 'in' is a tarball with the context directory, 'inDigest' is its SHA256 hash.
//
// On success returns "sha256:..." digest of the built and pushed image and
// a Cloud Build build that produced it.
//
// On errors may return a build if the failure happened after the build started.
func doCloudBuild(ctx context.Context, in *storage.Object, inDigest string, p buildParams) (string, *cloudbuild.Build, error) {
	logging.Infof(ctx, "Triggering new Cloud Build build...")

	// Cloud Build always pushes the tagged image to the registry. The default tag
	// is "latest", and we don't want to use it in case someone decides to rely
	// on it. So pick something more cryptic. Note that we don't really care if
	// this tag is moved concurrently by someone else. We never read it, we
	// consume only the image digest returned directly by Cloud Build API.
	image := p.Image
	if image != "" {
		image += ":cbh"
	}
	build, err := p.Builder.Trigger(ctx, cloudbuild.Request{
		Source: in,
		Image:  image,
		Labels: docker.Labels{
			Created:      clock.Now(ctx).UTC(),
			BuildTool:    UserAgent,
			BuildMode:    "cloudbuild",
			Inputs:       inDigest,
			BuildID:      p.BuildID,
			CanonicalTag: p.CanonicalTag,
			Extra:        p.Labels,
		},
	})
	if err != nil {
		return "", nil, errors.Annotate(err, "failed to trigger Cloud Build build").Err()
	}
	logging.Infof(ctx, "Triggered build %s", build.ID)
	logging.Infof(ctx, "Logs are available at %s (may require special permissions to view)", build.LogURL)

	// Babysit it until it completes.
	logging.Infof(ctx, "Waiting for the build to finish...")
	if build, err = waitBuild(ctx, p.Builder, build); err != nil {
		return "", build, errors.Annotate(err, "waiting for the build to finish").Err()
	}
	if build.Status != cloudbuild.StatusSuccess {
		return "", build, errors.Reason("build failed, see its logs at %s", build.LogURL).Err()
	}

	// Make sure Cloud Build worker really consumed the tarball we prepared.
	if got := build.InputHashes[in.String()]; got != inDigest {
		return "", build, errors.Reason("build consumed file with digest %q, but we produced %q", got, inDigest).Err()
	}
	// And it pushed the image we asked it to push.
	if build.OutputImage != image {
		return "", build, errors.Reason("build produced image %q, but we expected %q", build.OutputImage, image).Err()
	}

	return build.OutputDigest, build, nil
}

// waitBuild polls Build until it is in some terminal state (successful or not).
func waitBuild(ctx context.Context, bldr builderImpl, b *cloudbuild.Build) (*cloudbuild.Build, error) {
	errs := 0 // number of errors observed sequentially thus far
	for {
		// Report the status line even if the build is already done, still useful.
		status := string(b.Status)
		if b.StatusDetails != "" {
			status += ": " + b.StatusDetails
		}
		logging.Infof(ctx, "    ... %s", status)

		if b.Status.IsTerminal() {
			return b, nil
		}
		if err := clock.Sleep(clock.Tag(ctx, "sleep-timer"), 5*time.Second).Err; err != nil {
			return nil, err
		}

		build, err := bldr.Check(ctx, b.ID)
		if err != nil {
			if errs++; errs > 5 {
				return nil, errors.Annotate(err, "too many errors, the last one").Err()
			}
			logging.Warningf(ctx, "Error when checking build status - %s", err)
			continue // sleep and try again
		}
		errs = 0

		if build.ID != b.ID {
			return nil, errors.Reason("got unexpected build with ID %q, expecting %q", build.ID, b.ID).Err()
		}
		b = build
	}
}

////////////////////////////////////////////////////////////////////////////////
// Dealing with Cloud Storage metadata.

const (
	imageRefMetaKey = "cbh-image-ref" // metadata key for imageRef{...} JSON blobs
	buildRefMetaKey = "cbh-build-ref" // metadata key for buildRef{...} JSON blobs
)

// reuseExistingImage examines metadata of 'obj' to find references to an
// already built image.
//
// Additionally verifies such image actually exists in the registry. On success
// returns information about the image and an approximate timestamp when it was
// built.
//
// Returns:
//   (ref, ts,   nil) if there's an existing image built from the tarball.
//   (nil, zero, nil) if there's no such image.
//   (nil, zero, err) on errors communicating with the registry.
func reuseExistingImage(ctx context.Context, obj *storage.Object, image string, r registryImpl) (*imageRef, time.Time, error) {
	for _, md := range obj.Metadata.Values(imageRefMetaKey) {
		var ref imageRef
		if err := json.Unmarshal([]byte(md.Value), &ref); err != nil {
			logging.Warningf(ctx, "Skipping bad metadata value %q", md.Value)
			continue
		}
		if ref.Image != image || ref.Digest == "" || ref.CanonicalTag == "" {
			logging.Warningf(ctx, "Skipping inappropriate metadata value %q", md.Value)
			continue
		}

		// Verify such image *actually* exists in the registry.
		switch img, err := getImage(ctx, r, fmt.Sprintf("%s:%s", ref.Image, ref.CanonicalTag)); {
		case err != nil:
			return nil, time.Time{}, err // already annotated
		case img == nil:
			logging.Warningf(ctx, "Metadata record refers to missing image")
		case img.Digest != ref.Digest:
			logging.Warningf(ctx, "Digest of %s:%s in metadata is stale (%q, but the tag points to %q)",
				ref.Image, ref.CanonicalTag, ref.Digest, img.Digest)
		default:
			return &ref, time.Unix(0, md.Timestamp*1000), nil
		}
	}

	return nil, time.Time{}, nil // no images we can reuse
}

// updateMetadata appends to the metadata of the tarball in the storage.
//
// Adds serialized 'img' and 'b' there (if they are non-nil).
func updateMetadata(ctx context.Context, obj *storage.Object, s storageImpl, img *imageRef, b *buildRef) error {
	ts := clock.Now(ctx).UnixNano() / 1000

	var imgRefJSON []byte
	if img != nil {
		var err error
		if imgRefJSON, err = json.Marshal(img); err != nil {
			return errors.Annotate(err, "marshalling imageRef %v", img).Err()
		}
	}

	var buildRefJSON []byte
	if b != nil {
		var err error
		if buildRefJSON, err = json.Marshal(b); err != nil {
			return errors.Annotate(err, "marshalling buildRef %v", b).Err()
		}
	}

	err := s.UpdateMetadata(ctx, obj, func(m *storage.Metadata) error {
		if imgRefJSON != nil {
			m.Add(storage.Metadatum{
				Key:       imageRefMetaKey,
				Timestamp: ts,
				Value:     string(imgRefJSON),
			})
		}
		if buildRefJSON != nil {
			m.Add(storage.Metadatum{
				Key:       buildRefMetaKey,
				Timestamp: ts,
				Value:     string(buildRefJSON),
			})
		}
		m.Trim(50) // to avoid growing metadata size indefinitely
		return nil
	})

	return errors.Annotate(err, "failed to update tarball metadata").Err()
}
