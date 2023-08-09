// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import * as TestMetricsContext from '../../features/resources/tests/TestMetricsContext';
import { Period, SortType } from '../../api/resources';
import { ROWS_PER_PAGE } from '../../features/resources/tests/TestMetricsSearchParams';
import { formatDate } from '../../utils/formatUtils';
import TestsPage from './TestsPage';

export function renderWithBrowserRouter(
    ui: ReactElement,
) {
  render(
      <BrowserRouter>
        {ui}
      </BrowserRouter>,
  );
}

describe('when rendering the TestsPage', () => {
  // This is needed to allow us to modify window.location
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { assign: jest.fn() },
  });

  it('should pass in default values', async () => {
    const mockContext = jest.fn();
    jest.spyOn(TestMetricsContext, 'TestMetricsContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    renderWithBrowserRouter(<TestsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 0,
          rowsPerPage: 50,
          filter: '',
          period: Period.WEEK,
          sort: SortType.SORT_AVG_CORES,
          ascending: false,
          sortIndex: 0,
          timelineView: false,
          directoryView: false,
        }),
    );
    // Adding this check here to verify dates are correct
    expect(formatDate(mockContext.mock.calls[0][0].date)).toEqual(formatDate(new Date()));
  });

  it('should pass in url param values', async () => {
    const mockContext = jest.fn();
    jest.spyOn(TestMetricsContext, 'TestMetricsContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    window.location.search = 'https://localhost/?placeholder'+
    '=placeholder&p=10&rows=500&filter=filter&period=1&sort=2&asc=true&sidx=2&tl=true&dir=true';
    renderWithBrowserRouter(<TestsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 10,
          rowsPerPage: 500,
          filter: 'filter',
          period: Period.DAY,
          sort: SortType.SORT_NUM_RUNS,
          ascending: true,
          sortIndex: 2,
          timelineView: true,
          directoryView: true,
        }),
    );
    // Adding this check here to verify dates are correct
    expect(formatDate(mockContext.mock.calls[0][0].date)).toEqual(formatDate(new Date()));
  });

  it('should pass in local storage values', async () => {
    const mockContext = jest.fn();
    window.location.search = '';
    jest.spyOn(TestMetricsContext, 'TestMetricsContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    localStorage.setItem(ROWS_PER_PAGE, '250');
    renderWithBrowserRouter(<TestsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          rowsPerPage: 250,
        }),
    );
  });

  it('if local storage rows per page is 0, it should use the default', async () => {
    const mockContext = jest.fn();
    window.location.search = '';
    jest.spyOn(TestMetricsContext, 'TestMetricsContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    localStorage.setItem(ROWS_PER_PAGE, '0');
    renderWithBrowserRouter(<TestsPage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          rowsPerPage: 50,
        }),
    );
  });
});
