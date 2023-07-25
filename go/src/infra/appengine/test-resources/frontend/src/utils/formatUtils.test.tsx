// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { formatDate, formatNumber, formatTime } from './formatUtils';

// Test formatTime
test.each([
  ['987654', '11d 10h'],
  ['10000', '2h 46m'],
  ['500', '8m 20s'],
  ['0', '0.0s'],
  ['1', '1.0s'],
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
