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

// PromptFunc obtains consent from the user for the given request string.
//
// This function is used to provide the user some context through the provided
// string and then obtain a yes/no answer from the user.
type PromptFunc func(string) string

// CLIPrompt returns a PromptFunc to prompt user on CLI.
//
// In case of erroneous input from user, the returned PromptFunc prompts the
// user again.
// defaultResponse is returned on empty response from the user.
// In case of other system errors, the returned promptFunc returns false.
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
	fmt.Fprintf(b, "%s\n", reason)
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
