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

package types

import (
	"infra/qscheduler/qslib/types/account"
	"infra/qscheduler/qslib/types/task"
	"infra/qscheduler/qslib/types/vector"
)

func NewConfig() *Config {
	return &Config{
		AccountConfigs: map[string]*account.Config{},
	}
}

func NewState() *State {
	return &State{
		Balances:     map[string]*vector.Vector{},
		RequestQueue: map[string]*task.Request{},
		Running:      []*task.Run{},
		Workers:      map[string]*Worker{},
	}
}
