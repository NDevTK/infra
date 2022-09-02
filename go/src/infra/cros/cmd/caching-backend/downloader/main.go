// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// downloader is fleetware cache service tool.
// It serves in between cache server and google storage.
// To start the server:
// 		downloader -address <address:port>
//				   -credential-file <service account credential file>
// After started on the specified address,
// it listens on the specified TCP port.

// The server accepts below requests:
//   - HEAD /download/<bucket>/path/to/file
//     Return only the meta data of file in header.
//   - GET /download/<bucket>/path/to/file
//     Download the file from google storage.
//   - GET /extract/<bucket>/path/to/archive-tar?file=path/to/file
//     Download the archive tar and return specified file.
//   - GET /decompress/<bucket>/path/to/comopressed-file
//     Download the compressed file and return the decompressed data.
package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/ulikunitz/xz"
)

var (
	credentialFile       = flag.String("credential-file", "", "credential json file. Example: ./service-credential.json")
	archiveServerAddress = flag.String("address", ":8080", "archive server address with listening port.")
	cacheServerURL       = flag.String("cache-server-url", "http://127.0.0.1:8082", "cache-server url.")
	shutdownGracePeriod  = flag.Duration("shutdown-grace-period", 30*time.Minute, "The time duration allowed for tasks to complete before completely shutdown archive-server.")
)

type archiveServer struct {
	gsClient       gsClient
	cacheServerURL string
}

func main() {
	if err := innerMain(); err != nil {
		log.Fatalf("Exiting due to an error: %s", err)
	}
	log.Printf("Exiting successfully")
}

func innerMain() error {
	flag.Parse()
	ctx := context.Background()
	gsClient, err := newRealClient(ctx, *credentialFile)
	if err != nil {
		return fmt.Errorf("google storage client error: %s", err)
	}
	defer gsClient.close()

	c := &archiveServer{
		gsClient:       gsClient,
		cacheServerURL: *cacheServerURL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/download/", c.downloadHandler)
	mux.HandleFunc("/extract/", c.extractHandler)
	mux.HandleFunc("/decompress/", c.decompressHandler)

	idleConnsClosed := make(chan struct{})
	svr := http.Server{Addr: *archiveServerAddress, Handler: mux}
	ctx = cancelOnSignals(ctx, idleConnsClosed, &svr, *shutdownGracePeriod)
	log.Println("starting archive-server...")
	if err = svr.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
	return err
}

// downloadHandler handles the /download/bucket/path/to/file requests.
// It writes file stat to header for HEAD, GET method.
// It writes file content to body for GET method.
func (c *archiveServer) downloadHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	id := fmt.Sprintf("%s%s", r.Method, r.URL.RequestURI())
	log.Printf("%s request started", id)
	defer func() { log.Printf("%s request completed in %s", id, time.Since(startTime)) }()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	gsClient := c.gsClient

	switch r.Method {
	case http.MethodHead:
		handleDownloadHEAD(ctx, w, r, gsClient, id)
	case http.MethodGet:
		handleDownloadGET(ctx, w, r, gsClient, id)
	default:
		errStr := fmt.Sprintf("%s unsupported method", id)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
	}
}

