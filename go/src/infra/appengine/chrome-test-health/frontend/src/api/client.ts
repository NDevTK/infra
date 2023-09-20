// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from './auth';

export const prpcClient = {
  call: async function <Type>(
      auth: Auth,
      service: string,
      method: string,
      message: unknown,
  ): Promise<Type> {
    return auth.validateOrRedirect().then(async (auth) => {
      if (auth === undefined) {
        return;
      }
      const url = `/prpc/${service}/${method}`;
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          'Authorization': 'OAuth ' + auth?.accessToken,
        },
        body: JSON.stringify(message),
      });
      const text = await response.text();
      if (text.startsWith(')]}\'')) {
        return JSON.parse(text.substring(4));
      } else {
        throw text;
      }
    });
  },
};
