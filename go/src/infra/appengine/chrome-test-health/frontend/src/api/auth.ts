// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

interface AuthStateResponse {
  identity: string,
  email?: string,
  picture?: string,
  accessToken?: string,
  accessTokenExpiry?: number,
  accessTokenExpiresIn?: number,
  idToken?: string,
  idTokenExpiry?: number,
  idTokenExpiresIn?: number,
}

async function fetchAuthState(): Promise<AuthStateResponse> {
  const url = '/auth/openid/state';
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  });
  if (response.status == 200) {
    return response.json();
  } else {
    throw response.text();
  }
}

function redirect(url: string) {
  window.location.assign(url);
}

export class Auth {
  accessToken: string;
  accessTokenExpiry: Date;

  constructor(accessToken: string, accessTokenExpiry: Date) {
    this.accessToken = accessToken;
    this.accessTokenExpiry = accessTokenExpiry;
  }

  async validateOrRedirect(): Promise<Auth | undefined> {
    if (new Date().getTime() > this.accessTokenExpiry.getTime()) {
      // Expired
      return loginOrRedirect().then((auth) => {
        if (auth !== undefined) {
          this.accessToken = auth.accessToken;
          this.accessTokenExpiry = auth.accessTokenExpiry;
          return this;
        }
        return undefined;
      });
    } else {
      return this;
    }
  }
}

export async function loginOrRedirect(): Promise<Auth | undefined> {
  return fetchAuthState().then((response) => {
    if (response.accessToken && response.accessTokenExpiry ) {
      return new Auth(
          response.accessToken,
          // The expiry is in seconds since epoch while JS uses ms since epoch
          new Date(response.accessTokenExpiry * 1000),
      );
    } else {
      const ret = encodeURIComponent(
          window.location.pathname + window.location.search,
      );
      redirect('/auth/openid/login?r=' + ret);
      return undefined;
    }
  });
}
