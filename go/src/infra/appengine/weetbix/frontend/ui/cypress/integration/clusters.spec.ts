// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

describe('Clusters Page', () => {
  beforeEach(() => {
    cy.visit('/').contains('LOGIN').click();
    cy.get('body').contains('Logout');
    // Use fuchsia for now as the loading time is faster.
    cy.visit('/p/fuchsia/clusters');
  });
  it('loads rules table', () => {
    // Navigate to the bug cluster page
    cy.contains('Rules').click();
    // check for the header text in the bug cluster table.
    cy.contains('Rule Definition');
  });
  it('loads cluster table', () => {
    // check for an entry in the cluster table.
    cy.get('[data-testid=clusters_table_body]').contains('test = ');
  });
  it('loads a cluster page', () => {
    cy.get('[data-testid=clusters_table_title] > a').first().click();
    cy.get('body').contains('Recent Failures');
    // Check that the analysis section is showing at least one group.
    cy.get('[data-testid=failures_table_group_cell]');
  });
});
