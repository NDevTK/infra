// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { fireEvent, render } from '@testing-library/react';
import { MetricType } from '../../api/resources';
import { Test } from '../context/MetricsContext';
import ResourcesRow, { displayMetrics } from './ResourcesRow';

const mockMetrics: Map<MetricType, number> = new Map<MetricType, number>(
    [
      [MetricType.NUM_RUNS, 1],
      [MetricType.NUM_FAILURES, 2],
      [MetricType.AVG_RUNTIME, 3],
      [MetricType.TOTAL_RUNTIME, 4],
      [MetricType.AVG_CORES, 5],
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

describe('when calling formatMetrics', () => {
  it('should return 5 table cells for snapshot metrics', () => {
    const formattedMetrics = render(
        <table>
          <tbody>
            <tr>
              { displayMetrics(mockMetrics) }
            </tr>
          </tbody>
        </table>,
    );
    expect(formattedMetrics.getAllByTestId('tableCell')).toHaveLength(5);
  });
});
