// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmdlib

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// PromptFunc asks for a string response from the user for a given request string.
type PromptFunc func(string) string

// CLIPrompt returns a PromptFunc to prompt user on CLI.
//
// In case of erroneous input from user, the returned PromptFunc prompts the
// user again.
func CLIPrompt(w io.Writer, r io.Reader) PromptFunc {
	return func(reason string) string {
		if err := prompt(w, reason); err != nil {
			return emptyResponse
		}
		for {
			res, err := getPromptResponse(r)
			if err != nil {
				return emptyResponse
			}
			switch res {
			case "":
				if err := reprompt(w, res); err != nil {
					return emptyResponse
				}
			default:
				return res
			}
		}
	}
}

// emptyResponse is the empty response for string user input.
const emptyResponse = ""

func prompt(w io.Writer, reason string) error {
	b := bufio.NewWriter(w)
	fmt.Fprintf(b, "%s\t", reason)
	return b.Flush()
}

func getPromptResponse(r io.Reader) (string, error) {
	b := bufio.NewReader(r)
	i, err := b.ReadString('\n')
	if err != nil {
		return "", errors.Annotate(err, "get prompt response").Err()
	}
	return strings.Trim(strings.ToLower(i), " \n\t"), nil
}

func reprompt(w io.Writer, response string) error {
	b := bufio.NewWriter(w)
	fmt.Fprintf(b, "\n\tInvalid response %s. Please re-enter: ", response)
	return b.Flush()
}
