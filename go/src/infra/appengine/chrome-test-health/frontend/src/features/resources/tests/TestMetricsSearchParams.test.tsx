/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { act } from '@testing-library/react';
import { MetricType, Period, SortType } from '../../../api/resources';
import { renderWithComponents } from '../../components/testUtils';
import { renderWithContext } from './testUtils';
import { Params } from './TestMetricsContext';
import TestMetricsSearchParams, {
  ASCENDING,
  DATE,
  FILTER,
  PAGE,
  PERIOD,
  ROWS_PER_PAGE,
  SORT_BY,
  TIMELINE_VIEW,
  DIRECTORY_VIEW,
  SORT_INDEX,
  TIMELINE_VIEW_METRIC,
} from './TestMetricsSearchParams';

describe('when rendering the ResourcesSearchParams', () => {
  it('should render url corrently', async () => {
    const params: Params = {
      page: 12,
      rowsPerPage: 25,
      filter: 'filter',
      date: new Date('2020-01-02T00:00:00'),
      period: Period.DAY,
      sort: SortType.SORT_NAME,
      ascending: true,
      sortIndex: 0,
      timelineMetric: MetricType.AVG_CORES,
      timelineView: true,
      directoryView: false,
    };

    await act(async () => {
      renderWithContext(<>
        <TestMetricsSearchParams/>
      </>
      , { params },
      );
    });
    const searchParams = new URLSearchParams(window.location.search);
    expect(searchParams.get(PAGE)).toBe('12');
    expect(searchParams.get(ROWS_PER_PAGE)).toBe('25');
    expect(searchParams.get(FILTER)).toBe('filter');
    expect(searchParams.get(DATE)).toBe('2020-01-02');
    expect(searchParams.get(PERIOD)).toBe(Period.DAY.toString());
    expect(searchParams.get(SORT_BY)).toBe(SortType.SORT_NAME.toString());
    expect(searchParams.get(ASCENDING)).toBe('true');
    expect(searchParams.get(TIMELINE_VIEW)).toBe('true');
    expect(searchParams.get(TIMELINE_VIEW_METRIC)).toBe('AVG_CORES');
    expect(searchParams.get(DIRECTORY_VIEW)).toBe('false');
    expect(searchParams.get(SORT_INDEX)).toBe('0');
    expect(global.localStorage.getItem(ROWS_PER_PAGE)).toEqual('25');
  });

  it('should render url without empty params', async () => {
    const params: Params = {
      page: 0,
      rowsPerPage: 25,
      filter: '',
      date: new Date('2020-01-02T00:00:00'),
      period: Period.DAY,
      sort: SortType.SORT_NAME,
      ascending: true,
      sortIndex: 0,
      timelineMetric: MetricType.AVG_CORES,
      timelineView: false,
      directoryView: true,
    };

    await act(async () => {
      renderWithContext(<>
        <TestMetricsSearchParams/>
      </>
      , { params },
      );
    });
    const searchParams = new URLSearchParams(window.location.search);
    expect(searchParams.get(ROWS_PER_PAGE)).toBe(null);
    expect(searchParams.get(FILTER)).toBe(null);
    expect(searchParams.get(PAGE)).toBe(null);
    expect(searchParams.get(SORT_INDEX)).toBe(null);
  });

  it('should render components in url', async () => {
    await act(async () => {
      renderWithComponents((
        <>
          <TestMetricsSearchParams/>
        </>
      ), { components: ['a', 'b'] },
      );
    });
    const searchParams = new URLSearchParams(window.location.search);
    expect(searchParams.getAll('c')).toEqual(['a', 'b']);
  });
});
