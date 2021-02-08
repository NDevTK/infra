// Copyright 2020 The Chromium Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"go.chromium.org/luci/common/data/rand/mathrand"

	"infra/chromeperf/pinpoint/cli"
)

func main() {
	mathrand.SeedRandomly()
	os.Exit(cli.Main(cli.Param{
		DefaultServiceDomain: "pinpoint-dev.endpoints.chromeperf.cloud.goog",
		OIDCProviderURL:      "https://accounts.google.com",
	}, os.Args[1:]))
}
