// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert, expect} from 'chai';
import {IssueWizardPersona, IssueCategory, CustomQuestionType} from '../issue-wizard/IssueWizardTypes.tsx';
import {GetCategoriesByPersona, GetQuestionsByCategory, buildIssueDescription, getChromeVersion} from '../issue-wizard/IssueWizardUtils.tsx';

describe('IssueWizardUtils', () => {
  it('generate the issue categories to user persona map', () => {
    const categories: IssueCategory[]= [
      {
        name: 't1',
        description: 'd1',
        persona: IssueWizardPersona.EndUser,
        enabled: true,
      },
      {
        name: 't2',
        description: 'd2',
        persona: IssueWizardPersona.EndUser,
        enabled: false,
      },
    ];

    const categoriesByPersonaMap = GetCategoriesByPersona(categories);
    const validCategories = categoriesByPersonaMap.get(IssueWizardPersona.EndUser);

    assert.equal(validCategories?.length, 1);
    assert.equal(validCategories[0].name, 't1');
    assert.equal(validCategories[0].description, 'd1');
  });

  it('generate custom questions to issue categories map', () => {
    const categories: IssueCategory[]= [
      {
        name: 't1',
        description: 'd1',
        persona: IssueWizardPersona.EndUser,
        enabled: true,
        customQuestions: [
          {
            type: CustomQuestionType.Text,
            question: 'q1',
          }
        ]
      },
    ];

    const questionsByCategoryMap = GetQuestionsByCategory(categories);
    const questions = questionsByCategoryMap.get('t1');

    assert.equal(questions?.length, 1);
    assert.equal(questions[0].question, 'q1');
  });

  it('create issue description', () => {
    const description = buildIssueDescription('reproduce', 'description', 'comments', 'Mac', 'Chrome');
    expect(description).to.contains('Steps to reproduce the problem:');
    expect(description).to.contains('Problem Description:');
    expect(description).to.contains('Additional Comments:');
  });

  it('test the chrome version regex match', () => {
    const navigatorMock = {
      userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.109 Safari/537.36'
    };
    Object.defineProperty(window, 'navigator', {
        value: navigatorMock
      });
    const chrome_version = getChromeVersion();
    assert(chrome_version, '98.0.4758.109');
  });
});
