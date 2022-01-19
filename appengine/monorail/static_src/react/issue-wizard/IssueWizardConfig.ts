// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// TODO: create a `monorail/frontend/config/` folder to store all the feature config file
import {IssueCategory, IssueWizardPersona} from "./IssueWizardTypes";

export const ISSUE_WIZARD_QUESTIONS: IssueCategory[] = [
  {
    name: 'UI',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Network / Downloading',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Audio / Video',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Content',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Apps',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Extensions / Themes',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Webstore',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Sync',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Enterprise',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Installation',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Crashes',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Security',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'Other',
    user: IssueWizardPersona.EndUser,
    enabled: true,
  },
  {
    name: 'API',
    user: IssueWizardPersona.Developer,
    enabled: true,
  },
  {
    name: 'JavaScript',
    user: IssueWizardPersona.Developer,
    enabled: true,
  },
  {
    name: 'Developer Tools',
    user: IssueWizardPersona.Developer,
    enabled: true,
  },
];
