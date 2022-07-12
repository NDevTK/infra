// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import { setupTestRule } from './test_data';

describe('Rule Section', () => {
  beforeEach(() => {
    // Login.
    cy.visit('/').contains('LOGIN').click();

    setupTestRule();

    cy.visit('/p/chromium/rules/4165d118c919a1016f42e80efe30db59');
  });

  it('loads rule', () => {
    cy.get('[data-testid=bug-summary]').contains('Weetbix Cypress Test Bug');
    cy.get('[data-testid=bug-status]').contains('Verified');
    cy.get('[data-testid=rule-definition]').contains('test = "cypress test 1"');
    cy.get('[data-testid=rule-archived]').contains('No');
    cy.get('[data-testid=update-bug-toggle]').get('[type=checkbox]').should('be.checked');
  });

  it('edit rule definition', () => {
    cy.get('[data-testid=rule-definition-edit]').click();
    cy.get('[data-testid=rule-input]').type('{selectall}test = "cypress test 2"');
    cy.get('[data-testid=rule-edit-dialog-save]').click();
    cy.get('[data-testid=rule-definition]').contains('test = "cypress test 2"');
    cy.get('[data-testid=reclustering-progress-description]').contains('Weetbix is re-clustering test results');
  });

  it('validation error while editing rule definition', () => {
    cy.get('[data-testid=rule-definition-edit]').click();
    cy.get('[data-testid=rule-input]').type('{selectall}test = "cypress test 2"a');
    cy.get('[data-testid=rule-edit-dialog-save]').click();
    cy.get('[data-testid=snackbar]').contains('rule definition is not valid: syntax error: 1:24: unexpected token "a"');
    cy.get('[data-testid=rule-edit-dialog-cancel]').click();
    cy.get('[data-testid=rule-definition]').contains('test = "cypress test 1"');
  });

  it('edit bug', () => {
    cy.get('[data-testid=bug-edit]').click();
    cy.get('[data-testid=bug-number').type('{selectall}920869');
    cy.get('[data-testid=bug-edit-dialog-save]').click();
    cy.get('[data-testid=bug]').contains('crbug.com/920869');
    cy.get('[data-testid=bug-summary]').contains('Weetbix Cypress Alternate Test Bug');
    cy.get('[data-testid=bug-status]').contains('Fixed');
  });

  it('validation error while editing bug', () => {
    cy.get('[data-testid=bug-edit]').click();
    cy.get('[data-testid=bug-number').type('{selectall}125a');
    cy.get('[data-testid=bug-edit-dialog-save]').click();
    cy.get('[data-testid=snackbar]').contains('not a valid monorail bug ID');
    cy.get('[data-testid=bug-edit-dialog-cancel]').click();
    cy.get('[data-testid=bug]').contains('crbug.com/920867');
  });

  it('archive and restore', () => {
    cy.get('[data-testid=rule-archived-toggle]').contains('Archive').click();
    cy.get('[data-testid=confirm-dialog-cancel]').click();
    cy.get('[data-testid=rule-archived]').contains('No');

    cy.get('[data-testid=rule-archived-toggle]').contains('Archive').click();
    cy.get('[data-testid=confirm-dialog-confirm]').click();
    cy.get('[data-testid=rule-archived]').contains('Yes');

    cy.get('[data-testid=rule-archived-toggle]').contains('Restore').click();
    cy.get('[data-testid=confirm-dialog-confirm]').click();
    cy.get('[data-testid=rule-archived]').contains('No');
  });

  it('toggle bug updates', () => {
    cy.get('[data-testid=update-bug-toggle]').click();
    // Cypress assertion should('not.be.checked') does not work for MUI Switch.
    cy.get('[data-testid=update-bug-toggle]').should('not.have.class', 'Mui-checked');

    cy.get('[data-testid=update-bug-toggle]').click();
    cy.get('[data-testid=update-bug-toggle]').should('have.class', 'Mui-checked');
  });
});
