// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Contains data models shared between multiple services.

export interface ClusterId {
    algorithm: string;
    id: string;
}

export interface AssociatedBug {
    system: string;
    id: string;
    linkText: string;
    url: string;
}
