/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { render, screen } from '@testing-library/react';
import { useContext } from 'react';
import { act } from 'react-dom/test-utils';
import * as Resources from '../../api/resources';
import { MetricsContext, MetricsContextProvider } from './MetricsContext';

const TestingComponent = () => {
  const { tests, lastPage } = useContext(MetricsContext);
  return (
    <>
      <div>{'id-' + tests[0]?.testId}</div>
      <div>{'testName-' + tests[0]?.testName}</div>
      <div>{'fileName-' + tests[0]?.fileName}</div>
      <div>{'numRuns-' + tests[0]?.metrics.get(Resources.MetricType.NUM_RUNS)}</div>
      <div>{'numFailures-' + tests[0]?.metrics.get(Resources.MetricType.NUM_FAILURES)}</div>
      <div>{'avgRuntime-' + tests[0]?.metrics.get(Resources.MetricType.AVG_RUNTIME)}</div>
      <div>{'totalRuntime-' + tests[0]?.metrics.get(Resources.MetricType.TOTAL_RUNTIME)}</div>
      <div>{'avgCores-' + tests[0]?.metrics.get(Resources.MetricType.AVG_CORES)}</div>
      <div>{'lastPage-' + lastPage}</div>
      <div>{'suite-' + tests[0]?.variants[0].suite}</div>
      <div>{'builder-' + tests[0]?.variants[0].builder}</div>
      <div>{'variant-numRuns-' + tests[0]?.variants[0].metrics.get(Resources.MetricType.NUM_RUNS)}</div>
      <div>{'variant-numFailures-' + tests[0]?.variants[0].metrics.get(Resources.MetricType.NUM_FAILURES)}</div>
      <div>{'variant-avgRuntime-' + tests[0]?.variants[0].metrics.get(Resources.MetricType.AVG_RUNTIME)}</div>
      <div>{'variant-totalRuntime-' + tests[0]?.variants[0].metrics.get(Resources.MetricType.TOTAL_RUNTIME)}</div>
      <div>{'variant-avgCores-' + tests[0]?.variants[0].metrics.get(Resources.MetricType.AVG_CORES)}</div>
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
      last_page: true,
    });
    await act(async () => {
      render(
          <MetricsContextProvider>
            <TestingComponent/>
          </MetricsContextProvider>,
      );
    });
    expect(screen.getByText('id-12')).toBeInTheDocument();
    expect(screen.getByText('testName-A')).toBeInTheDocument();
    expect(screen.getByText('fileName-A')).toBeInTheDocument();
    expect(screen.getByText('numRuns-2')).toBeInTheDocument();
    expect(screen.getByText('numFailures-3')).toBeInTheDocument();
    expect(screen.getByText('avgRuntime-4')).toBeInTheDocument();
    expect(screen.getByText('totalRuntime-5')).toBeInTheDocument();
    expect(screen.getByText('avgCores-6')).toBeInTheDocument();
    expect(screen.getByText('lastPage-true')).toBeInTheDocument();
    expect(screen.getByText('suite-suite')).toBeInTheDocument();
    expect(screen.getByText('builder-builder')).toBeInTheDocument();
    expect(screen.getByText('variant-numRuns-2')).toBeInTheDocument();
    expect(screen.getByText('variant-numFailures-3')).toBeInTheDocument();
    expect(screen.getByText('variant-avgRuntime-4')).toBeInTheDocument();
    expect(screen.getByText('variant-totalRuntime-5')).toBeInTheDocument();
    expect(screen.getByText('variant-avgCores-6')).toBeInTheDocument();
  });
});
