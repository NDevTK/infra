// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package version

import (
	"regexp"
	"strconv"
)

// numberRegex gets all the numbers in a string
var numberRegex = regexp.MustCompile(`[0-9]+`)

// GEQ Compare the given two version strings and return true if a >= b
// The comparison is done by comparing all the numbers in the string.
//
// Note:
//  1. This extracts all the numbers in the strings and compares them
//     in order. For example comparing shivas/11.2.111-rc4 with shivas/11.3.36
//     is done by comparing [11, 2, 111, 4] with [11, 3, 36], first 11 is
//     compared to 11, followed by comparing 2 to 3. It stops on the first diff
//     in version numbers.
//  2. There might be certain drawbacks for this implementation. Using dates
//     as version strings might not work. Ex 1.2.2022126 and 1.2.20221212.
//     It will work if the dates and month are always written as two digits
//     Ex: 1.2.20221212 and 1.2.20220606
//  3. The current implementation is transitive,(If A >= B and B >= C then,
//     A >= C). The proof for this simple. Suppose we have a set of strings
//     A = 11.2.111-rc-4 = [11, 2, 111, 4]
//     B = 11.3.36 = [11, 3, 36]
//     You can write A and B as a number in base 112 [Max of the numbers + 1]
//     Now, a. GEQ will still give the same result
//     b. As A and B are numbers transitive property holds
//  4. This was done to compare the versions of user agents that connect to UFS
//     and works well for shivas and pRPC clients. Shivas uses `shivas/x.x.x`
//     format and prpc uses `pRPC client x.x` format. This seems to work well
//     for both of them. If you plan to use it elsewhere, be sure to check
//     the versioning format and decide for yourself
//  5. We briefly considered recognizing the different formats and implementing
//     different comparison for each of them. We decided against it as it would
//     be too complicated to maintain and wouldn't be flexible enough for our
//     use case.
func GEQ(a, b string) bool {
	aVersions := numberRegex.FindAllString(a, -1)
	bVersions := numberRegex.FindAllString(b, -1)
	for idx, bVersion := range bVersions {
		if len(aVersions) > idx {
			aNum, _ := strconv.Atoi(aVersions[idx])
			bNum, _ := strconv.Atoi(bVersion)
			if aNum < bNum {
				// b is bigger
				return false
			}
			if aNum > bNum {
				// a is bigger
				return true
			}
		} else {
			// b has an extra version number
			return false
		}
	}
	return true
}
