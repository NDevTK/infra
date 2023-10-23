// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';

// If days >= 1 don't show minutes and seconds
// If minutes >= 1 don't show decimal in seconds
// If decimal show up to 2 points of precision.
// If <0.01 show '<0.01s'
export function formatTime(seconds: number) {
  if (Number.isNaN(seconds)) {
    return '-';
  }
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

  const sec = seconds % 60;
  if (hour < 1 && day < 1) {
    if (sec === 0) {
      result += '0';
    } else if (sec < 0.01) {
      result += ' <0.01';
    } else if (sec % 1 != 0 && min < 1) {
      result += ` ${sec.toFixed(2)}`;
    } else {
      result += ` ${sec.toFixed(0)}`;
    }
    result += 's';
  }

  return result.trimStart();
}

// Return '-' if not a number
// If decimal places return a precision of exactly 2
// All other whole numbers should be returned as is
export function formatNumber(num: number): string {
  if (Number.isNaN(num)) {
    return '-';
  }
  if (num % 1 !== 0) {
    if (num < 0.01) {
      return '<0.01';
    }
    return new Intl.NumberFormat('en', { maximumFractionDigits: 2, minimumFractionDigits: 2 }).format(num);
  }
  return new Intl.NumberFormat().format(num);
}

export function formatDate(date : Date) {
  return dayjs(date).format('YYYY-MM-DD');
}
