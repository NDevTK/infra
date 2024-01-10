// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build !windows
// +build !windows

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"debug/elf"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"infra/cros/internal/gs"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockClient struct {
	expectedResponses map[string]string
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	// if the request matches a string in the expected responses then generate a
	// response and return it. If it isn't apart of the map then return an error
	// and a nil response.
	if val, ok := c.expectedResponses[req.URL.String()]; ok {
		response := http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(val))}
		return &response, nil
	} else {
		return nil, fmt.Errorf("error: %s is not an expected url", req.URL.String())
	}
}

func initCrashConnectionMock(mockURL, mockKey string, responseMap map[string]string) crashConnectionInfo {
	return crashConnectionInfo{
		url:    mockURL,
		key:    mockKey,
		client: &mockClient{expectedResponses: responseMap},
	}
}

func buildFakeELFWithNote(buildID string) []byte {
	var buf bytes.Buffer

	decoded, err := hex.DecodeString(buildID)
	if err != nil {
		return nil
	}

	idLen := len(decoded)
	sectionHeaderOffset := 52 + 1 + 10 + 19 + 16 + idLen

	buf.Grow(sectionHeaderOffset + 2*40)

	h := elf.Header32{
		Ident:     [16]byte{0x7F, 'E', 'L', 'F', 0x1, 0x1, 0x1},
		Type:      1,
		Machine:   3,
		Version:   1,
		Shoff:     uint32(sectionHeaderOffset),
		Ehsize:    0x34,
		Shentsize: 0x28,
		Shnum:     3,
		Shstrndx:  1,
	}
	binary.Write(&buf, binary.LittleEndian, h)

	buf.Write([]byte{0})
	buf.Write([]byte(".shstrtab\x00"))
	buf.Write([]byte(".note.gnu.build-id\x00"))

	noteStart := buf.Len()
	binary.Write(&buf, binary.LittleEndian, []uint32{0x4, uint32(idLen), 3})
	buf.Write([]byte("GNU\x00"))
	buf.Write(decoded)
	noteEnd := buf.Len()

	binary.Write(&buf, binary.LittleEndian, elf.Section32{Name: 0})

	binary.Write(&buf, binary.LittleEndian, elf.Section32{
		Name:      1,
		Type:      uint32(elf.SHT_STRTAB),
		Off:       52,
		Size:      30,
		Addralign: 0x1,
	})

	binary.Write(&buf, binary.LittleEndian, elf.Section32{
		Name:      0xb,
		Type:      uint32(elf.SHT_NOTE),
		Flags:     uint32(elf.SHF_ALLOC),
		Addr:      uint32(noteStart),
		Off:       uint32(noteStart),
		Addralign: 0x4,
		Size:      uint32(noteEnd - noteStart),
	})

	return buf.Bytes()
}

