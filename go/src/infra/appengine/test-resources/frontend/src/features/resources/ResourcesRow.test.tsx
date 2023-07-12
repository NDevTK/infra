// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { fireEvent, render, screen } from '@testing-library/react';
import { act } from 'react-dom/test-utils';
import { MetricType } from '../../api/resources';
import { MetricsContextProvider, MetricsContextValue, Test, MetricsContext } from '../context/MetricsContext';
import * as Resources from '../../api/resources';
import ResourcesRow from './ResourcesRow';


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

const mockMetricTypeToNum: Map<MetricType, number> = new Map<MetricType, number>(
    [
      [MetricType.NUM_RUNS, 1],
      [MetricType.NUM_FAILURES, 2],
      [MetricType.AVG_RUNTIME, 3],
      [MetricType.TOTAL_RUNTIME, 4],
      [MetricType.AVG_CORES, 5],
    ],
);

const mockMetrics: Map<string, Map<MetricType, number>> = new Map<string, Map<MetricType, number>>(
    [
      ['2023-01-01', mockMetricTypeToNum],
      ['2023-01-02', mockMetricTypeToNum],
      ['2023-01-03', mockMetricTypeToNum],
    ],
);


describe('when rendering the ResourcesRow', () => {
  it('should render a single row', () => {
    const test: Test = {
      id: 'testId',
      name: 'testName',
      fileName: 'fileName',
      metrics: mockMetrics,
      isLeaf: true,
      nodes: [],
    };

    const { getByTestId } = render(
        <table>
          <tbody>
            <ResourcesRow data={test} depth={0}/>
          </tbody>
        </table>,
    );
    const tableRow = getByTestId('tablerow-testId');
    expect(tableRow).toBeInTheDocument();
  });
  it('should render expandable rows', () => {
    const test: Test = {
      id: 'testId',
      name: 'testName',
      fileName: 'fileName',
      metrics: mockMetrics,
      isLeaf: false,
      nodes: [
        {
          id: 'v1',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isLeaf: true,
          nodes: [],
        },
        {
          id: 'v2',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isLeaf: true,
          nodes: [],
        },
      ],
    };

    const { getByTestId } = render(
        <table>
          <tbody>
            <ResourcesRow data={test} depth={0}/>
          </tbody>
        </table>,
    );
    const testRow = getByTestId('tablerow-testId');
    expect(testRow).toBeInTheDocument();
    expect(testRow.getAttribute('data-depth')).toEqual('0');

    const button = getByTestId('clickButton-testId');
    fireEvent.click(button);

    const v1Row = getByTestId('tablerow-v1');
    expect(v1Row).toBeInTheDocument();
    expect(v1Row.getAttribute('data-depth')).toEqual('1');

    expect(getByTestId('tablerow-v2')).toBeInTheDocument();
  });
});

describe('when rendering ResourcesRow', () => {
  beforeEach(() => {
    jest.spyOn(Resources, 'fetchTestMetrics').mockResolvedValue({
      tests: [],
      lastPage: true,
    });
  });
  it('should render test snapshot view properly', async () => {
    const test: Test = {
      id: 'testId',
      name: 'testName',
      fileName: 'fileName',
      metrics: mockMetrics,
      isLeaf: false,
      nodes: [
        {
          id: 'v1',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isLeaf: true,
          nodes: [],
        },
        {
          id: 'v2',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isLeaf: true,
          nodes: [],
        },
      ],
    };
    await contextRender(() => {
      return (
        <>
          <table>
            <tbody>
              <ResourcesRow data={test} depth={0}/>
            </tbody>
          </table>,
        </>

      );
    });
    expect(screen.getAllByTestId('tableCell')).toHaveLength(5);
  });
  it('should render test timeline view properly', async () => {
    const test: Test = {
      id: 'testId',
      name: 'testName',
      fileName: 'fileName',
      metrics: mockMetrics,
      isLeaf: false,
      nodes: [
        {
          id: 'v1',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isLeaf: true,
          nodes: [],
        },
        {
          id: 'v2',
          name: 'suite',
          subname: 'builder',
          metrics: mockMetrics,
          isLeaf: true,
          nodes: [],
        },
      ],
    };
    await contextRender(() => (
      <>
        <table>
          <tbody>
            <ResourcesRow data={test} depth={0}/>
          </tbody>
        </table>,
      </>
    ), { props: { timelineView: true } });
    expect(screen.getAllByTestId('timelineTest')).toHaveLength(5);
  });
});
