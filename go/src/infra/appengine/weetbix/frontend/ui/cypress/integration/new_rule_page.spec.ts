// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

describe('New Rule Page', () => {
  beforeEach(() => {
    // Login.
    cy.visit('/').contains('LOGIN').click();
  });
  it('create rule from scratch', () => {
    cy.visit('/p/chromium/rules/new');

    cy.get('new-rule-page').get('[data-cy=bug-system-dropdown]').contains('crbug.com');
    cy.get('new-rule-page').get('[data-cy=bug-number-textbox]').get('[type=text]').type('{selectall}101');
    cy.get('new-rule-page').get('[data-cy=rule-definition-textbox]').get('textarea').type('{selectall}test = "create test 1"');

    cy.intercept('POST', '/prpc/weetbix.v1.Rules/Create', (req) => {
      const requestBody = req.body;
      assert.strictEqual(requestBody.rule.ruleDefinition, 'test = "create test 1"');
      assert.deepEqual(requestBody.rule.bug, { system: 'monorail', id: 'chromium/101' });
      assert.deepEqual(requestBody.rule.sourceCluster, { algorithm: '', id: '' });

      const response = {
        project: 'chromium',
        // This is a real rule that exists in the dev database, the
        // same used for rule section UI tests.
        ruleId: '4165d118c919a1016f42e80efe30db59',
      };
      // Construct pRPC response.
      const body = ')]}\'' + JSON.stringify(response);
      req.reply(body, {
        'X-Prpc-Grpc-Code': '0',
      });
    }).as('createRule');

    cy.get('new-rule-page').get('[data-cy=create-button]').click();
    cy.wait('@createRule');

    // Verify the rule page loaded.
    cy.get('body').contains('Associated Bug');
  });
  it('create rule from cluster', () => {
    // Use an invalid rule to ensure it does not get created in dev by
    // accident.
    const rule = 'test = CREATE_TEST_2';
    cy.visit(`/p/chromium/rules/new?rule=${encodeURIComponent(rule)}&sourceAlg=reason-v1&sourceId=1234567890abcedf1234567890abcedf`);

    cy.get('new-rule-page').get('[data-cy=bug-system-dropdown]').contains('crbug.com');
    cy.get('new-rule-page').get('[data-cy=bug-number-textbox]').get('[type=text]').type('{selectall}101');

    cy.intercept('POST', '/prpc/weetbix.v1.Rules/Create', (req) => {
      const requestBody = req.body;
      assert.strictEqual(requestBody.rule.ruleDefinition, 'test = CREATE_TEST_2');
      assert.deepEqual(requestBody.rule.bug, { system: 'monorail', id: 'chromium/101' });
      assert.deepEqual(requestBody.rule.sourceCluster, { algorithm: 'reason-v1', id: '1234567890abcedf1234567890abcedf' });

      const response = {
        project: 'chromium',
        // This is a real rule that exists in the dev database, the
        // same used for rule section UI tests.
        ruleId: '4165d118c919a1016f42e80efe30db59',
      };
      // Construct pRPC response.
      const body = ')]}\'' + JSON.stringify(response);
      req.reply(body, {
        'X-Prpc-Grpc-Code': '0',
      });
    }).as('createRule');

    cy.get('new-rule-page').get('[data-cy=create-button]').click();
    cy.wait('@createRule');

    // Verify the rule page loaded.
    cy.get('body').contains('Associated Bug');
  });
  it('displays validation errors', () => {
    cy.visit('/p/chromium/rules/new');
    cy.get('new-rule-page').get('[data-cy=bug-system-dropdown]').contains('crbug.com');
    cy.get('new-rule-page').get('[data-cy=bug-number-textbox]').get('[type=text]').type('{selectall}101');
    cy.get('new-rule-page').get('[data-cy=rule-definition-textbox]').get('textarea').type('{selectall}test = INVALID');

    cy.get('new-rule-page').get('[data-cy=create-button]').click();

    cy.get('body').contains('Validation error: rule definition is not valid: undeclared identifier "invalid".');
  });
});
