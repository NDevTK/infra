// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

describe('Index Page', () => {
    beforeEach(() => {
        cy.visit('/').get('button').click();
        cy.get('body').contains('Logout');
    })
    it('loads monorail issue', () => {
        // Navigate to the monorail test page
        cy.contains('Monorail Test').click();
        // check for some text in the monorail issue.
        cy.get('monorail-test').contains('chromium id');
    })
    it('loads bug cluster table', () => {
        // Navigate to the bug cluster page
        cy.contains('Bug Clusters').click();
        // check for the header text in the bug cluster table.
        cy.get('bug-cluster-table').contains('Associated Cluster ID');
    })
    it('loads cluster table', () => {
        // check for the header text in the cluster table.
        cy.get('cluster-table').contains('Unexonerated');
    })
    it('loads a cluster page', () => {
        // navigate to the cluster page for the highest impact cluster.
        cy.get('cluster-table').get('td').first().click();
        cy.get('body').contains('Example Failure');

        // Note that this assumes the cluster we are looking at has a bug filed.
        // This should be safe since we selected the highest impact cluster in the system.
        cy.get('.bug');
    })
})
