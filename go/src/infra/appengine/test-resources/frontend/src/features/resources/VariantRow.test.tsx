// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { MetricType, TestMetricsArray, TestVariantData } from '../../api/resources';
import VariantRow from './VariantRow';

const variantRowProps: TestVariantData = {
  metrics: new Map<string, TestMetricsArray>(
      Object.entries(
          {
            '01-02-2012': {
              'data': [
                {
                  metric_type: MetricType.NUM_RUNS,
                  metric_value: 2,
                },
              ],
            },
          },
      ),
  ),
  builder: 'builder',
  suite: 'suite',
};

describe('when rendering the VariantRow', () => {
  it('should render the TableCell with variantRowCellTest id', () => {
    const { getByTestId } = render(
        <table>
          <tbody>
            <tr>
              <VariantRow {...variantRowProps}/>
            </tr>
          </tbody>
        </table>,
    );
    const variantCellRow = getByTestId('variantRowCellTest');
    expect(variantCellRow).toBeInTheDocument();
  });
});
