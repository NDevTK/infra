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

const mockMetricsWithDataTimeline: Map<string, Resources.TestMetricsArray> =
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
            '2012-01-03': {
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
            '2012-01-04': {
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
            '2012-01-05': {
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
            '2012-01-06': {
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

describe('MetricsContext rendering for test snapshot', () => {
  it('populate tests correctly', async () => {
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
    await contextRender((value) => {
      return (
        <>
          <div>id-{value.data[0]?.id}</div>
          <div>name-{value.data[0]?.name}</div>
          <div>numRuns-{value.data[0]?.metrics.get('2012-01-02')?.get(Resources.MetricType.NUM_RUNS)}</div>
          <div>numFailures-{value.data[0]?.metrics.get('2012-01-02')?.get(Resources.MetricType.NUM_FAILURES)}</div>
          <div>avgRuntime-{value.data[0]?.metrics.get('2012-01-02')?.get(Resources.MetricType.AVG_RUNTIME)}</div>
          <div>totalRuntime-{value.data[0]?.metrics.get('2012-01-02')?.get(Resources.MetricType.TOTAL_RUNTIME)}</div>
          <div>avgCores-{value.data[0]?.metrics.get('2012-01-02')?.get(Resources.MetricType.AVG_CORES)}</div>
          <div>variant-name-{value.data[0]?.nodes[0].name}</div>
          <div>variant-subname-{value.data[0]?.nodes[0].subname}</div>
          <div>variant-numRuns-{value.data[0]?.nodes[0].metrics.get('2012-01-02')?.get(Resources.MetricType.NUM_RUNS)}</div>
          <div>variant-numFailures-{value.data[0]?.nodes[0].metrics.get('2012-01-02')?.get(Resources.MetricType.NUM_FAILURES)}</div>
          <div>variant-avgRuntime-{value.data[0]?.nodes[0].metrics.get('2012-01-02')?.get(Resources.MetricType.AVG_RUNTIME)}</div>
          <div>variant-totalRuntime-{value.data[0]?.nodes[0].metrics.get('2012-01-02')?.get(Resources.MetricType.TOTAL_RUNTIME)}</div>
          <div>variant-avgCores-{value.data[0]?.nodes[0].metrics.get('2012-01-02')?.get(Resources.MetricType.AVG_CORES)}</div>
        </>
      );
    });
    expect(screen.getByText('id-12')).toBeInTheDocument();
    expect(screen.getByText('name-A')).toBeInTheDocument();
    expect(screen.getByText('numRuns-2')).toBeInTheDocument();
    expect(screen.getByText('numFailures-3')).toBeInTheDocument();
    expect(screen.getByText('avgRuntime-4')).toBeInTheDocument();
    expect(screen.getByText('totalRuntime-5')).toBeInTheDocument();
    expect(screen.getByText('avgCores-6')).toBeInTheDocument();
    expect(screen.getByText('variant-name-builder')).toBeInTheDocument();
    expect(screen.getByText('variant-subname-suite')).toBeInTheDocument();
    expect(screen.getByText('variant-numRuns-2')).toBeInTheDocument();
    expect(screen.getByText('variant-numFailures-3')).toBeInTheDocument();
    expect(screen.getByText('variant-avgRuntime-4')).toBeInTheDocument();
    expect(screen.getByText('variant-totalRuntime-5')).toBeInTheDocument();
    expect(screen.getByText('variant-avgCores-6')).toBeInTheDocument();
  });
});

describe('MetricsContext rendering for test timeline', () => {
  it('populate tests correctly', async () => {
    jest.spyOn(Resources, 'fetchTestMetrics').mockResolvedValue({
      tests: [
        {
          testId: '12',
          testName: 'A',
          fileName: 'A',
          metrics: mockMetricsWithDataTimeline,
          variants: [
            {
              suite: 'suite',
              builder: 'builder',
              metrics: mockMetricsWithDataTimeline,
            },
          ],
        },
      ],
      lastPage: true,
    });
    await contextRender((value) => (
      <>
        <div>id-{value.data[0]?.metrics.get('2012-01-06')?.get(Resources.MetricType.AVG_RUNTIME)}</div>
        <div>datesWithDataSize-{value.datesToShow.length}</div>
      </>
    ), { props: { timelineView: true } });
    expect(screen.getByText('id-4')).toBeInTheDocument();
    expect(screen.getByText('datesWithDataSize-5')).toBeInTheDocument();
  });
});

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
