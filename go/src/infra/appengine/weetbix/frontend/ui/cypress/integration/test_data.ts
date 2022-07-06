// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

export function setupTestRule() {
    cy.request({
        url: '/api/authState',
        headers: {
            'Sec-Fetch-Site': 'same-origin',
        }
    }).then((response) => {
        assert.strictEqual(response.status, 200);
        const body = response.body;
        const accessToken = body.accessToken;
        assert.isString(accessToken);
        assert.notEqual(accessToken, '');

        // Set initial rule state.
        cy.request({
            method: 'POST',
            url:  '/prpc/weetbix.v1.Rules/Update',
            body: {
                rule: {
                    name: 'projects/chromium/rules/4165d118c919a1016f42e80efe30db59',
                    ruleDefinition: 'test = "cypress test 1"',
                    bug: {
                        system: 'monorail',
                        id: 'chromium/920867',
                    },
                    isActive: true,
                    isManagingBug: true,
                },
                updateMask: 'ruleDefinition,bug,isActive,isManagingBug',
            },
            headers: {
                Authorization: 'Bearer ' + accessToken,
            },
        });
    });
}