// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ReactElement } from 'react';
import * as MetricsContextP from '../../features/context/MetricsContext';
import { Period, SortType } from '../../api/resources';
import { ROWS_PER_PAGE } from '../../features/resources/ResourcesSearchParams';
import { formatDate } from '../../utils/formatUtils';
import ResourcesPage from './ResourcesPage';

export function renderWithBrowserRouter(
    ui: ReactElement,
) {
  render(
      <BrowserRouter>
        {ui}
      </BrowserRouter>,
  );
}

describe('when rendering the ResourcesPage', () => {
  // This is needed to allow us to modify window.location
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { assign: jest.fn() },
  });
  it('should pass in default values', async () => {
    const mockMetricsContext = jest.fn();
    jest.spyOn(MetricsContextP, 'MetricsContextProvider').mockImplementation((props) => {
      return mockMetricsContext(props);
    });
    renderWithBrowserRouter(<ResourcesPage/>);
    expect(mockMetricsContext).toHaveBeenCalledWith(
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
    expect(formatDate(mockMetricsContext.mock.calls[0][0].date)).toEqual(formatDate(new Date()));
  });
  it('should pass in url param values', async () => {
    const mockMetricsContext = jest.fn();
    jest.spyOn(MetricsContextP, 'MetricsContextProvider').mockImplementation((props) => {
      return mockMetricsContext(props);
    });
    window.location.search = 'https://test.com/?placeholder'+
    '=placeholder&page=10&rows=500&filter=filter&period=1&sort=2&asc=true&ind=2&tl=true&dir=true';
    renderWithBrowserRouter(<ResourcesPage/>);
    expect(mockMetricsContext).toHaveBeenCalledWith(
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
    expect(formatDate(mockMetricsContext.mock.calls[0][0].date)).toEqual(formatDate(new Date()));
  });
  it('should pass in local storage values', async () => {
    const mockMetricsContext = jest.fn();
    window.location.search = '';
    jest.spyOn(MetricsContextP, 'MetricsContextProvider').mockImplementation((props) => {
      return mockMetricsContext(props);
    });
    localStorage.setItem(ROWS_PER_PAGE, '250');
    renderWithBrowserRouter(<ResourcesPage/>);
    expect(mockMetricsContext).toHaveBeenCalledWith(
        expect.objectContaining({
          rowsPerPage: 250,
        }),
    );
  });
});
