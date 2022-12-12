// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import { render, screen, cleanup, fireEvent } from '@testing-library/react';
import { assert } from 'chai';
import sinon from 'sinon';

import { RadioDescription } from './RadioDescription.tsx';
import {IssueWizardPersona} from '../IssueWizardTypes.tsx';

describe('RadioDescription', () => {
  afterEach(cleanup);

  it('renders', () => {
    render(<RadioDescription />);
    // look for blue radios
    const radioOne = screen.getByRole('radio', { name: /Web Developer/i });
    assert.isNotNull(radioOne)

    const radioTwo = screen.getByRole('radio', { name: /End User/i });
    assert.isNotNull(radioTwo)

    const radioThree = screen.getByRole('radio', { name: /Chromium Contributor/i });
    assert.isNotNull(radioThree)
  });

  it('checks selected radio value', () => {
    // We're passing in the "Web Developer" value here manually
    // to tell our code that that radio button is selected.
    render(<RadioDescription selectedRadio={IssueWizardPersona.Developer} />);

    const checkedRadio = screen.getByRole('radio', { name: /Web Developer/i });
    assert.isTrue(checkedRadio.checked);

    // Extra check to make sure we haven't checked every single radio button.
    const uncheckedRadio = screen.getByRole('radio', { name: /End User/i });
    assert.isFalse(uncheckedRadio.checked);
  });

  it('sets radio value when radio button is clicked', () => {
    // Using the sinon.js testing library to create a function for testing.
    const setValue = sinon.stub();

    render(<RadioDescription onClickRadio={setValue} />);

    const radio = screen.getByRole('radio', { name: /Web Developer/i });
    fireEvent.click(radio);

    // Asserts that "Web Developer" was passed into our "setValue" function.
    sinon.assert.calledWith(setValue, IssueWizardPersona.Developer);
  });

  it('sets radio value when any part of the parent RoleSelection is clicked', () => {
    const setValue = sinon.stub();

    render(<RadioDescription onClickRadio={setValue} />);

    // Click text in the RoleSelection component
    const p = screen.getByText('End User');
    fireEvent.click(p);

    // Asserts that "End User" was passed into our "setValue" function.
    sinon.assert.calledWith(setValue, IssueWizardPersona.EndUser);
  });
});
