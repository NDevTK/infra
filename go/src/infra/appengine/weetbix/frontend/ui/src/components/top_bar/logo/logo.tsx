// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';

const Logo = () => {
  return (
    <img
      style={{ width: '100%' }}
      alt="logo"
      id="chromium-icon"
      src="https://storage.googleapis.com/chrome-infra/lucy-small.png" />
  );
};

export default React.memo(Logo);
