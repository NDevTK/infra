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

function ajax(url) {
  const xhr = new XMLHttpRequest();
  xhr.withCredentials = true;
  xhr.onabort = function() {
    console.log("onabort")
  }
  xhr.onerror = function() {
    console.log("onerror")
  }
  xhr.onreadystatechange = function() {
    console.log("xhr.readyState: " + xhr.readyState);
    console.log("xhr.getAllResponseHeaders: " + xhr.getAllResponseHeaders());
    console.log("xhr.status: " + xhr.status);
    console.log("xhr.statusText: " + xhr.statusText);
    // return if not ready state 4
    if (xhr.readyState !== 4) {
      return;
    }

    // check for redirect
    if (xhr.status === 302) {
      const location = xhr.getResponseHeader("Location");
      console.log("location: " + location);
      return ajax.call(xhr, location);
    }
  };
  xhr.open("GET", url, false);
  xhr.send();
}

export async function loginOrRedirect(): Promise<Auth | undefined> {
  return fetchAuthState().then((response) => {
    if (response.accessToken && response.accessTokenExpiry ) {
      ajax('https://teamsgraph.corp.googleapis.com/v1/suggest:teams?access_token=' + response.accessToken) // pragma: nocover

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
