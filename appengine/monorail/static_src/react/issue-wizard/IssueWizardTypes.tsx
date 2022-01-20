// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// this const is used on issue wizard lading page for render user role  options
export const IssueWizardPersonas  = Object.freeze({
  EndUser: {
    name: 'End User',
    description: 'I am a user trying to do something on a website.',
  },
  Developer: {
    name: 'Web Developer',
    description: 'I am a web developer trying to build something.',
  },
  Contributer: {
    name: 'Chromium Contributor',
    description: 'I know about a problem in specific tests or code.',
  }
});

export type IssueWizardPersonasType = typeof IssueWizardPersonas [keyof typeof IssueWizardPersonas];

export enum CustomQuestionType {
  EMPTY, // this is used to define there is no subquestions
  Text,
  Input,
  Select,
}
export type CustomQuestion = {
  type: CustomQuestionType,
  question: string,
  options?: string[],
  subQuestions?: CustomQuestion[],
};

export type IssueCategory = {
  name: string,
  persona: IssueWizardPersonasType,
  enabled: boolean,
  customQuestions?: CustomQuestion[],
};
