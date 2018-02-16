package frontend

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"infra/appengine/test-results/model"

	"golang.org/x/net/context"

	"cloud.google.com/go/storage"
	"go.chromium.org/gae/service/memcache"
	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
)

// getZipHandler handles a request to get a file from a zip archive.
// This saves content etags in memcache, to save round trip time on fetching
// zip files over the network, so that clients can cache the data.
func getZipHandler(ctx *router.Context) {
	c, w, r, p := ctx.Context, ctx.Writer, ctx.Request, ctx.Params

	builder := p.ByName("builder")
	buildNum := p.ByName("buildnum")
	filepath := strings.Trim(p.ByName("filepath"), "/")

	// Special case, since this isn't the zip file.
	if filepath == "layout-test-results.zip" {
		newURL := fmt.Sprintf("https://storage.googleapis.com/chromium-layout-test-archives/%s/%s/%s", builder, buildNum, filepath)
		http.Redirect(w, r, newURL, http.StatusPermanentRedirect)
		return
	}
	mkey := fmt.Sprintf("gs_etag%s/%s/%s", builder, buildNum, filepath)

	// This content should never change; safe to cache for 1 day.
	w.Header().Set("Cache-Control", "public, max-age=86400")

	// Check to see if the client has this cached on their side.
	ifNoneMatch := r.Header.Get("If-None-Match")
	itm := memcache.NewItem(c, mkey)
	if ifNoneMatch != "" {
		err := memcache.Get(c, itm)
		if err == nil && r.Header.Get("If-None-Match") == string(itm.Value()) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	contents, err := getZipFile(c, builder, buildNum, filepath)
	if err != nil {
		panic(err)
	}

	if contents == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}

	h := sha256.New()
	h.Write(contents)
	itm.SetValue([]byte(fmt.Sprintf("%x", h.Sum(nil))))
	err = memcache.Set(c, itm)
	if err != nil {
		logging.Warningf(c, "Error while setting memcache key for etag digest: %v", err)
	} else {
		w.Header().Set("ETag", string(itm.Value()))
	}

	// The order of these statements matters. See net/http docs for more info.
	w.Header().Set("Content-Type", http.DetectContentType(contents))
	w.WriteHeader(http.StatusOK)
	w.Write(contents)
}

const megabyte = 1 << 20
const chunkSize = megabyte * 31

// knownPrefixes is a list of strings that we know will exist in the google storage bucket.
// If the first path element in the filepath is not on this list, it's assumed to be a substep,
// and the zip file in that subdirectory is used, instead of at the root of the build number bucket.
var knownPrefixes = []string{
	"layout-test-results",
	"retry_summary.json",
}

// getZipFile retrieves a file from a layout test archive for a build number from a builder.
var getZipFile = func(c context.Context, builder, buildNum, filepath string) ([]byte, error) {
	prefix := ""
	found := false
	for _, prefix := range knownPrefixes {
		if strings.HasPrefix(filepath, prefix) {
			found = true
		}
	}

	if !found && strings.Contains(filepath, "/") {
		prefix = strings.Split(filepath, "/")[0] + "/"
		filepath = strings.Join(strings.Split(filepath, "/")[1:], "/")
	}
	gsPath := gs.Path(fmt.Sprintf("gs://chromium-layout-test-archives/%s/%s/%slayout-test-results.zip", builder, buildNum, prefix))

	itm := memcache.NewItem(c, fmt.Sprintf("%s|%s", gsPath, filepath))
	err := memcache.Get(c, itm)
	if err != memcache.ErrCacheMiss && err != nil {
		return nil, err
	}

	logging.Debugf(c, "Getting google storage path %s", gsPath)
	if err == memcache.ErrCacheMiss {
		zr, err := readZipFile(c, gsPath)
		if err != nil {
			return nil, fmt.Errorf("while creating zip reader: %v", err)
		}

		// If we're serving the results.html file, we expect users to want to look at
		// the failed test artifacts. Cache these.
		if strings.Contains(filepath, "results.html") {
			if err := cacheFailedTests(c, zr, string(gsPath)); err != nil {
				logging.Warningf(c, "while caching failed tests: %v", err)
			}
		}

		for _, f := range zr.File {
			if f.Name == filepath {
				freader, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("while opening zip file: %v", err)
				}

				res, err := ioutil.ReadAll(freader)
				if err != nil {
					return nil, err
				}
				itm.SetValue(res)
			}
		}

		logging.Debugf(c, "len %v limit %v", len(itm.Value()), megabyte/2)
		if itm.Value() != nil && len(itm.Value()) < megabyte/2 {
			logging.Debugf(c, "setting %s", itm.Key())
			err := memcache.Set(c, itm)
			if err != nil {
				return nil, err
			}
		}
	}

	return itm.Value(), nil
}

