// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { formatDate, formatNumber, formatTime } from './formatUtils';

const M = 60;
const H = 60 * M;
const D = 24 * H;

// Test formatTime
test.each([
  [(11 * D + 10 * H + 5 * M + 1.2).toString(), '11d 10h'],
  [(1 * D + 2 * H + 3 * M + 4), '1d 2h'],
  [(1 * D + 3 * M + 4), '1d'],
  [(2 * H + 46 * M + 5).toString(), '2h 46m'],
  [(8 * M + 20).toString(), '8m 20s'],
  [(M + 2.2).toString(), '1m 2s'],
  ['0', '0s'],
  ['1', '1s'],
  ['6.5555', '6.56s'],
  ['.056', '0.06s'],
  ['.0000566', '<0.01s'],
  [Number.NaN, '-'],
])('.formatTime(%p, %p)', (seconds, expected) => {
  expect(formatTime(Number(seconds))).toBe(expected);
});

// Test formatNumber
test.each([
  ['NaN', '-'],
  ['10000', '10,000'],
  ['500', '500'],
  ['0', '0'],
  ['1', '1'],
  ['2.2222', '2.22'],
  ['2.2', '2.20'],
  ['.00052', '<0.01'],
])('.formatNumber(%p, %p)', (num, expected) => {
  expect(formatNumber(Number(num))).toBe(expected);
});

// Test formatDate
test.each([
  // 2023-01-01 creates a date at GMT, whereas a timestamp is in current TZ
  ['2023-01-01T00:00:00', '2023-01-01'],
])('.formatDate(%p, %p)', (date, expected) => {
  expect(formatDate(new Date(date))).toBe(expected);
});
