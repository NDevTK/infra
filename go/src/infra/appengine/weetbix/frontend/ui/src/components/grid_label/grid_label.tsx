// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';

interface Props {
    text?: string;
    children?: React.ReactNode,
    xs?: number;
    lg?: number;
    testid?: string;
}

const GridLabel = ({
  text,
  children,
  xs = 2,
  lg = xs,
  testid,
}: Props) => {
  return (
    <Grid item xs={xs} lg={lg} data-testid={testid}>
      <Box
        sx={{
          display: 'inline-block',
          wordBreak: 'break-all',
          overflowWrap: 'break-word',
        }}
        paddingTop={1}>
        {text}
      </Box>
      {children}
    </Grid>
  );
};

export default GridLabel;
