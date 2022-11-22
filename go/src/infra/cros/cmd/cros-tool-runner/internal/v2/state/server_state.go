// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

// serverState tracks state of a CTRv2 server instance.
type serverState struct {
	TemplateRequest TemplateRequestRecorder
}

// ServerState is the singleton server state of the current CTRv2 instance.
// TODO(mingkong) move ownership state here
var ServerState = serverState{
	TemplateRequest: newTemplateRequestState(),
}
