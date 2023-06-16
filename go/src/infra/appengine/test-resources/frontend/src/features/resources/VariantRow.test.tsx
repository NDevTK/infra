// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { MetricType } from '../../api/resources';
import VariantRow from './VariantRow';

const mockMetrics: Map<MetricType, number> = new Map<MetricType, number>(
    [
      [MetricType.NUM_RUNS, 1],
      [MetricType.NUM_FAILURES, 2],
      [MetricType.AVG_RUNTIME, 3],
      [MetricType.TOTAL_RUNTIME, 4],
      [MetricType.AVG_CORES, 5],
    ],
);

const mockTestVariant = {
  suite: 'suite',
  builder: 'builder',
  metrics: mockMetrics,
};


describe('when rendering the VariantRow', () => {
  it('should render the TableCell with variantRowCellTest id', () => {
    const { getByTestId } = render(
        <table>
          <tbody>
            <VariantRow {...{ variant: mockTestVariant, tableKey: 1 }}/>
          </tbody>
        </table>,
    );
    const variantCellRow = getByTestId('variantRowCellTest');
    expect(variantCellRow).toBeInTheDocument();
  });
});
