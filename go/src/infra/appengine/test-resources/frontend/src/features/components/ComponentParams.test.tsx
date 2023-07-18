// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter } from 'react-router-dom';
import { act, render } from '@testing-library/react';
import { ReactElement } from 'react';
import ComponentParams, { COMPONENT } from './ComponentParams';
import { ComponentContext, ComponentContextValue } from './ComponentContext';

export function renderWithComponentContext(
    ui: ReactElement,
    ctx: ComponentContextValue,
) {
  render(
      <BrowserRouter>
        <ComponentContext.Provider value= {ctx}>
          {ui}
        </ComponentContext.Provider>,
      </BrowserRouter>,
  );
}

describe('when rendering ComponentParams', () => {
  it('should render url correctly', async () => {
    await act(async () => {
      renderWithComponentContext(
          <>
            <ComponentParams/>
          </>
          , { component: 'Blink', allComponents: [], api: { updateComponent: () =>{/**/} } });
    });
    const searchParams = new URLSearchParams(window.location.search);
    expect(searchParams.get(COMPONENT)).toBe('Blink');
  });
});
