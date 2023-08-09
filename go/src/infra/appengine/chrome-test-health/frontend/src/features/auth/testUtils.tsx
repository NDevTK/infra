// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { Auth } from '../../api/auth';
import { AuthContext } from './AuthContext';

export function renderWithAuth(ui: React.ReactElement, auth: Auth = new Auth('', new Date())) {
  return render((
    <AuthContext.Provider value={{ auth: auth }}>
      {ui}
    </AuthContext.Provider>
  ));
}
