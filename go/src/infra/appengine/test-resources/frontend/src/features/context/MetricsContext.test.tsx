/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { fireEvent, render, screen } from '@testing-library/react';
import { act } from 'react-dom/test-utils';
import { Button } from '@mui/material';
import * as Resources from '../../api/resources';
import { MetricsContext, MetricsContextProvider, MetricsContextValue } from './MetricsContext';

async function contextRender(ui: (value: MetricsContextValue) => React.ReactElement, { props } = { props: {} }) {
  await act(async () => {
    render(
        <MetricsContextProvider {... props}>
          <MetricsContext.Consumer>
            {(value) => ui(value)}
          </MetricsContext.Consumer>
        </MetricsContextProvider>,
    );
  },
  );
}

describe('MetricsContext params', () => {
  beforeEach(() => {
    jest.spyOn(Resources, 'fetchTestMetrics').mockResolvedValue({
      tests: [],
      lastPage: true,
    });
  });

  it('page', async () => {
    await contextRender((value) => (
      <Button data-testid='updatePage' onClick={() => value.api.updatePage(20)}>{'page-' + value.params.page}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePage'));
    });
    expect(screen.getByText('page-20')).toBeInTheDocument();
  });

  it('filter', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updateFilter' onClick={() => value.api.updateFilter('filt')}>{'filter-' + value.params.filter}</Button>
        <div>page-{value.params.page}</div>
      </>
    ), { props: { page: 1 } });
    expect(screen.getByText('page-1')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateFilter'));
    });
    expect(screen.getByText('filter-filt')).toBeInTheDocument();
    expect(screen.getByText('page-0')).toBeInTheDocument();
  });
});
