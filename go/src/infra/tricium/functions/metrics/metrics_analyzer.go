// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	tricium "infra/tricium/api/v1"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/waigani/diffparser"
	"go.chromium.org/luci/common/data/stringset"
)

// enum contains all the data about a particular enum.
type enum struct {
	Name     string `xml:"name,attr"`
	Elements []struct {
		Value string `xml:"value,attr"`
		Label string `xml:"label,attr"`
	} `xml:"int"`
}

// enumFile contains all the data in an enums file.
type enumFile struct {
	Enums struct {
		EnumList []enum `xml:"enum"`
	} `xml:"enums"`
}

type diffsPerFile struct {
	addedLines   map[string][]int
	removedLines map[string][]int
}

func main() {
	inputDir := flag.String("input", "", "Path to directory with current versions of changed files")
	outputDir := flag.String("output", "", "Path to root of Tricium output")
	prevDir := flag.String("previous", "", "Path to directory with previous versions of changed files")
	// patchPath is an absolute path to the patch.
	patchPath := flag.String("patch", "", "Path to patch of changed files")
	// enumsPath is a relative path to the enums file.
	enumsPath := flag.String("enums", "tools/metrics/histograms/enums.xml", "Path to enums file")
	commitMessage := flag.String("message", "", "Commit message")
	flag.Parse()
	if *inputDir == "" || *outputDir == "" || *prevDir == "" || *patchPath == "" {
		log.Fatalf("Please specify non-empty values for the following flags: -input, -output, -previous, -patch, -message")
	}
	filePaths := flag.Args()
	filesChanged, err := getDiffsPerFile(filePaths, *patchPath)
	if err != nil {
		log.Panicf("Failed to get diffs per file: %v", err)
	}
	singletonEnums := getSingleElementEnums(filepath.Join(*inputDir, *enumsPath))

	results := &tricium.Data_Results{}
	allAddedHistograms := make(stringset.Set)
	allRemovedHistograms := make(stringset.Set)
	for _, filePath := range filePaths {
		inputPath := filepath.Join(*inputDir, filePath)
		f := openFileOrDie(inputPath)
		defer closeFileOrDie(f)
		if ext := filepath.Ext(filePath); ext == ".xml" {
			switch strings.TrimSuffix(filepath.Base(filePath), ext) {
			case "histograms":
				comments, addedHistograms, removedHistograms := analyzeHistogramFile(f, filePath, *prevDir, filesChanged, singletonEnums)
				results.Comments = append(results.Comments, comments...)
				allAddedHistograms = allAddedHistograms.Union(addedHistograms)
				allRemovedHistograms = allRemovedHistograms.Union(removedHistograms)
			case "histogram_suffixes_list":
				results.Comments = append(results.Comments, analyzeHistogramSuffixesFile(f, filePath, filesChanged)...)
			}
		} else if filepath.Ext(filePath) == ".json" {
			results.Comments = append(results.Comments, analyzeFieldTrialTestingConfig(f, filePath)...)
		}
	}

	globalObsoleteTagAdded := regexp.MustCompile(`OBSOLETE_HISTOGRAMS=(.+)`).Match([]byte(*commitMessage))
	removedHistograms := allRemovedHistograms.Difference(allAddedHistograms)
	// Check if all obsoletion messages and all removed histograms have their counterpart.
	results.Comments = append(results.Comments, analyzeCommitMessage(
		getObsoletedHistograms(*commitMessage), removedHistograms, globalObsoleteTagAdded)...)

	// Record all removed histograms in the CL.
	if removedHistograms.Len() > 0 {
		comment := &tricium.Data_Comment{
			Category: category + "/Removed",
			Message:  fmt.Sprintf(allRemovedHistogramInfo, strings.Join(removedHistograms.ToSlice(), ", ")),
		}
		results.Comments = append(results.Comments, comment)
	}

	// Write Tricium RESULTS data.
	path, err := tricium.WriteDataType(*outputDir, results)
	if err != nil {
		log.Panicf("Failed to write RESULTS data: %v. Did you specify an output directory with -output?", err)
	}
	log.Printf("Wrote RESULTS data to path %q.", path)
}

// getDiffsPerFile gets the added and removed line numbers for a particular file.
func getDiffsPerFile(filePaths []string, patchPath string) (*diffsPerFile, error) {
	patch, err := ioutil.ReadFile(patchPath)
	if err != nil {
		return &diffsPerFile{}, err
	}
	diff, err := diffparser.Parse(string(patch))
	if err != nil {
		return &diffsPerFile{}, err
	}
	diffInfo := &diffsPerFile{
		addedLines:   map[string][]int{},
		removedLines: map[string][]int{},
	}
	fileSet := stringset.NewFromSlice(filePaths...)
	for _, diffFile := range diff.Files {
		if diffFile.Mode == diffparser.DELETED || !fileSet.Has(diffFile.NewName) {
			continue
		}
		for _, hunk := range diffFile.Hunks {
			for _, line := range hunk.WholeRange.Lines {
				if line.Mode == diffparser.ADDED {
					diffInfo.addedLines[diffFile.NewName] = append(diffInfo.addedLines[diffFile.NewName], line.Number)
				} else if line.Mode == diffparser.REMOVED {
					diffInfo.removedLines[diffFile.NewName] = append(diffInfo.removedLines[diffFile.NewName], line.Number)
				}
			}
		}
	}
	return diffInfo, nil
}

func getSingleElementEnums(inputPath string) stringset.Set {
	singletonEnums := make(stringset.Set)
	f := openFileOrDie(inputPath)
	defer closeFileOrDie(f)
	enumBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Panicf("Failed to read enums into buffer: %v. Did you specify the enums file correctly with -enums?", err)
	}
	var enumFile enumFile
	if err := xml.Unmarshal(enumBytes, &enumFile); err != nil {
		log.Panicf("Failed to unmarshal enums: %v", err)
	}
	for _, enum := range enumFile.Enums.EnumList {
		if len(enum.Elements) == 1 {
			singletonEnums.Add(enum.Name)
		}
	}
	return singletonEnums
}

func openFileOrDie(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		log.Panicf("Failed to open file: %v, path: %s", err, path)
	}
	return f
}

func closeFileOrDie(f *os.File) {
	if err := f.Close(); err != nil {
		log.Panicf("Failed to close file: %v", err)
	}
}

func getObsoletedHistograms(commitMessage string) stringset.Set {
	re := regexp.MustCompile(`OBSOLETE_HISTOGRAM\[(.+?)\]`)
	histograms := re.FindAllStringSubmatch(commitMessage, -1)
	histogramsSet := make(stringset.Set)
	for _, match := range histograms {
		histogramsSet.Add(match[1])
	}
	return histogramsSet
}