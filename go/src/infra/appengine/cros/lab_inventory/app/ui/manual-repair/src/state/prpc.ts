// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {PrpcClient} from '@chopsui/prpc-client';

export const prpcClient = new PrpcClient({
  // host: 'cros-lab-inventory.appspot.com',
  host: '0.0.0.0:8082',
  // insecure: false,
  insecure: Boolean(location.hostname === 'localhost'),
  fetchImpl: (url, options) => {
    if (options !== undefined) {
      options.credentials = 'include';
    }
    return fetch(url, options);
  },
});