// handleDownloadHEAD handles download HEAD request.
// It writes file stat to ResponseWriter.
// It returns gsObject which is used by handleDownloadGET to send file content.
func handleDownloadHEAD(ctx context.Context, w http.ResponseWriter, r *http.Request, gsClient gsClient, reqID string) (gsObject, error) {
	objectName, err := parseURL(r.URL.Path)
	if err != nil {
		err := fmt.Errorf("%s parseURL error: %w", reqID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf(err.Error())
		return nil, err
	}

	gsObject := gsClient.getObject(objectName)

	gsAttrs, err := gsObject.Attrs(ctx)
	if err != nil {
		var retStatus int
		if errors.Is(err, storage.ErrObjectNotExist) {
			retStatus = http.StatusNotFound
		} else {
			retStatus = http.StatusInternalServerError
		}
		err := fmt.Errorf("%s Attrs error: %w", reqID, err)
		http.Error(w, err.Error(), retStatus)
		log.Printf(err.Error())
		return nil, err
	}

	writeHeaderAndStatusOK(gsAttrs, w)
	return gsObject, nil
}

// handleDownloadGET handles download GET request.
// It writes file stat to ResponseWriter header, and content to body.
func handleDownloadGET(ctx context.Context, w http.ResponseWriter, r *http.Request, gsClient gsClient, reqID string) {
	gsObject, err := handleDownloadHEAD(ctx, w, r, gsClient, reqID)
	if err != nil {
		return
	}

	rc, err := gsObject.NewReader(ctx)
	if err != nil {
		log.Printf("%s NewReader error: %s", reqID, err)
		return
	}
	defer rc.Close()

	if n, err := io.Copy(w, rc); err != nil {
		log.Printf("%s copy to body failed at byte %v: %s", reqID, n, err)
	}
}

// writeHeaderAndStatusOK writes various attributes to response header.
func writeHeaderAndStatusOK(objAttr *storage.ObjectAttrs, w http.ResponseWriter) {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Hash-CRC32C", convertCRC32CToString(objAttr.CRC32C))
	// Object may or may not have MD5. https://cloud.google.com/storage/docs/hashes-etags
	if objAttr.MD5 != nil {
		w.Header().Set("Content-Hash-MD5", base64.StdEncoding.EncodeToString(objAttr.MD5))
	}
	w.Header().Set("Content-Length", strconv.FormatInt(objAttr.Size, 10))
	w.Header().Set("Content-Type", objAttr.ContentType)
	w.WriteHeader(http.StatusOK)
}

func convertCRC32CToString(i uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return base64.StdEncoding.EncodeToString(b)
}

// parseURL parses URL.
// It returns the storage object name(bucket and object path).
// Typical path for archive server is '/RPC/bucket/...object-path/'.
// After splitting, the fields would be like
// ["", RPC, bucket, ...object-path].
// Example: url = "/download/release/build/image.tar"
// bucket = "release", objectPath = "build/image.tar"
func parseURL(url string) (*gsObjectName, error) {
	fields := strings.Split(url, "/")
	if len(fields) < 4 {
		return nil, fmt.Errorf("the URL doesn't have all of RPC, bucket and object path")
	}
	if fields[2] == "" {
		return nil, fmt.Errorf("bucket cannot be empty")
	}
	path := strings.Join(fields[3:], "/")
	if path == "" {
		return nil, fmt.Errorf("object cannot be empty")
	}
	return &gsObjectName{bucket: fields[2], path: path}, nil
}

// extractHandler handles /extract/bucket/path/to/file?file=target_file requests.
// It extracts target_file from tar writes the stat to header,
// the content to body.
func (c *archiveServer) extractHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	id := fmt.Sprintf("%s%s", r.Method, r.URL.RequestURI())
	log.Printf("%s request started", id)
	defer func() { log.Printf("%s request completed in %s", id, time.Since(startTime)) }()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	switch r.Method {
	case http.MethodHead:
		handleExtract(ctx, w, r, c.cacheServerURL, id, false)
	case http.MethodGet:
		handleExtract(ctx, w, r, c.cacheServerURL, id, true)
	default:
		errStr := fmt.Sprintf("%s unsupported method", id)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
	}
}

// handleExtract handles extract HEAD/GET requests.
// It downloads the tar from cache server(nginx). It then, extracts
// the target file and writes stat to ResponseWriter header.
// If wantBody is true which essentially is GET, it will copy content to
// ResponseWriter body.
func handleExtract(ctx context.Context, w http.ResponseWriter, r *http.Request, cacheServerURL string, reqID string, wantBody bool) {
	objectName, err := parseURL(r.URL.Path)
	if err != nil {
		errStr := fmt.Sprintf("%s parseURL error: %s", reqID, err)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
		return
	}

	queryFile := r.URL.Query().Get("file")
	if queryFile == "" {
		errStr := fmt.Sprintf("%s extract file query not specified from %s", reqID, objectName.path)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
		return
	}

	action := "download"
	if _, ok := compressReaderMap[filepath.Ext(objectName.path)]; ok {
		action = "decompress"
	}
	reqURL := fmt.Sprintf("%s/%s/%s/%s", cacheServerURL, action, objectName.bucket, objectName.path)
	res, err := downloadURL(ctx, w, reqURL, reqID)
	if err != nil {
		return
	}
	defer res.Body.Close()

	tarReader, err := extractTarAndWriteHeader(ctx, res.Body, queryFile, w)
	if err != nil {
		log.Printf(fmt.Sprintf("%s extractTarAndWriteHeader failed: %s", reqID, err))
		return
	}

	if wantBody {
		if n, err := io.Copy(w, tarReader); err != nil {
			log.Printf("%s copy to body failed at byte %v: %s", reqID, n, err)
		}
	}
}

