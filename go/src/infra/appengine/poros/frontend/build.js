// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import rewire from 'rewire';
const defaults = rewire('react-scripts/scripts/build.js');
const config = defaults.__get__('config');

config.optimization.splitChunks = {
  cacheGroups: {
    default: false,
  },
};

config.optimization.runtimeChunk = false;

// JS
config.output.filename = 'js/[name].js';
// CSS
config.plugins[5].options.filename = 'css/[name].css';
