package cloudsdk

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"infra/libs/vmlab/api"
)

const (
	project = "betty-cloud-prototype"
	license = "https://www.googleapis.com/compute/v1/projects/vm-options/global/licenses/enable-vmx"
	timeout = 10 * time.Minute
)

type cloudsdkImageApi struct{}

// New constructs a new api.ImageApi with CloudSDK backend.
func New() (api.ImageApi, error) {
	return &cloudsdkImageApi{}, nil
}

// computeImagesClient interfaces partial GCE client API
type computeImagesClient interface {
	Get(context.Context, *computepb.GetImageRequest, ...gax.CallOption) (*computepb.Image, error)
	Insert(context.Context, *computepb.InsertImageRequest, ...gax.CallOption) (*compute.Operation, error)
}

func (c *cloudsdkImageApi) GetImage(buildPath string, wait bool) (*api.GceImage, error) {
	imageName, err := convertName(buildPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert image name from GCS to GCE format: %v", err)
	}

	gceImage := &api.GceImage{
		Project: project,
		Name:    imageName,
		Source:  getGcsImagePath(buildPath),
	}

	var ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewImagesRESTClient: %v", err)
	}
	defer client.Close()

	op, gceImage, err := c.handle(client, gceImage)

	if wait {
		if op != nil {
			err = op.Wait(ctx)
		} else {
			return c.poll(ctx, client, gceImage)
		}
	}
	return c.describeImage(client, gceImage)
}

func (c *cloudsdkImageApi) handle(client computeImagesClient, gceImage *api.GceImage) (*compute.Operation, *api.GceImage, error) {
	gceImage, _ = c.describeImage(client, gceImage)

	if gceImage.GetStatus() == api.GceImage_NOT_FOUND {
		op, err := c.importImage(client, gceImage)
		if err == nil {
			gceImage.Status = api.GceImage_PENDING
		}
		return op, gceImage, err
	}

	return nil, gceImage, nil
}

func (c *cloudsdkImageApi) poll(ctx context.Context, client computeImagesClient, gceImage *api.GceImage) (*api.GceImage, error) {
	return gceImage, poll(ctx, func(ctx context.Context) (bool, error) {
		_, gceImage, err := c.handle(client, gceImage)
		if gceImage.GetStatus() == api.GceImage_READY {
			return true, err
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, 30*time.Second)
}

// getPendingOperation finds the pending LRO, which is not useful, polling is still needed.
// TODO(mingkong): delete the method
func (c *cloudsdkImageApi) getPendingOperation(gceImage *api.GceImage) (*computepb.Operation, error) {
	ctx := context.Background()
	client, _ := compute.NewGlobalOperationsRESTClient(ctx)
	defer client.Close()
	filter := fmt.Sprintf("status=RUNNING AND operationType=insert AND targetLink=\"https://www.googleapis.com/compute/v1/projects/mingkong-cros/global/images/%s\"", gceImage.GetName())
	limit := uint32(1)
	req := &computepb.ListGlobalOperationsRequest{
		Project:    gceImage.Project,
		Filter:     &filter,
		MaxResults: &limit,
	}
	it := client.List(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return nil, nil
}

// describeImage detects the readiness of the given GCE image: READ, PENDING, or
// NOT_FOUND, which means import is needed.
func (c *cloudsdkImageApi) describeImage(client computeImagesClient, gceImage *api.GceImage) (*api.GceImage, error) {
	ctx := context.Background()

	// Create the gceImage
	req := &computepb.GetImageRequest{
		Image:   gceImage.Name,
		Project: gceImage.Project,
	}
	i, err := client.Get(ctx, req)

	gceImage.Status = api.GceImage_NOT_FOUND
	if i != nil {
		switch i.GetStatus() {
		case "READY":
			gceImage.Status = api.GceImage_READY
		case "PENDING":
			gceImage.Status = api.GceImage_PENDING
		}
	}

	return gceImage, err
}

// importImage imports an image from GCS into GCE
func (c *cloudsdkImageApi) importImage(client computeImagesClient, image *api.GceImage) (*compute.Operation, error) {
	ctx := context.Background()

	// Create the image
	req := &computepb.InsertImageRequest{
		ImageResource: &computepb.Image{
			Licenses: []string{license},
			Name:     &image.Name,
			RawDisk: &computepb.RawDisk{
				Source: &image.Source,
			},
		},
		Project: image.Project,
	}

	return client.Insert(ctx, req)
}

// convertName constructs GCE image name from the build path in GCS.
// buildPath is the path between "chromeos-image-archive/" and "/chromiumos_test_image_gce.tar.gz"
// as found in autotest_keyvals in CtrRequest. Examples:
// betty-arc-r-cq/R108-15164.0.0-71927-8801111609984657185
// betty-arc-r-release/R108-15178.0.0
// The image name convention is to swap the order of two directories, rejoin
// with --, replace any . and _ with -, in lowercase and maximum length of 63.
// After conversion, the above examples would become:
// r108-15164-0-0-71927-8801111609984657185--betty-arc-r-cq
// r108-15178-0-0--betty-arc-r-release
func convertName(buildPath string) (string, error) {
	// Split the source build path
	bucket, build, found := strings.Cut(buildPath, "/")
	if !found {
		return "", fmt.Errorf("Invalid build path format: %s", buildPath)
	}

	// Switch order of build and bucket
	imageName := fmt.Sprintf("%s--%s", build, bucket)

	// Shorten name
	if len(imageName) > 63 {
		imageName = imageName[0:62]
	}

	// Convert to legit GCE name
	imageName = strings.ToLower(imageName)
	imageName = strings.ReplaceAll(imageName, ".", "-")
	imageName = strings.ReplaceAll(imageName, "_", "-")

	return imageName, nil
}

// getGcsImagePath returns the full path of the tarball on GCS.
func getGcsImagePath(buildPath string) string {
	return fmt.Sprintf("https://storage.googleapis.com/chromeos-image-archive/%s/chromiumos_test_image_gce.tar.gz", buildPath)
}

// poll provides a generic implementation of calling f at interval, exits on
// error or ctx timeout. f return true to end poll early.
func poll(ctx context.Context, f func(context.Context) (bool, error), interval time.Duration) error {
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("context must have a deadline to avoid infinite polling")
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			success, err := f(ctx)
			if err != nil {
				return err
			}
			if success {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
