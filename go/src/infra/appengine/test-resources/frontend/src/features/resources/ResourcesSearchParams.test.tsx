/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { useSearchParams } from 'react-router-dom';
import { act, screen } from '@testing-library/react';
import { renderWithContext } from '../../utils/testUtils';
import { Params } from '../context/MetricsContext';
import { Period, SortType } from '../../api/resources';
import ResourcesParamControls, { ASCENDING, DATE, FILTER, PAGE, PERIOD, ROWS_PER_PAGE, SORT_BY } from './ResourcesSearchParams';

const TestingComponent = () => {
  const [search] = useSearchParams();
  return (
    <>
      <div>page-{search.get(PAGE)}</div>
      <div>rowsPerPage-{search.get(ROWS_PER_PAGE)}</div>
      <div>filter-{search.get(FILTER)}</div>
      <div>date-{search.get(DATE)}</div>
      <div>period-{search.get(PERIOD)}</div>
      <div>sortby-{search.get(SORT_BY)}</div>
      <div>ascending-{search.get(ASCENDING)}</div>
    </>
  );
};

const params: Params = {
  page: 12,
  rowsPerPage: 25,
  filter: 'filter',
  date: new Date('2020-01-01T00:00:00'),
  period: Period.DAY,
  sort: SortType.SORT_NAME,
  ascending: true,
};

describe('when rendering the ResourcesTableToolbar', () => {
  it('should render toolbar elements', async () => {
    await act(async () => {
      renderWithContext(<>
        <ResourcesParamControls/>
        <TestingComponent/>
      </>
      , { params },
      );
    });
    expect(screen.getByText('page-12')).toBeInTheDocument();
    expect(screen.getByText('rowsPerPage-25')).toBeInTheDocument();
    expect(screen.getByText('filter-filter')).toBeInTheDocument();
    expect(screen.getByText('date-2020-01-01')).toBeInTheDocument();
    expect(screen.getByText('period-0')).toBeInTheDocument();
    expect(screen.getByText('sortby-0')).toBeInTheDocument();
    expect(screen.getByText('ascending-true')).toBeInTheDocument();
  });
});
