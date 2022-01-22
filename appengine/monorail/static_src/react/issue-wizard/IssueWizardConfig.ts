// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// TODO: create a `monorail/frontend/config/` folder to store all the feature config file
import {IssueCategory, IssueWizardPersonas} from "./IssueWizardTypes";

export const ISSUE_WIZARD_QUESTIONS: IssueCategory[] = [
  {
    name: 'UI',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Network / Downloading',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Audio / Video',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Content',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Apps',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Extensions / Themes',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Webstore',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Sync',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Enterprise',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Installation',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Crashes',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Security',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'Other',
    persona: IssueWizardPersonas.EndUser,
    enabled: true,
  },
  {
    name: 'API',
    persona: IssueWizardPersonas.Developer,
    enabled: true,
  },
  {
    name: 'JavaScript',
    persona: IssueWizardPersonas.Developer,
    enabled: true,
  },
  {
    name: 'Developer Tools',
    persona: IssueWizardPersonas.Developer,
    enabled: true,
  },
];
