// Copyright 2018 The LUCI Authors.
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
// package clients exports wrappers for client side bindings for API used by
// crosskylabadmin app. These interfaces provide a way to fake/stub out the API
// calls for tests.

// Package mock contains mocks for client interfaces from the parent package.
package mock

//go:generate go run runner/runner.go
