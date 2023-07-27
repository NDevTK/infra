// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';

export function formatTime(seconds: number) {
  let result = '';

  const day = Math.floor(seconds / 86400);
  if (day >= 1) {
    result += `${day.toFixed(0)}d`;
  }

  const hour = Math.floor(seconds % 86400 / 3600);
  if (hour >= 1) {
    result += ` ${hour.toFixed(0)}h`;
  }

  const min = Math.floor(seconds % 3600 / 60);
  if (day < 1 && min >= 1) {
    result += ` ${min.toFixed(0)}m`;
  }

  const sec = Math.floor(seconds % 60);
  if (hour < 1) {
    result += ` ${sec.toPrecision(2)}s`;
  }

  return result.trimStart();
}

export function formatNumber(num: number) {
  return Number.isNaN(num) ? '-' : new Intl.NumberFormat().format(num);
}

export function formatDate(date : Date) {
  return dayjs(date).format('YYYY-MM-DD');
}
