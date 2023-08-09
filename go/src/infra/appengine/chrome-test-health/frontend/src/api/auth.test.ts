// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth, loginOrRedirect } from './auth';

describe('loginOrRedirect', () => {
  it('returns auth when logged in', async () => {
    jest.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve({
        status: 200,
        json: () => Promise.resolve({
          accessToken: 'token',
          accessTokenExpiry: new Date('3000-01-01').getTime() / 1000,
        }),
      } as Response),
    );
    const resp = await loginOrRedirect();
    expect(resp).toBeDefined();
  });

  it('redirects auth when not logged in', async () => {
    jest.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve({
        status: 200,
        json: () => Promise.resolve({
          identity: 'anonymous:anonymous',
        }),
      } as Response),
    );
    const assignMock = jest.fn();
    Object.defineProperty(window, 'location', {
      value: { assign: assignMock },
    });
    const resp = await loginOrRedirect();
    expect(resp).toBeUndefined();
    expect(assignMock.mock.calls).toHaveLength(1);
    expect(assignMock.mock.calls[0][0]).toMatch(/^\/auth\/openid\/login/);
  });

  it('refetches auth when expired', async () => {
    jest.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve({
        status: 200,
        json: () => Promise.resolve({
          accessToken: 'token2',
          accessTokenExpiry: new Date('3000-01-01').getTime() / 1000,
        }),
      } as Response),
    );
    const auth = new Auth('token', new Date(0));
    const resolveMock = jest.fn();
    await auth.validateOrRedirect().then(resolveMock);
    expect(resolveMock.mock.calls).toHaveLength(1);
    expect(resolveMock.mock.calls[0][0]).toBe(auth);
    expect(auth.accessToken).toBe('token2');
    expect(auth.accessTokenExpiry.getTime()).toBeGreaterThan(0);
  });
});