// downloadURL downloads the reqURL and returns the content in response.
// It writes to client header if error occurs or relays non 200 status code
// from upstream.
func downloadURL(ctx context.Context, w http.ResponseWriter, reqURL string, reqID string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		err := fmt.Errorf("%s download request %q: %w", reqID, reqURL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf(err.Error())
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err := fmt.Errorf("%s download %q: %w", reqID, reqURL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf(err.Error())
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		upstreamErr, err := io.ReadAll(res.Body)
		if err != nil {
			err = fmt.Errorf("%s failed to read upstream %v response of %q: %w", reqID, res.StatusCode, reqURL, err)
		} else {
			err = fmt.Errorf("%s %s respond %v status: %s", reqID, reqURL, res.StatusCode, upstreamErr)
		}
		http.Error(w, err.Error(), res.StatusCode)
		log.Printf(err.Error())
		return nil, err
	}
	return res, nil
}

// extractTarAndWriteHeader extracts file from r reader.
// It writes file stat to the header and returns
// the tar reader for GET handling.
func extractTarAndWriteHeader(ctx context.Context, r io.Reader, fileName string, w http.ResponseWriter) (*tar.Reader, error) {
	tarReader := tar.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		header, err := tarReader.Next()
		if err != nil {
			err = fmt.Errorf("tarReader error: %w", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, err
		}

		if header.Typeflag == tar.TypeReg && header.Name == fileName {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", strconv.FormatInt(header.Size, 10))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)

			return tarReader, nil
		}
	}
}

// decompressHandler handles the /decompress/bucket/path/to/file requests.
// It decompresses compressed file and returns content to body for GET method.
func (c *archiveServer) decompressHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	id := fmt.Sprintf("%s%s", r.Method, r.URL.RequestURI())
	log.Printf("%s request started", id)
	defer func() { log.Printf("%s request completed in %s", id, time.Since(startTime)) }()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		handleDecompressGET(ctx, w, r, c.cacheServerURL, id)
	default:
		errStr := fmt.Sprintf("%s unsupported method", id)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
	}
}

type compressReaderFunc func(io.Reader) (io.ReadCloser, error)

func newGZIPReader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func newBZ2Reader(r io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(bzip2.NewReader(r)), nil
}

func newXZReader(r io.Reader) (io.ReadCloser, error) {
	dReader, err := xz.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("new xz reader: %w", err)
	}
	return io.NopCloser(dReader), nil
}

var compressReaderMap = map[string]compressReaderFunc{
	".gz":  newGZIPReader,
	".tgz": newGZIPReader,
	".bz2": newBZ2Reader,
	".xz":  newXZReader,
}

// handleDecompressGET handles decompress GET method.
// It supports file types in allowedCompressExt.
// Due to the content-size requirement, it decompresses the downloaded file
// into the memory to get the size, then copies content to ResonpseWriter.
func handleDecompressGET(ctx context.Context, w http.ResponseWriter, r *http.Request, cacheServerURL string, reqID string) {
	objectName, err := parseURL(r.URL.Path)
	if err != nil {
		errStr := fmt.Sprintf("%s parseURL error: %s", reqID, err)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
		return
	}

	fileExt := filepath.Ext(objectName.path)
	newReader, ok := compressReaderMap[fileExt]
	if !ok {
		errStr := fmt.Sprintf("%s decompress does not support %s extension", reqID, fileExt)
		http.Error(w, errStr, http.StatusBadRequest)
		log.Printf(errStr)
		return
	}

	reqURL := fmt.Sprintf("%s/download/%s/%s", cacheServerURL, objectName.bucket, objectName.path)
	res, err := downloadURL(ctx, w, reqURL, reqID)
	if err != nil {
		return
	}
	defer res.Body.Close()

	dReader, err := newReader(res.Body)
	if err != nil {
		errStr := fmt.Sprintf("%s newReader error: %s", reqID, err)
		http.Error(w, errStr, http.StatusInternalServerError)
		log.Printf(errStr)
		return
	}
	defer dReader.Close()

	rMem, err := io.ReadAll(dReader)
	if err != nil {
		errStr := fmt.Sprintf("%s ReadAll failed after %v bytes: %s", reqID, len(rMem), err)
		http.Error(w, errStr, http.StatusInternalServerError)
		log.Printf(errStr)
		return
	}

	if err := decompressWrite(ctx, w, rMem); err != nil {
		log.Printf("%s decompressWrite failed: %s", reqID, err)
	}
}

// decompressWrite writes memory buffer to w Response
func decompressWrite(ctx context.Context, w http.ResponseWriter, mem []byte) error {
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(mem)))
	w.WriteHeader(http.StatusOK)

	n, err := w.Write(mem)
	if err != nil {
		return fmt.Errorf("write to client failed at byte %v: %w", n, err)
	}
	return nil
}
