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

package cli

import "context"

type jwtCredentials struct {
	jwt []byte
}

func (t *jwtCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"Authorization": "Bearer " + string(t.jwt)}, nil
}

func (*jwtCredentials) RequireTransportSecurity() bool { return true }

func newJWTCredentials(jwt []byte) *jwtCredentials { return &jwtCredentials{jwt: jwt} }
