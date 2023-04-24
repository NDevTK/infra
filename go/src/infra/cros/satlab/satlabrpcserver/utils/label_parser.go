// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import "regexp"

// LabelParser a parser for parsing the label string.
//
// We get some blob name from gcs bucket, and then we can use this parser
// to extract the information.
type LabelParser struct {
	// a regexp object for parsing the milestone
	MilestonePattern *regexp.Regexp

	// a regexp object for parsing the board and model
	BoardAndModelPattern *regexp.Regexp
}

// NewLabelParser create a `LabelParser`. it returns error if any of regexp can't compile.
func NewLabelParser() (*LabelParser, error) {
	milestonePattern, err := regexp.Compile("\\Amilestones/(?P<Milestone>\\d+)$")
	if err != nil {
		return nil, err
	}
	boardAndModelPattern, err := regexp.Compile("^buildTargets/(?P<Board>\\w+)/models/(?P<Model>\\w+)$")
	if err != nil {
		return nil, err
	}

	return &LabelParser{
		MilestonePattern:     milestonePattern,
		BoardAndModelPattern: boardAndModelPattern,
	}, nil
}

// ExtractBoardAndModelFrom extract board and model information from the given string.
//
// string s the string we want to get the information from.
func (l *LabelParser) ExtractBoardAndModelFrom(s string) (BoardAndModelPair, error) {
	if !l.BoardAndModelPattern.MatchString(s) {
		return BoardAndModelPair{}, NotMatch
	}

	matches := l.BoardAndModelPattern.FindStringSubmatch(s)
	boardIndex := l.BoardAndModelPattern.SubexpIndex("Board")
	modelIndex := l.BoardAndModelPattern.SubexpIndex("Model")

	return BoardAndModelPair{Board: matches[boardIndex], Model: matches[modelIndex]}, nil
}

// ExtractMilestone extract the milestone information from the given string.
//
// string s the string we want to get the information from.
func (l *LabelParser) ExtractMilestone(s string) (string, error) {
	if !l.MilestonePattern.MatchString(s) {
		return "", NotMatch
	}

	matches := l.MilestonePattern.FindStringSubmatch(s)
	index := l.MilestonePattern.SubexpIndex("Milestone")

	return matches[index], nil
}
