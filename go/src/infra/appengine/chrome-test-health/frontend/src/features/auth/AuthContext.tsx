// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useState } from 'react';
import { Auth, loginOrRedirect } from '../../api/auth';

export interface AuthContextValue {
  auth?: Auth,
}

export const AuthContext = createContext<AuthContextValue>({
});

export const AuthContextProvider = (props: { children: React.ReactNode }) => {
  const [auth, setAuthToken] = useState<Auth | undefined>(undefined);

  useEffect(() => {
    // On mount, log in.
    loginOrRedirect().then((auth) => {
      setAuthToken(auth);
    });
  }, []);

  return (
    <AuthContext.Provider value={{ auth }}>
      { props.children }
    </AuthContext.Provider>
  );
};
