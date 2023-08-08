// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { obtainAuthState, clearAuthState } from './common/auth_state';

export interface IUtilityService {
  getUserPicture(): Promise<string>;
  logout(): Promise<void>;
}

export class UtilityService implements IUtilityService {
  constructor() {
    this.getUserPicture = this.getUserPicture.bind(this);
  }

  getUserPicture = (): Promise<string> => {
    return obtainAuthState().then((authState) => authState.picture);
  };

  logout = (): Promise<void> => {
    return clearAuthState();
  };
}
