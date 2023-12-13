// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useContext } from 'react';
import { LineChart } from '@mui/x-charts/LineChart';
import { Box } from '@mui/material';
import { TrendsContext } from './TrendsContext';

function TrendsChart() {
  const { data, isLoading, isConfigLoaded } = useContext(TrendsContext);

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
              data: data.map((i) => new Date(i.date)),
              scaleType: 'time',
            }]}
            series={[
              {
                data: data.map((i) => {
                  const linesCov = parseInt(i.covered + '');
                  const totalLines = parseInt(i.total + '');
                  return (linesCov*100)/totalLines;
                }),
                area: true,
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
