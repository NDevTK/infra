// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';

export function formatTime(seconds: number) {
  let result = '';

  const hour = seconds / 3600;
  if (hour >= 1) {
    result += `${hour.toFixed(0)}h`;
  }

  const min = seconds % 3600 / 60;
  if (min >= 1) {
    result += ` ${min.toFixed(0)}m`;
  }

  const sec = seconds % 60;
  result += ` ${sec.toPrecision(2)}s`;

  return result.trimStart();
}

export function formatNumber(num: number) {
  return new Intl.NumberFormat().format(num);
}

export function formatDate(date : Date) {
  return dayjs(date).format('YYYY-MM-DD');
}
