// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

export const EMPTY_VALUE = '-';

export enum SummaryStyle {
  DEFAULT = 'default',

  // The style of a summary line that has been tagged as highlighted by the
  // comparative analysis.
  HIGHLIGHTED = 'highlighted',

  // The style of a summary line that the comparative analysis failed to find
  // any references for.
  GREYED = 'greyed',
}

/**
 * Logs table sort order.
 */
export enum SortOrder {
  ASC = 'asc',
  DESC = 'desc',
}
