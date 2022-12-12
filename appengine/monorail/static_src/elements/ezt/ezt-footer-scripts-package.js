// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This file bundles together scripts to be loaded through the legacy
// EZT footer.

import 'monitoring/client-logger.js';
import 'monitoring/track-copy.js';

// Allow EZT pages to import AutoRefreshPrpcClient.
import AutoRefreshPrpcClient from 'prpc.js';

window.AutoRefreshPrpcClient = AutoRefreshPrpcClient;
