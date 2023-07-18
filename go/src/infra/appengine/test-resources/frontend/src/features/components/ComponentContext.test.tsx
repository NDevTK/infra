// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { act } from 'react-dom/test-utils';
import { fireEvent, render, screen } from '@testing-library/react';
import { Button } from '@mui/material';
import * as Resources from '../../api/resources';
import { ComponentContext, ComponentContextProvider, ComponentContextValue } from './ComponentContext';

async function contextRender(ui: (value: ComponentContextValue) => React.ReactElement, { props } = { props: {} }) {
  await act(async () => {
    render(
        <ComponentContextProvider {... props}>
          <ComponentContext.Consumer>
            {(value) => ui(value)}
          </ComponentContext.Consumer>
        </ComponentContextProvider>,
    );
  },
  );
}

describe('ComponentContext values', () => {
  beforeEach(() => {
    jest.spyOn(Resources, 'listComponents').mockResolvedValue({
      components: ['1', '2', '3'],
    });
  });
  it('allComponents', async () => {
    await contextRender((value) => (
      <>
        {value.allComponents.map((c) => (<div data-testid='component' key={c}>{c}</div>))}
      </>
    ));
    const components = screen.getAllByTestId('component');
    expect(components).toHaveLength(3);
    expect(components[0]).toHaveTextContent('1');
    expect(components[1]).toHaveTextContent('2');
    expect(components[2]).toHaveTextContent('3');
  });
  it('component', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updateComponent' onClick={() => value.api.updateComponent('comp')}>{'component-' + value.component}</Button>
      </>
    ), { props: { component: 'blink' } });
    expect(screen.getByText('component-Blink')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateComponent'));
    });
    expect(screen.getByText('component-comp')).toBeInTheDocument();
  });
});