// TestGetBuildId ensures that we can extract a build ID from a splitdebug
// file.
func TestGetBuildId(t *testing.T) {
	const expected = "BFCF6FA6CCBDEF00501810DE869C8A2F40ABC321"

	testDir, err := ioutil.TempDir("", "getBuildIdTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	mockPath := filepath.Join(testDir, "fake.debug")
	err = ioutil.WriteFile(mockPath, buildFakeELFWithNote(expected), 0644)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	actual, err := getBuildId(mockPath)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	if actual != expected {
		t.Error(actual + " doesn't match expected ID " + expected)
	}
}

// TestDownloadZippedSymbols ensures that we are fetching from the correct service and
// handling the response appropriately.
func TestDownloadZippedSymbols(t *testing.T) {
	gsPath := "gs://some-debug-symbol/degbug.tgz"

	expectedDownloads := map[string][]byte{
		gsPath: []byte("hello world"),
	}
	fakeClient := &gs.FakeClient{
		T:                 t,
		ExpectedDownloads: expectedDownloads,
	}

	tarballDir, err := ioutil.TempDir("", "tarball")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(tarballDir)

	tgzPath, err := downloadZippedSymbols(fakeClient, gsPath, tarballDir)

	if err != nil {
		t.Error("error: " + err.Error())
	}

	if _, err := os.Stat(tgzPath); os.IsNotExist(err) {
		t.Error("error: " + err.Error())
	}
}

// TestUnzipTgz confirms that we can properly unzip a given tgz file.
func TestUnzipSymbols(t *testing.T) {

	targetString := "gzip test"

	// Create temp dir to work in.
	testDir, err := ioutil.TempDir("", "tarballTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	// Generate file information.
	outputFilePath := filepath.Join(testDir, "test.tar")
	inputFilePath := filepath.Join(testDir, "test.tgz")
	inputFile, err := os.Create(inputFilePath)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer inputFile.Close()

	// Create a mock .tgz to test unzipping.
	zipWriter := gzip.NewWriter(inputFile)
	zipWriter.Name = "test.tgz"
	zipWriter.Comment = "hello world"

	_, err = zipWriter.Write([]byte(targetString))
	if err != nil {
		t.Error("error: " + err.Error())
	}

	err = zipWriter.Close()
	if err != nil {
		t.Error("error: " + err.Error())
	}

	err = unzipSymbols(inputFilePath, outputFilePath)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	// Check if the output file was created.
	if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
		t.Error("error: " + err.Error())
	}

	outputFile, err := os.Open(outputFilePath)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer outputFile.Close()

	output, err := io.ReadAll(outputFile)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	// Verify contents unzipped correctly.
	if string(output) != targetString {
		t.Errorf("error: expected %s got %s", targetString, string(output))
	}

}

// TestUnpackTarball confirms that we can properly unpack a given tarball and
// return filepaths to it's contents. Basic testing pulled from
// https://pkg.go.dev/archive/tar#pkg-overview.
func TestUnpackTarball(t *testing.T) {
	// Create working directory and tarball.
	testDir, err := ioutil.TempDir("", "tarballTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	debugSymbolsDir, err := ioutil.TempDir(testDir, "symbols")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	// Generate file information.
	tarPath := filepath.Join(testDir, "test.tar")
	inputFile, err := os.Create(tarPath)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	tarWriter := tar.NewWriter(inputFile)
	// Struct for file info
	type file struct {
		name, body string
		modeType   fs.FileMode
	}

	// Create an array holding some basic info to build headers. Contains regular
	// files and directories.
	files := []file{
		{"/test1.so.sym", "MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F40 blkid", 0600},
		{"./test2.so.sym", "MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F41 blkid", 0600},
		{"b/c", "", fs.ModeDir},
		// Different Debug ID as other test1.so.sym so should be included.
		{"b/c/test1.so.sym", "MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F42 blkid", 0600},
		{"../test3.so.sym", "MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F43 blkid", 0600},
		{"a/b/c/d/", "", fs.ModeDir},
		{"./test4.so.sym", "MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F44 blkid", 0600},
		// Same Debug ID as previous file so should not be included.
		{"a/b/c/d/test4.so.sym", "MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F44 blkid", 0600},
		{"a/shouldntadd.txt", "not a symbol file", 0600},
	}

	// List of files we expect to see return after the test call.
	expectedSymbolFiles := map[string]bool{
		debugSymbolsDir + "/test1.so.sym":   false,
		debugSymbolsDir + "/test1.so.sym-1": false,
		debugSymbolsDir + "/test2.so.sym":   false,
		debugSymbolsDir + "/test3.so.sym":   false,
		debugSymbolsDir + "/test4.so.sym":   false,
	}

	// Write the mock files to the tarball.
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Mode: int64(file.modeType),
			Size: int64(len(file.body)),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			t.Error("error: " + err.Error())
		}

		if file.modeType == 0600 {
			if _, err := tarWriter.Write([]byte(file.body)); err != nil {
				t.Error("error: " + err.Error())
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Error("error: " + err.Error())
	}

	// Call the Function
	symbolPaths, err := unpackTarball(tarPath, debugSymbolsDir)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	if symbolPaths == nil || len(symbolPaths) <= 0 {
		t.Error("error: Empty list of paths returned")
	}
	// Verify that we received a list pointing to all the expected files and no
	// others.
	for _, path := range symbolPaths {
		if val, ok := expectedSymbolFiles[path]; ok {
			if val {
				t.Error("error: symbol file appeared multiple times in function return")
			}
			expectedSymbolFiles[path] = true
		} else {
			t.Errorf("error: unexpected symbol file returned %s", path)
		}
	}
}

// TestUnpackSplitdebugTarballs confirms that we can properly unpack a given tarball and
// return filepaths to it's contents. Basic testing pulled from
// https://pkg.go.dev/archive/tar#pkg-overview.
func TestUnpackSplitdebugTarballs(t *testing.T) {
	// Create working directory and tarball.
	testDir, err := ioutil.TempDir("", "splitdebugTarballTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	debugSymbolsDir, err := ioutil.TempDir(testDir, "symbols")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	// Struct for symbolFile info
	type symbolFile struct {
		name, body string
		modeType   fs.FileMode
	}

	// The debug Id is a byte swapped and truncated version of the build Id.
	const buildId1 = "F4F6FA6CCBDEF455039C8DE869C8A2F4AB"
	const debugId1 = "6CFAF6F4F455CBDE039C8DE869C8A2F40"
	const buildId1Dup = "F4F6FA6CCBDEF455039C8DE869C8A2F6AB"
	const debugId1Dup = "6CFAF6F4F455CBDE039C8DE869C8A2F60"
	const buildId2 = "F4F6FA6CCBDEF455039C8DE869C8A2F5AC"
	const debugId2 = "6CFAF6F4F455CBDE039C8DE869C8A2F50"
	const buildId3 = "F4F6FA6CCBDEF455039C8DE869C8A2F7AD"
	const debugId3 = "6CFAF6F4F455CBDE039C8DE869C8A2F70"
	const buildId4 = "F4F6FA6CCBDEF455039C8DE869C8A2F8AE"
	const debugId4 = "6CFAF6F4F455CBDE039C8DE869C8A2F80"

	// Create an array holding some basic info to build headers. Contains regular
	// files and directories.
	symbolFiles := []symbolFile{
		{"/test1.so.sym", "MODULE Linux arm " + debugId1 + " test1.so\nINFO CODE_ID " + buildId1, 0600},
		{"./test2.so.sym", "MODULE Linux arm " + debugId2 + " test2.so\nINFO CODE_ID " + buildId2, 0600},
		{"b/c", "", fs.ModeDir},
		// Different Debug ID as other test1.so.sym so should be included.
		{"b/c/test1.so.sym", "MODULE Linux arm " + debugId1Dup + " test1.so\nINFO CODE_ID " + buildId1Dup, 0600},
		{"../test3.so.sym", "MODULE Linux arm " + debugId3 + " test3.so\nINFO CODE_ID " + buildId3, 0600},
		{"a/b/c/d/", "", fs.ModeDir},
		{"./test4.so.sym", "MODULE Linux arm " + debugId4 + " test4.so\nINFO CODE_ID " + buildId4, 0600},
		// Same Debug ID as previous file so should not be included.
		{"a/b/c/d/test4.so.sym", "MODULE Linux arm " + debugId4 + " test4.so\nINFO CODE_ID " + buildId4, 0600},
		{"a/shouldntadd.txt", "not a symbol file", 0600},
	}

	// List of files we expect to see return after the test call.
	expectedSymbolFiles := map[string]bool{
		debugSymbolsDir + "/test1.so.sym":   false,
		debugSymbolsDir + "/test1.so.sym-1": false,
		debugSymbolsDir + "/test2.so.sym":   false,
		debugSymbolsDir + "/test3.so.sym":   false,
		debugSymbolsDir + "/test4.so.sym":   false,
	}

	// Struct for file info
	type splitdebugFile struct {
		name     string
		body     []byte
		modeType fs.FileMode
	}
	// Create an array holding some basic info to build headers. Contains regular
	// files and directories.
	splitdebugFiles := []splitdebugFile{
		{"/test1.so.debug", buildFakeELFWithNote(buildId1), 0600},
		{"./test2.so.debug", buildFakeELFWithNote(buildId2), 0600},
		{"b/c", nil, fs.ModeDir},
		// Different Debug ID as other test1.so.debug so should be included.
		{"b/c/test1.so.debug", buildFakeELFWithNote(buildId1Dup), 0600},
		// Different name as breakpad symbol (test3.so.sym) should be included
		{"../test3.so.4.0.5.debug", buildFakeELFWithNote(buildId3), 0600},
		{"a/b/c/d/", nil, fs.ModeDir},
		{"./test4.so.debug", buildFakeELFWithNote(buildId4), 0600},
		// Same Debug ID as previous file so should not be included.
		{"a/b/c/d/test4.so.debug", buildFakeELFWithNote(buildId4), 0600},
		{"a/shouldntadd.txt", []byte("not a symbol file\x00"), 0600},
	}

	// List of files we expect to see return after the test call.
	expectedSplitdebugFiles := map[string]bool{
		debugSymbolsDir + "/" + debugId1 + "/test1.so.debug":       false,
		debugSymbolsDir + "/" + debugId1Dup + "/test1.so.debug":    false,
		debugSymbolsDir + "/" + debugId2 + "/test2.so.debug":       false,
		debugSymbolsDir + "/" + debugId3 + "/test3.so.4.0.5.debug": false,
		debugSymbolsDir + "/" + debugId4 + "/test4.so.debug":       false,
	}

	// Generate file information.
	splitdebugTarPath := filepath.Join(testDir, "test-debug.tar")
	splitdebugInputFile, err := os.Create(splitdebugTarPath)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	splitdebugTarWriter := tar.NewWriter(splitdebugInputFile)

	// Write the mock files to the tarball.
	for _, file := range splitdebugFiles {
		hdr := &tar.Header{
			Name: file.name,
			Mode: int64(file.modeType),
			Size: int64(len(file.body)),
		}
		if err := splitdebugTarWriter.WriteHeader(hdr); err != nil {
			t.Error("error: " + err.Error())
		}

		if file.modeType == 0600 {
			if _, err := splitdebugTarWriter.Write(file.body); err != nil {
				t.Error("error: " + err.Error())
			}
		}
	}
	if err := splitdebugTarWriter.Close(); err != nil {
		t.Error("error: " + err.Error())
	}

	symbolTarPath := filepath.Join(testDir, "test-symbols.tar")
	symbolInputFile, err := os.Create(symbolTarPath)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	symbolTarWriter := tar.NewWriter(symbolInputFile)

	// Write the mock files to the tarball.
	for _, file := range symbolFiles {
		hdr := &tar.Header{
			Name: file.name,
			Mode: int64(file.modeType),
			Size: int64(len(file.body)),
		}
		if err := symbolTarWriter.WriteHeader(hdr); err != nil {
			t.Error("error: " + err.Error())
		}

		if file.modeType == 0600 {
			if _, err := symbolTarWriter.Write([]byte(file.body)); err != nil {
				t.Error("error: " + err.Error())
			}
		}
	}
	if err := symbolTarWriter.Close(); err != nil {
		t.Error("error: " + err.Error())
	}

	// Call the function under test.
	splitdebugPaths, breakpadPaths, err := unpackSplitdebugTarballs(splitdebugTarPath, symbolTarPath, debugSymbolsDir)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	if splitdebugPaths == nil || len(splitdebugPaths) <= 0 {
		t.Error("error: Empty list of splitdebug paths returned")
	}
	if breakpadPaths == nil || len(breakpadPaths) <= 0 {
		t.Error("error: Empty list of breakpad paths returned")
	}
	// Verify that we received a list pointing to all the expected files and no
	// others.
	for _, path := range splitdebugPaths {
		if val, ok := expectedSplitdebugFiles[path]; ok {
			if val {
				t.Errorf("error: splitdebug file %q appeared multiple times in function return", path)
			}
			expectedSplitdebugFiles[path] = true
		} else {
			t.Errorf("error: unexpected splitdebug file returned %q", path)
		}
	}
	for path, ok := range expectedSplitdebugFiles {
		if !ok {
			t.Errorf("error: splitdebug file %q not found", path)
		}
	}
	// Verify that we received a list pointing to all the expected files and no
	// others.
	for _, path := range breakpadPaths {
		if val, ok := expectedSymbolFiles[path]; ok {
			if val {
				t.Errorf("error: breakpad file %q appeared multiple times in function return", path)
			}
			expectedSymbolFiles[path] = true
		} else {
			t.Errorf("error: unexpected breakpad file returned %q", path)
		}
	}
	for path, ok := range expectedSymbolFiles {
		if !ok {
			t.Errorf("error: breakpad file %q not found", path)
		}
	}
}

// TestUnpackKernelTarball confirms that we can properly unpack a given tarball and
// return filepath to vmlinux.debug. Basic testing pulled from
// https://pkg.go.dev/archive/tar#pkg-overview.
func TestUnpackKernelTarball(t *testing.T) {
	// Create working directory and tarball.
	testDir, err := ioutil.TempDir("", "tarballTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	debugSymbolsDir, err := ioutil.TempDir(testDir, "symbols")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	// Generate file information.
	tarPath := filepath.Join(testDir, "test.tar")
	inputFile, err := os.Create(tarPath)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	tarWriter := tar.NewWriter(inputFile)
	// Struct for file info
	type file struct {
		name, body string
		modeType   fs.FileMode
	}

	// Create an array holding some basic info to build headers. Contains regular
	// files and directories.
	files := []file{
		{"vmlinuz", "fake-vmlinuz-content", 0600},
		{"vmlinux.debug", "fake-vmlinux-content", 0600},
	}

	// Expected filename after test called.
	expectedSymbolFile := debugSymbolsDir + "/vmlinux.debug"

	// Write the mock files to the tarball.
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Mode: int64(file.modeType),
			Size: int64(len(file.body)),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			t.Error("error: " + err.Error())
		}

		if file.modeType == 0600 {
			if _, err := tarWriter.Write([]byte(file.body)); err != nil {
				t.Error("error: " + err.Error())
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Error("error: " + err.Error())
	}

	// Call the Function
	symbolPath, err := unpackKernelTarball(tarPath, debugSymbolsDir)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	if symbolPath != expectedSymbolFile {
		t.Errorf("error: unexpected symbol file returned %s", symbolPath)
	}
}

// TestGenerateConfigs validates that proper task configs are generated when a
// list of filepaths are given.
func TestGenerateConfigs(t *testing.T) {
	// Init the mock files and verifying structures.
	expectedTasks := map[taskConfig]bool{}
	type responseInfo struct {
		filename string
		symbol   string
		status   string
		// Local path to write the file to. Used for dupe debug symbols
		// (mimicking behavior in unpackTarball)
		localPath string
	}
	mockResponses := []*responseInfo{
		{
			filename: "test1.so.sym",
			symbol:   "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:   "FOUND",
		}, {
			filename:  "test1.so.sym",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F41",
			status:    "MISSING",
			localPath: "test1.so.sym-1",
		}, {
			filename: "test2.so.sym",
			symbol:   "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:   "FOUND",
		}, {
			filename: "test3.so.sym",
			symbol:   "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:   "Missing",
		},
		{
			filename: "test4.so.sym",
			symbol:   "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:   "MISSING",
		},
		{
			filename: "test5.so.sym",
			symbol:   "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:   "STATUS_UNSPECIFIED",
		},
		{
			filename: "test6.so.sym",
			symbol:   "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:   "STATUS_UNSPECIFIED",
		},
	}

	// Make the expected request body.
	responseBody := filterResponseBody{Pairs: []filterResponseStatusPair{}}

	// Test for all 3 cases found in http://google3/net/crash/symbolcollector/symbol_collector.proto?l=19
	for _, response := range mockResponses {
		symbol := filterSymbolFileInfo{response.filename, response.symbol}
		responseBody.Pairs = append(responseBody.Pairs, filterResponseStatusPair{SymbolId: symbol, Status: response.status})
	}
	mockResponseBody, err := json.Marshal(responseBody)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	// Mock the symbol files locally.
	testDir, err := ioutil.TempDir("", "configGenTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	mockPaths := []string{}
	for _, response := range mockResponses {
		localPath := response.filename
		if response.localPath != "" {
			localPath = response.localPath
		}
		mockPath := filepath.Join(testDir, localPath)
		err = ioutil.WriteFile(mockPath, []byte(fmt.Sprintf("MODULE Linux arm %s %s", response.symbol, filepath.Base(response.filename))), 0644)
		if err != nil {
			t.Error("error: " + err.Error())
		}
		task := taskConfig{mockPath, "BREAKPAD", response.filename, response.symbol, false, false}

		if response.status != "FOUND" {
			expectedTasks[task] = false
		}
		mockPaths = append(mockPaths, mockPath)
	}

	// Init global variables and
	mockCrash := initCrashConnectionMock("google.com", "1234", map[string]string{"google.com/symbols:checkStatuses?key=1234": string(mockResponseBody)})

	tasks, err := generateConfigs(context.Background(), mockPaths, 0, false, mockCrash)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	// Check that returns aren't nil.
	if tasks == nil {
		t.Error("error: recieved tasks when nil was expected")
	}

	// Verify that we received a list pointing to all the expected files and no
	// others.
	for _, task := range tasks {
		if val, ok := expectedTasks[task]; ok {
			if val {
				t.Errorf("error: task %v appeared multiple times in function return", task)
			}
			expectedTasks[task] = true
		} else {
			t.Errorf("error: unexpected task returned %+v", task)
		}
	}

	for task, value := range expectedTasks {
		if value == false {
			t.Errorf("error: task for file %s never seen", task.debugFile)
		}
	}
}

// TestGenerateSplitdebugConfigs validates that proper task configs are generated when a
// list of filepaths are given.
func TestGenerateSplitdebugConfigs(t *testing.T) {
	// Init the mock files and verifying structures.
	expectedTasks := map[taskConfig]bool{}
	type responseInfo struct {
		filename  string
		symbol    string
		status    string
		localPath string
	}
	mockResponses := []*responseInfo{
		{
			filename:  "test1.so",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:    "FOUND",
			localPath: "test1.so.debug",
		}, {
			filename:  "test2.so",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:    "FOUND",
			localPath: "test2.so.debug",
		}, {
			filename:  "test3.so",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:    "Missing",
			localPath: "test3.so.debug",
		},
		{
			filename:  "test4.so",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:    "MISSING",
			localPath: "test4.so.debug",
		},
		{
			filename:  "test5.so",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:    "STATUS_UNSPECIFIED",
			localPath: "test5.so.debug",
		},
		{
			filename:  "test6.so",
			symbol:    "F4F6FA6CCBDEF455039C8DE869C8A2F40",
			status:    "STATUS_UNSPECIFIED",
			localPath: "test6.so.debug",
		},
	}

	// Make the expected request body.
	responseBody := filterResponseBody{Pairs: []filterResponseStatusPair{}}

	// Test for all 3 cases found in http://google3/net/crash/symbolcollector/symbol_collector.proto?l=19
	for _, response := range mockResponses {
		symbol := filterSymbolFileInfo{response.filename, response.symbol}
		responseBody.Pairs = append(responseBody.Pairs, filterResponseStatusPair{SymbolId: symbol, Status: response.status})
	}
	mockResponseBody, err := json.Marshal(responseBody)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	// Mock the symbol files locally.
	testDir, err := ioutil.TempDir("", "configGenTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	//defer os.RemoveAll(testDir)

	mockPaths := []string{}
	for _, response := range mockResponses {
		localPath := response.localPath
		localDir := filepath.Join(testDir, response.symbol)
		err := os.Mkdir(localDir, 0750)
		if err != nil && !os.IsExist(err) {
			t.Error("error: " + err.Error())
		}
		mockPath := filepath.Join(localDir, localPath)
		err = ioutil.WriteFile(mockPath, buildFakeELFWithNote(response.symbol), 0644)
		if err != nil {
			t.Error("error: " + err.Error())
		}
		task := taskConfig{mockPath, "DEBUG_ONLY", response.filename, response.symbol, false, false}

		if response.status != "FOUND" {
			expectedTasks[task] = false
		}
		mockPaths = append(mockPaths, mockPath)
	}

	// Init global variables and
	mockCrash := initCrashConnectionMock("google.com", "1234", map[string]string{"google.com/symbols:checkStatuses?key=1234": string(mockResponseBody)})

	tasks, err := generateSplitdebugConfigs(context.Background(), mockPaths, 0, false, mockCrash)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	// Check that returns aren't nil.
	if tasks == nil {
		t.Error("error: recieved tasks when nil was expected")
	}

	// Verify that we received a list pointing to all the expected files and no
	// others.
	for _, task := range tasks {
		if val, ok := expectedTasks[task]; ok {
			if val {
				t.Errorf("error: task %v appeared multiple times in function return", task)
			}
			expectedTasks[task] = true
		} else {
			t.Errorf("error: unexpected task returned %+v", task)
		}
	}

	for task, value := range expectedTasks {
		if value == false {
			t.Errorf("error: task for file %s never seen", task.debugFile)
		}
	}
}

// TestGenerateKernelConfigs validates that proper task configs are generated when a
// list of filepaths are given.
func TestGenerateKernelConfigs(t *testing.T) {
	// Init the mock files and verifying structures.
	type responseInfo struct {
		filename string
		symbol   string
		status   string
	}
	mockResponse := responseInfo{
		filename: "vmlinux.debug",
		symbol:   "jacuzzi-R117-15544.0.0",
		status:   "MISSING",
	}

	// Make the expected request body
	responseBody := filterResponseBody{Pairs: []filterResponseStatusPair{}}
	symbol := filterSymbolFileInfo{mockResponse.filename, mockResponse.symbol}
	responseBody.Pairs = append(responseBody.Pairs, filterResponseStatusPair{SymbolId: symbol, Status: mockResponse.status})

	mockResponseBody, err := json.Marshal(responseBody)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	// Mock the symbol files locally.
	testDir, err := ioutil.TempDir("", "configGenTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	// Init mock.
	mockCrash := initCrashConnectionMock("google.com", "1234", map[string]string{"google.com/symbols:checkStatuses?key=1234": string(mockResponseBody)})

	mockPath := filepath.Join(testDir, mockResponse.filename)
	err = ioutil.WriteFile(mockPath, []byte("fake-vmlinux-content"), 0644)
	if err != nil {
		t.Error("error: " + err.Error())
	}

	tasks, err := generateKernelConfigs(context.Background(), mockPath, mockResponse.symbol, 0, false, mockCrash)
	if err != nil {
		t.Error("error: " + err.Error())
	}
	// Check that returns aren't nil.
	if tasks == nil {
		t.Error("error: expect task to be not nil")
	}
}

// TestUploadSymbols affirms that the worker design and retry model are valid.
func TestUploadSymbols(t *testing.T) {
	// Create tasks and expected returns.
	tasks := []taskConfig{
		{"", "BREAKPAD", "test1.so.sym", "", false, false},
		{"", "BREAKPAD", "test2.so.sym", "", false, false},
		{"", "BREAKPAD", "test3.so.sym", "", false, false},
		{"", "BREAKPAD", "test4.so.sym", "", false, false},
	}

	// Mock the symbol files locally.
	testDir, err := ioutil.TempDir("", "uploadSymbolsTest")
	if err != nil {
		t.Error("error: " + err.Error())
	}
	defer os.RemoveAll(testDir)

	// Write mock files locally.
	for index, task := range tasks {
		mockPath := filepath.Join(testDir, task.debugFile)
		err = ioutil.WriteFile(mockPath, []byte("MODULE Linux arm F4F6FA6CCBDEF455039C8DE869C8A2F40 blkid"), 0644)
		if err != nil {
			t.Error("error: " + err.Error())
		}
		tasks[index].symbolPath = mockPath
	}

	// unique URL and key
	mockURLKeyPair := crashUploadInformation{UploadUrl: "crashupload.com", UploadKey: "abc"}
	mockCompleteResponse := crashSubmitResposne{Result: "OK"}
	mockURLKeyPairJSON, err := json.Marshal(mockURLKeyPair)
	if err != nil {
		t.Error("error: could not marshal response")
	}
	mockCompleteResponseJSON, err := json.Marshal(mockCompleteResponse)
	if err != nil {
		t.Error("error: could not marshal response")
	}

	expectedResponses := map[string]string{
		"google.com/uploads:create?key=1234":       string(mockURLKeyPairJSON),
		"crashupload.com":                          "",
		"google.com/uploads/abc:complete?key=1234": string(mockCompleteResponseJSON),
	}

	crashMock := initCrashConnectionMock("google.com", "1234", expectedResponses)
	retcode, err := uploadSymbols(tasks, 64, 2, false, crashMock)

	if err != nil {
		t.Error("error: " + err.Error())
	}

	if retcode != 0 {
		t.Errorf("error: recieved non-zero retcode %d", retcode)
	}
}

// TestCleanErrorMessage tests to make sure we won't accidentally expose the api
// key.
func TestCleanErrorMessage(t *testing.T) {
	mockKey := "abcde12345!@#$"
	mockErr := fmt.Errorf("could not connect to www.google.com/secret?key=%s&v1=1&v2=test", mockKey)

	cleanedErrMessage := cleanErrorMessage(mockErr.Error(), mockKey)

	if strings.Contains(cleanedErrMessage, mockKey) {
		t.Error("error: secret key found in error response")
	}

	if !strings.Contains(cleanedErrMessage, "-HIDDEN-KEY-") {
		t.Error("error: replaced key not found")
	}

	mockErr = fmt.Errorf("nil http repsonse was recieved")
	cleanedErrMessage = cleanErrorMessage(mockErr.Error(), mockKey)

	if strings.Contains(cleanedErrMessage, "-HIDDEN-KEY-") {
		t.Error("error: non-existant key was fouind in error")
	}
}

// TestGenerateKernelDebugId tests parsing kernel debugId.
func TestGenerateKernelDebugId(t *testing.T) {
	mockPath := "gs://chromeos-image-archive/brya-release/R117-15544.0.0/vmlinuz.tar.xz"
	_, err := generateKernelDebugId(mockPath)
	if err != nil {
		t.Error("error: " + err.Error())
	}
}
