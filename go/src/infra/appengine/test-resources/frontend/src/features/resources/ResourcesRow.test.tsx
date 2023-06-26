// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { fireEvent, render } from '@testing-library/react';
import { MetricType } from '../../api/resources';
import { Api, Test } from '../context/MetricsContext';
import ResourcesRow, { ResourcesRowProps, displayMetrics } from './ResourcesRow';

// Mock api. We will just test if they are called.
const mockApi : Api = {
  nextPage: () => {
    // do nothing.
  },
  prevPage: () => {
    // do nothing.
  },
  firstPage: () => {
    // do nothing.
  },
};

const mockMetrics: Map<MetricType, number> = new Map<MetricType, number>(
    [
      [MetricType.NUM_RUNS, 1],
      [MetricType.NUM_FAILURES, 2],
      [MetricType.AVG_RUNTIME, 3],
      [MetricType.TOTAL_RUNTIME, 4],
      [MetricType.AVG_CORES, 5],
    ],
);

const test: Test = {
  testId: 'testId',
  testName: 'testName',
  fileName: 'fileName',
  metrics: mockMetrics,
  variants: [
    {
      suite: 'suite',
      builder: 'builder',
      metrics: mockMetrics,
    },
    {
      suite: 'suite',
      builder: 'builder',
      metrics: mockMetrics,
    },
  ],
};

const resoucresRowProps: ResourcesRowProps = {
  test: test,
  lastPage: false,
  api: mockApi,
};

describe('when rendering the ResourcesRow', () => {
  it('should render the TableRow and VariantRow', () => {
    const { getByTestId, getAllByTestId } = render(
        <table>
          <tbody>
            <ResourcesRow {...resoucresRowProps}/>
          </tbody>
        </table>,
    );
    const button = getByTestId('clickButton');
    fireEvent.click(button);
    const tableRow = getByTestId('tableRowTest');
    const variantRow = getAllByTestId('variantRowTest');
    expect(variantRow).toHaveLength(2);
    expect(tableRow).toBeInTheDocument();
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
