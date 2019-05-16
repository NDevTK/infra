// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

export class UserInputError extends Error {
  get name() {
    return 'UserInputError';
  }
}
