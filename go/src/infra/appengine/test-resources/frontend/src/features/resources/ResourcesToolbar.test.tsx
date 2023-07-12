/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { screen } from '@testing-library/react';
import { renderWithContext } from '../../utils/testUtils';
import { Params } from '../context/MetricsContext';
import { Period, SortType } from '../../api/resources';
import ResourcesToolbar from './ResourcesToolbar';

const params: Params = {
  page: 12,
  rowsPerPage: 25,
  filter: 'filter',
  date: new Date(),
  period: Period.DAY,
  sort: SortType.SORT_NAME,
  ascending: true,
  timelineView: false,
};

describe('when rendering the ResourcesTableToolbar', () => {
  it('should render toolbar elements', () => {
    renderWithContext(<ResourcesToolbar/>, { params: params, lastPage: false });
    expect(screen.getByTestId('CalendarIcon')).toBeInTheDocument();
    expect(screen.getByRole('textbox', {
      name: /date/i,
    })).toBeInTheDocument();
    expect(screen.getByTestId('formControlTest')).toBeInTheDocument();
    expect(screen.getByTestId('textFieldTest')).toBeInTheDocument();
    expect(screen.getByTestId('timelineViewToggle')).toBeInTheDocument();
  });
});
