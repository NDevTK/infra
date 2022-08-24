// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Alert from '@mui/material/Alert';

export const NotFoundPage = () => {
  return (
    <main>
      <Alert severity='error'>Page not found!</Alert>
    </main>
  );
};
