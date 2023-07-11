/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { fireEvent, render, screen } from '@testing-library/react';
import { useContext } from 'react';
import { act } from 'react-dom/test-utils';
import { Button } from '@mui/material';
import * as Resources from '../../api/resources';
import { MetricsContext, MetricsContextProvider } from './MetricsContext';

const TestingComponent = () => {
  const { api, params, data, lastPage, isLoading } = useContext(MetricsContext);
  return (
    <>
      <div>{'id-' + data[0]?.id}</div>
      <div>{'name-' + data[0]?.name}</div>
      <div>{'numRuns-' + data[0]?.metrics.get(Resources.MetricType.NUM_RUNS)}</div>
      <div>{'numFailures-' + data[0]?.metrics.get(Resources.MetricType.NUM_FAILURES)}</div>
      <div>{'avgRuntime-' + data[0]?.metrics.get(Resources.MetricType.AVG_RUNTIME)}</div>
      <div>{'totalRuntime-' + data[0]?.metrics.get(Resources.MetricType.TOTAL_RUNTIME)}</div>
      <div>{'avgCores-' + data[0]?.metrics.get(Resources.MetricType.AVG_CORES)}</div>
      <div>{'lastPage-' + lastPage}</div>
      <div>{'variant-name-' + data[0]?.nodes[0].name}</div>
      <div>{'variant-subname-' + data[0]?.nodes[0].subname}</div>
      <div>{'variant-numRuns-' + data[0]?.nodes[0].metrics.get(Resources.MetricType.NUM_RUNS)}</div>
      <div>{'variant-numFailures-' + data[0]?.nodes[0].metrics.get(Resources.MetricType.NUM_FAILURES)}</div>
      <div>{'variant-avgRuntime-' + data[0]?.nodes[0].metrics.get(Resources.MetricType.AVG_RUNTIME)}</div>
      <div>{'variant-totalRuntime-' + data[0]?.nodes[0].metrics.get(Resources.MetricType.TOTAL_RUNTIME)}</div>
      <div>{'variant-avgCores-' + data[0]?.nodes[0].metrics.get(Resources.MetricType.AVG_CORES)}</div>
      <div>{'isLoading-' + isLoading}</div>
      <Button data-testid='buttonPage' onClick={() => api.updatePage(20)}>{'paramsPage-' + params.page}</Button>
      <Button data-testid='buttonRowsPerPage' onClick={() => api.updateRowsPerPage(20)}>{'paramsRowsPerPage-' + params.rowsPerPage}</Button>
      <Button data-testid='buttonFilter' onClick={() => api.updateFilter('newFilt')}>{'paramsFilt-' + params.filter}</Button>
      <Button data-testid='buttonDate' onClick={() => api.updateDate(new Date())}>{'paramsDate-' + params.date}</Button>
      <Button data-testid='buttonPeriod' onClick={() => api.updatePeriod(Resources.Period.WEEK)}>{'paramsPeriod-' + params.period}</Button>
      <Button data-testid='buttonSort' onClick={() => api.updateSort(Resources.SortType.SORT_NUM_RUNS)}>{'paramsSort-' + params.sort}</Button>
      <Button data-testid='buttonAscending' onClick={() => api.updateAscending(false)}>{'paramsAscending-' + params.ascending}</Button>
    </>
  );
};

const mockMetricsWithData: Map<string, Resources.TestMetricsArray> =
  new Map<string, Resources.TestMetricsArray>(
      Object.entries(
          {
            '2012-01-02': {
              'data': [
                {
                  metricType: 'NUM_RUNS' as Resources.MetricType,
                  metricValue: 2,
                },
                {
                  metricType: 'NUM_FAILURES' as Resources.MetricType,
                  metricValue: 3,
                },
                {
                  metricType: 'AVG_RUNTIME' as Resources.MetricType,
                  metricValue: 4,
                },
                {
                  metricType: 'TOTAL_RUNTIME' as Resources.MetricType,
                  metricValue: 5,
                },
                {
                  metricType: 'AVG_CORES' as Resources.MetricType,
                  metricValue: 6,
                },
              ],
            },
          },
      ),
  );

describe('<MetricsContext />', () => {
  it('populate tests/lastPage correctly', async () => {
    jest.spyOn(Resources, 'fetchTestMetrics').mockResolvedValue({
      tests: [
        {
          testId: '12',
          testName: 'A',
          fileName: 'A',
          metrics: mockMetricsWithData,
          variants: [
            {
              suite: 'suite',
              builder: 'builder',
              metrics: mockMetricsWithData,
            },
          ],
        },
      ],
      lastPage: true,
    });
    await act(async () => {
      render(
          <MetricsContextProvider>
            <TestingComponent/>
          </MetricsContextProvider>,
      );
    });
    expect(screen.getByText('id-12')).toBeInTheDocument();
    expect(screen.getByText('name-A')).toBeInTheDocument();
    expect(screen.getByText('numRuns-2')).toBeInTheDocument();
    expect(screen.getByText('numFailures-3')).toBeInTheDocument();
    expect(screen.getByText('avgRuntime-4')).toBeInTheDocument();
    expect(screen.getByText('totalRuntime-5')).toBeInTheDocument();
    expect(screen.getByText('avgCores-6')).toBeInTheDocument();
    expect(screen.getByText('lastPage-true')).toBeInTheDocument();
    expect(screen.getByText('variant-name-builder')).toBeInTheDocument();
    expect(screen.getByText('variant-subname-suite')).toBeInTheDocument();
    expect(screen.getByText('variant-numRuns-2')).toBeInTheDocument();
    expect(screen.getByText('variant-numFailures-3')).toBeInTheDocument();
    expect(screen.getByText('variant-avgRuntime-4')).toBeInTheDocument();
    expect(screen.getByText('variant-totalRuntime-5')).toBeInTheDocument();
    expect(screen.getByText('variant-avgCores-6')).toBeInTheDocument();
    expect(screen.getByText('isLoading-false')).toBeInTheDocument();
  });
  it('update params accordingly when api is called', async () => {
    jest.spyOn(Resources, 'fetchTestMetrics').mockResolvedValue({
      tests: [
        {
          testId: '12',
          testName: 'A',
          fileName: 'A',
          metrics: mockMetricsWithData,
          variants: [
            {
              suite: 'suite',
              builder: 'builder',
              metrics: mockMetricsWithData,
            },
          ],
        },
      ],
      lastPage: true,
    });
    await act(async () => {
      render(
          <MetricsContextProvider>
            <TestingComponent/>
          </MetricsContextProvider>,
      );
    });
    await act(async () => {
      fireEvent.click(screen.getByTestId('buttonRowsPerPage'));
      fireEvent.click(screen.getByTestId('buttonFilter'));
      fireEvent.click(screen.getByTestId('buttonDate'));
      fireEvent.click(screen.getByTestId('buttonPeriod'));
      fireEvent.click(screen.getByTestId('buttonSort'));
      fireEvent.click(screen.getByTestId('buttonAscending'));
      fireEvent.click(screen.getByTestId('buttonPage'));
    });
    expect(screen.getByText('paramsPage-20')).toBeInTheDocument();
    expect(screen.getByText('paramsRowsPerPage-20')).toBeInTheDocument();
    expect(screen.getByText('paramsFilt-newFilt')).toBeInTheDocument();
    expect(screen.getByText('paramsPeriod-1')).toBeInTheDocument();
    expect(screen.getByText('paramsSort-1')).toBeInTheDocument();
    expect(screen.getByText('paramsAscending-false')).toBeInTheDocument();
  });
});
