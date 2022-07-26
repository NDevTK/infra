// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { styled } from '@mui/system';

import Table from '@mui/material/Table';
import { tableCellClasses } from '@mui/material/TableCell';

export const PlainTable = styled(Table)({
  [`& .${tableCellClasses.head}`]: {
    fontSize: '1rem',
    fontWeight: 'normal',
    color: 'dimgray',
    opacity: '80%',
    border: 'none',
    padding: 0,
  },
  [`& .${tableCellClasses.body}`]: {
    fontSize: '1rem',
    border: 'none',
    padding: 0,
  },
});
