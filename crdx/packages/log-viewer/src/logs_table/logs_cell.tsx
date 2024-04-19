// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { styled } from '@mui/material';
import { grey, red } from '@mui/material/colors';

export const LogsEntryTableCell = styled('td')(({ theme }) => ({
  verticalAlign: 'top',
  padding: 0,
  textAlign: 'left',
  boxSizing: 'content-box',
  fontSize: '12px',
  fontFamily: 'monospace',
  '.line': {
    width: '30px',
    height: '16px',
    display: 'inline-block',
  },
  '.severity': {
    width: '14px',
    height: '16px',
    textAlign: 'center',
    display: 'inline-block',
  },
  '.severity.verbose': {
    backgroundColor: theme.palette.mode === 'light' ? grey[400] : grey[800],
  },
  '.severity.debug': {
    backgroundColor:
      theme.palette.mode === 'light'
        ? theme.palette.success.light
        : theme.palette.success.dark,
  },
  '.severity.info': {
    backgroundColor:
      theme.palette.mode === 'light'
        ? theme.palette.info.light
        : theme.palette.info.dark,
  },
  '.severity.notice': {
    backgroundColor:
      theme.palette.mode === 'light'
        ? theme.palette.warning.light
        : theme.palette.warning.dark,
  },
  '.severity.warning': {
    backgroundColor: theme.palette.warning.main,
    color: 'white',
  },
  '.severity.error': {
    backgroundColor:
      theme.palette.mode === 'light'
        ? theme.palette.error.light
        : theme.palette.error.dark,
    color: theme.palette.mode === 'light' ? 'white' : 'black',
  },
  '.severity.fatal': {
    backgroundColor: theme.palette.error.main,
    color: 'white',
  },
  '.summary': {
    whiteSpace: 'pre-wrap',
    display: 'inline-block',
  },
  '.summary.highlighted': {
    color: theme.palette.mode === 'light' ? red[700] : red[200],
    fontWeight: 'bold',
  },
  '.summary.greyed': {
    color: theme.palette.mode === 'light' ? grey[700] : grey[200],
  },
  '.summary.limit': {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    display: '-webkit-box',
    lineClamp: 5,
    WebkitLineClamp: 5,
    WebkitBoxOrient: 'vertical',
  },
}));
