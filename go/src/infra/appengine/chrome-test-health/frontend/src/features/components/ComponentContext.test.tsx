// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { act } from 'react-dom/test-utils';
import { fireEvent, screen } from '@testing-library/react';
import { Button } from '@mui/material';
import * as Resources from '../../api/resources';
import { renderWithAuth } from '../auth/testUtils';
import { ComponentContext, ComponentContextProvider, ComponentContextValue, URL_COMPONENT, updateComponentsUrl } from './ComponentContext';

function createProps(
    param : TestProps) {
  return {
    components: param.components || ['Blink'],
  };
}

type TestProps = {
components?: string[],
}

async function renderWithContext(ui: (context: ComponentContextValue) => React.ReactElement, { props } = { props: { ...createProps({}) } }) {
  await act(async () => {
    renderWithAuth(
        <ComponentContextProvider {... props}>
          <ComponentContext.Consumer>
            {(context) => ui(context)}
          </ComponentContext.Consumer>
        </ComponentContextProvider>,
    );
  });
}

describe('ComponentContext values', () => {
  beforeEach(() => {
    jest.spyOn(Resources, 'listComponents').mockResolvedValue({
      components: ['1', '2', '3'],
    });
  });
  it('allComponents', async () => {
    await renderWithContext((context) => (
      <>
        {context.allComponents.map((c) => (<div data-testid='component' key={c}>{c}</div>))}
      </>
    ));
    const components = screen.getAllByTestId('component');
    expect(components).toHaveLength(3);
    expect(components[0]).toHaveTextContent('1');
    expect(components[1]).toHaveTextContent('2');
    expect(components[2]).toHaveTextContent('3');
  });
  it('components', async () => {
    await renderWithContext((context) => (
      <>
        <Button data-testid='updateComponent' onClick={
          () => context.api.updateComponents(['comp', 'comp1'])
        }>
          {'components-' + context.components}
        </Button>
      </>
    ), { props: { ...createProps({ components: ['blink'] }) } });
    expect(screen.getByText('components-blink')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateComponent'));
    });
    expect(screen.getByText('components-comp,comp1')).toBeInTheDocument();
  });
});

export const TEST_SEARCH_PARAMS_ALL_COMPONENTS = 'ac=true'

describe('updateComponentsUrl', () => {
  it('sets multiple components', () => {
    const search = new URLSearchParams();
    updateComponentsUrl(['a', 'b', 'c'], search);
    expect(search.getAll(URL_COMPONENT)).toEqual(['a', 'b', 'c']);
    expect(global.localStorage.getItem(URL_COMPONENT)).toEqual('a,b,c');
  });

  it('sets all components', () => {
    const search = new URLSearchParams();
    updateComponentsUrl([], search);
    expect(search.toString()).toEqual(TEST_SEARCH_PARAMS_ALL_COMPONENTS);
  });
});
