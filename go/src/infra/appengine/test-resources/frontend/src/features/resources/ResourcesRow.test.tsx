// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import MockMetrics from '../../utils/MockMetrics.json';
import { MetricType, TestMetricsArray } from '../../api/resources';
import ResourcesRow from './ResourcesRow';

const testDateMetricData = {
  test_id: MockMetrics[0].test_id,
  test_name: MockMetrics[0].test_name,
  file_name: MockMetrics[0].file_name,
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
  variants: Array(1),
};

describe('when rendering the ResourcesRow', () => {
  it('should render the TableRow', () => {
    const { getByTestId } = render(
        <table>
          <tbody>
            <ResourcesRow {...testDateMetricData}/>
          </tbody>
        </table>,
    );
    const tableRow = getByTestId('tableRowTest');
    expect(tableRow).toBeInTheDocument();
  });
});
