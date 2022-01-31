// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

let authState : AuthState | null = null;

// obtainAuthState obtains a current Auth state, for interacting
// with Weetbix pRPC APIs.
export async function obtainAuthState(): Promise<AuthState> {
    if (authState != null && authState.accessTokenExpiry * 1000 > Date.now()) {
        // Auth state is still valid.
        return authState;
    }

    // Refresh the auth state.
    const response = await queryAuthState();
    authState = response;
    return authState;
}

export interface AuthState {
    identity: string;
    email: string;
    picture: string;
    accessToken: string;
    /**
     * Expiration time (unix timestamp) of the access token.
     *
     * If zero/undefined, the access token does not expire.
     */
    accessTokenExpiry: number;
}

export async function queryAuthState(): Promise<AuthState> {
    const res = await fetch('/api/authState');
    if (!res.ok) {
        throw new Error('failed to get authState:\n' + (await res.text()));
    }
    return res.json();
}