// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { PrpcClient } from '@chopsui/prpc-client';

import { obtainAuthState } from '../api/auth_state';

export class AuthorizedPrpcClient {
  client: PrpcClient;
  // Whether the ID token should be used to authorize the request.
  // If false, use the access token instead
  useIDToken: boolean;

  // Initializes a new AuthorizedPrpcClient that connects to host.
  // To connect to LUCI Bisection, leave host unspecified.
  constructor(host?: string, useIDToken?: boolean) {
    // Only allow insecure connections in LUCI Bisection in local development,
    // where risk of man-in-the-middle attack to server is negligible.
    const insecure = document.location.protocol === 'http:' && !host;
    if (insecure && document.location.hostname !== 'localhost') {
      // Server misconfiguration.
      throw new Error(
        'LUCI Bisection should never be served over http: outside local development.'
      );
    }

    this.client = new PrpcClient({
      host: host,
      insecure: insecure,
    });

    this.useIDToken = useIDToken === true;
  }

  async call(
    service: string,
    method: string,
    message: object,
    additionalHeaders?:
      | {
          [key: string]: string;
        }
      | undefined
  ): Promise<any> {
    // Although PrpcClient allows us to pass a token to the constructor,
    // we prefer to inject it at request time to ensure the most recent
    // token is used.
    const authState = await obtainAuthState();
    let token: string;
    if (this.useIDToken) {
      token = authState.idToken;
    } else {
      token = authState.accessToken;
    }
    additionalHeaders = {
      Authorization: 'Bearer ' + token,
      ...additionalHeaders,
    };
    return this.client.call(service, method, message, additionalHeaders);
  }
}
