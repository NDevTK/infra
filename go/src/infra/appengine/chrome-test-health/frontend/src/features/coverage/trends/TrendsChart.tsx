// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useContext, useState, useEffect } from 'react';
import { LineChart } from '@mui/x-charts/LineChart';
import { Box } from '@mui/material';
import { CoverageTrend } from '../../../api/coverage';
import { TrendsContext } from './TrendsContext';

function TrendsChart() {
  const { data, isLoading, isConfigLoaded } = useContext(TrendsContext);
  const [chartsData, setChartsData] = useState([] as CoverageTrend[]);

  useEffect(() => {
    const dataArr = [...data];
    dataArr.sort((a, b) => {
      const d1 = new Date(a.date);
      const d2 = new Date(b.date);

      if (d1 == d2) return 0;

      return (d1<d2) ? -1 : 1;
    });
    setChartsData(dataArr);
  }, [data]);

  const renderLoading = () => {
    if (isConfigLoaded && isLoading) {
      return (
        <Box data-testid="loading">Loading...</Box>
      );
    }
    return;
  };

  return (
    <>
      {
        data.length > 0 ?
        <Box data-testid="trendsChart">
          <LineChart
            xAxis={[{
              data: chartsData.map((i) => new Date(i.date)),
              scaleType: 'time',
            }]}
            series={[
              {
                data: chartsData.map((i) => {
                  const linesCov = parseInt(i.covered + '');
                  const totalLines = parseInt(i.total + '');
                  return (linesCov*100)/totalLines;
                }),
                area: true,
                color: '#1976d2',
              },
            ]}
            width={1000}
            height={600}

          />
        </Box>:
        renderLoading()
      }
    </>
  );
}

export default TrendsChart;
