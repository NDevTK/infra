// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {store} from 'elements/reducers/base.js';
import * as sitewide from 'elements/reducers/sitewide.js';

// How long should we wait until asking the server status again.
const SERVER_STATUS_DELAY_MS = 20 * 60 * 1000; // 20 minutes

export function getServerStatusCron() {
  store.dispatch(sitewide.getServerStatus());
  setTimeout(getServerStatusCron.bind(null), SERVER_STATUS_DELAY_MS);
}
