// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {ReactElement} from 'react';
import * as React from 'react'
import ReactDOM from 'react-dom';
import styles from './IssueWizard.css';
import LandingStep from './issue-wizard/LandingStep.tsx';
import DetailsStep from './issue-wizard/DetailsStep.tsx'
import {IssueWizardPersona} from './issue-wizard/IssueWizardTypes.tsx';
import CustomQuestionsStep from './issue-wizard/CustomQuestionsStep.tsx';
import {getOs, getBrowser} from './issue-wizard/IssueWizardUtils.tsx'

import {GetQuestionsByCategory} from './issue-wizard/IssueWizardUtils.tsx';
import {ISSUE_WIZARD_QUESTIONS} from './issue-wizard/IssueWizardConfig.ts';

/**
 * Base component for the issue filing wizard, wrapper for other components.
 * @return Issue wizard JSX.
 */
export function IssueWizard(): ReactElement {
  const [userPersona, setUserPersona] = React.useState(IssueWizardPersona.EndUser);
  const [activeStep, setActiveStep] = React.useState(0);
  const [category, setCategory] = React.useState('');
  const [textValues, setTextValues] = React.useState(
    {
      oneLineSummary: '',
      stepsToReproduce: '',
      describeProblem: '',
      additionalComments: ''
    });
    const [osName, setOsName] = React.useState(getOs())
    const [browserName, setBrowserName] = React.useState(getBrowser())

  const questionByCategory = GetQuestionsByCategory(ISSUE_WIZARD_QUESTIONS);

  let page;
  if (activeStep === 0) {
    page = <LandingStep
        userPersona={userPersona}
        setUserPersona={setUserPersona}
        category={category}
        setCategory={setCategory}
        setActiveStep={setActiveStep}
        />;
      } else if (activeStep === 1) {
        page = <DetailsStep
          textValues={textValues}
          setTextValues={setTextValues}
          category={category}
          setActiveStep={setActiveStep}
          osName={osName}
          setOsName={setOsName}
          browserName={browserName}
          setBrowserName={setBrowserName}
    />;
   } else if (activeStep === 2) {

    page = <CustomQuestionsStep setActiveStep={setActiveStep} questions={questionByCategory.get(category)}/>;
  }

  return (
    <>
      <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Poppins"></link>
      <div className={styles.container}>
        {page}
      </div>
    </>
  );
}

/**
 * Renders the issue filing wizard page.
 * @param mount HTMLElement that the React component should be
 *   added to.
 */
export function renderWizard(mount: HTMLElement): void {
  ReactDOM.render(<IssueWizard />, mount);
}
