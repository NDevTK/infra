// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	tricium "infra/tricium/api/v1"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	category        = "HistogramsXMLCheck"
	histogramEndTag = "</histogram>"
	ownerStartTag   = "<owner"

	oneOwnerError       = "Number of owners must be greater than 1"
	firstOwnerTeamError = "First owner must be individual, not a team"
)

// Histogram contains all info about histogram
type Histogram struct {
	Name          string
	Enum          string
	Units         string
	Expiration    string
	Owners        []string
	OwnersLineNum int
	Summary       string
}

func main() {
	inputDir := flag.String("input", "", "Path to root of Tricium input")
	outputDir := flag.String("output", "", "Path to root of Tricium output")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatalf("Unexpected argument.")
	}
	// Read Tricium input FILES data.
	input := &tricium.Data_Files{}
	if err := tricium.ReadDataType(*inputDir, input); err != nil {
		log.Fatalf("Failed to read FILES data: %v", err)
	}
	log.Printf("Read FILES data.")

	results := &tricium.Data_Results{}

	for _, file := range input.Files {
		if !file.IsBinary {
			if fileExt := filepath.Ext(file.Path); fileExt == ".xml" {
				log.Printf("ANALYZING File: %s", file.Path)
				p := filepath.Join(*inputDir, file.Path)
				f := openFileOrDie(p)
				results.Comments = analyzeFile(bufio.NewScanner(f), p)
				closeFileOrDie(f)
			}
		}
	}

	// Write Tricium RESULTS data.
	path, err := tricium.WriteDataType(*outputDir, results)
	if err != nil {
		log.Fatalf("Failed to write RESULTS data: %v", err)
	}
	log.Printf("Wrote RESULTS data to path %q.", path)
}

func analyzeFile(scanner *bufio.Scanner, path string) []*tricium.Data_Comment {
	var comments []*tricium.Data_Comment
	lineNum := 1
	currHistogram := Histogram{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, histogramEndTag) {
			if len(currHistogram.Owners) == 0 {
				currHistogram.OwnersLineNum = lineNum
			}
			if c := checkHistogram(path, currHistogram); c != nil {
				comments = append(comments, c...)
			}
			currHistogram = Histogram{}
		}
		if strings.HasPrefix(line, ownerStartTag) {
			if len(currHistogram.Owners) == 0 {
				currHistogram.OwnersLineNum = lineNum
			}
			currHistogram.Owners = append(currHistogram.Owners, XMLTagVal(line))
		}
		lineNum++
	}
	return comments
}

// XMLTagVal gets the value between XML tags
func XMLTagVal(line string) string {
	start := strings.Index(line, ">") + 1
	end := strings.LastIndex(line, "<")
	return line[start:end]
}

func checkHistogram(path string, histogram Histogram) []*tricium.Data_Comment {
	var comments []*tricium.Data_Comment
	if c := checkNumOwners(path, histogram); c != nil {
		comments = append(comments, c)
	}
	if c := checkNonTeamOwner(path, histogram); c != nil {
		comments = append(comments, c)
	}
	return comments
}

func checkNumOwners(path string, histogram Histogram) *tricium.Data_Comment {
	if len(histogram.Owners) <= 1 {
		log.Printf("ADDING Comment: %s", oneOwnerError)
		comment := &tricium.Data_Comment{
			Category:  fmt.Sprintf("%s/%s", category, "Owners"),
			Message:   oneOwnerError,
			Path:      path,
			StartLine: int32(histogram.OwnersLineNum),
			EndLine:   int32(histogram.OwnersLineNum),
		}
		return comment
	}
	return nil
}

func checkNonTeamOwner(path string, histogram Histogram) *tricium.Data_Comment {
	if len(histogram.Owners) > 0 && strings.Contains(histogram.Owners[0], "team") {
		log.Printf("ADDING Comment: %s", firstOwnerTeamError)
		comment := &tricium.Data_Comment{
			Category:  fmt.Sprintf("%s/%s", category, "Owners"),
			Message:   firstOwnerTeamError,
			Path:      path,
			StartLine: int32(histogram.OwnersLineNum),
			EndLine:   int32(histogram.OwnersLineNum),
		}
		return comment
	}
	return nil
}

func openFileOrDie(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file: %v, path: %s", err, path)
	}
	return f
}

func closeFileOrDie(f *os.File) {
	if err := f.Close(); err != nil {
		log.Fatalf("Failed to close file: %v", err)
	}
}