// cacheFailedTests caches the failed tests stored in the resulting zip file.
// gsPath is used to construct the memcache keys as needed.
func cacheFailedTests(c context.Context, zr *zip.Reader, gsPath string) error {
	// First, read the results file to find a list of failed tests.
	failedTests := []string{}
	for _, f := range zr.File {
		if f.Name == "layout-test-results/full_results.json" {
			freader, err := f.Open()
			if err != nil {
				return err
			}

			fullResultsDat, err := ioutil.ReadAll(freader)
			if err != nil {
				return err
			}
			failedTests = getFailedTests(c, fullResultsDat)
			break
		}
	}
	if len(failedTests) == 0 {
		return nil
	}

	logging.Debugf(c, "caching artifacts for %v failed tests", len(failedTests))
	toPut := []memcache.Item{}
	for _, f := range zr.File {
		// I tried using goroutines to make this concurrent, but it seemed to be slower.
		for _, test := range failedTests {
			// Ignore retried results, results.html doesn't seem to fetch them
			byPath := strings.Split(f.Name, "/")
			if len(byPath) > 2 && strings.Contains(byPath[1], "retry") {
				break
			}
			if strings.Contains(f.Name, test) {
				newItm := memcache.NewItem(c, fmt.Sprintf("%s|%s", gsPath, f.Name))
				freader, err := f.Open()
				if err != nil {
					logging.Warningf(c, "failed in result caching: %v", err)
					continue
				}

				res, err := ioutil.ReadAll(freader)
				if err != nil {
					logging.Warningf(c, "failed in result caching: %v", err)
					continue
				}
				newItm.SetValue(res)
				toPut = append(toPut, newItm)
			}
		}
	}
	return memcache.Set(c, toPut...)
}

// Tests are named like url-format-any.html. Test artifacts are named like
// url-format-any-diff.png. To check if a test artifact should be cached, we do
// a string.Contains(test, test_artifact). To make this match, we need to strip
// the file extension from the test.
var knownTestArtifactExtensions = []string{
	"html", "svg", "xml",
}

// Gets a list of failed tests from a full results json file.
var getFailedTests = func(c context.Context, fullResultsBytes []byte) []string {
	fr := model.FullResult{}
	err := json.Unmarshal(fullResultsBytes, &fr)
	if err != nil {
		panic(err)
	}
	flattened := fr.Tests.Flatten("/")
	names := []string{}

	for name, test := range flattened {
		if len(test.Actual) != len(test.Expected) {
			continue
		}
		if test.Unexpected == nil || !*test.Unexpected {
			continue
		}
		ue := unexpected(test.Expected, test.Actual)

		hasPass := false
		// If there was a pass at all, count it.
		for _, r := range test.Actual {
			if r == "PASS" {
				hasPass = true
			}
		}

		if len(ue) > 0 && !hasPass {
			for _, suffix := range knownTestArtifactExtensions {
				if strings.HasSuffix(name, "."+suffix) {
					name = name[:len(name)-len("."+suffix)]
				}
			}
			names = append(names, name)
		}
	}
	return names
}

// unexpected returns the set of expected xor actual.
func unexpected(expected, actual []string) []string {
	e, a := make(map[string]bool), make(map[string]bool)
	for _, s := range expected {
		e[s] = true
	}
	for _, s := range actual {
		a[s] = true
	}

	ret := []string{}

	// Any value in the expected set is a valid test result.
	for k := range a {
		if !e[k] {
			ret = append(ret, k)
		}
	}

	return ret
}

var readZipFile = func(c context.Context, gsPath gs.Path) (*zip.Reader, error) {
	transport, err := auth.GetRPCTransport(c, auth.NoAuth)
	if err != nil {
		return nil, fmt.Errorf("while creating transport: %v", err)
	}

	cl, err := gs.NewProdClient(c, transport)
	if err != nil {
		return nil, fmt.Errorf("while creating client: %v", err)
	}

	var offset int64
	allBytes := []byte{}
	for {
		cloudReader, err := cl.NewReader(gsPath, offset, chunkSize)
		if err != nil && err != storage.ErrObjectNotExist {
			return nil, fmt.Errorf("while creating reader: %v", err)
		}
		if err == storage.ErrObjectNotExist {
			return nil, nil
		}

		readBytes, err := ioutil.ReadAll(cloudReader)
		if err != nil {
			return nil, fmt.Errorf("while reading bytes: %v", err)
		}

		allBytes = append(allBytes, readBytes...)
		offset += int64(len(readBytes))
		if len(readBytes) < chunkSize {
			break
		}
	}

	bytesReader := bytes.NewReader(allBytes)
	zr, err := zip.NewReader(bytesReader, int64(len(allBytes)))
	if err != nil {
		return nil, fmt.Errorf("while creating zip reader: %v", err)
	}
	return zr, nil
}
