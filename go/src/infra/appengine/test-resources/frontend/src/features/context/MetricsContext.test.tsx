/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { fireEvent, render, screen } from '@testing-library/react';
import { act } from 'react-dom/test-utils';
import { Button } from '@mui/material';
import * as Resources from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';
import { MetricsContext, MetricsContextProvider, MetricsContextValue } from './MetricsContext';

async function contextRender(ui: (value: MetricsContextValue) => React.ReactElement, { props } = { props: {} }) {
  await act(async () => {
    render(
        <MetricsContextProvider {... props}>
          <MetricsContext.Consumer>
            {(value) => ui(value)}
          </MetricsContext.Consumer>
        </MetricsContextProvider>,
    );
  },
  );
}

describe('MetricsContext params', () => {
  beforeEach(() => {
    jest.spyOn(Resources, 'fetchTestMetrics').mockResolvedValue({
      tests: [],
      lastPage: true,
    });
  });

  it('page', async () => {
    await contextRender((value) => (
      <Button data-testid='updatePage' onClick={() => value.api.updatePage(20)}>{'page-' + value.params.page}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePage'));
    });
    expect(screen.getByText('page-20')).toBeInTheDocument();
  });

  it('filter', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updateFilter' onClick={() => value.api.updateFilter('filt')}>{'filter-' + value.params.filter}</Button>
        <div>page-{value.params.page}</div>
      </>
    ), { props: { page: 1 } });
    expect(screen.getByText('page-1')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateFilter'));
    });
    expect(screen.getByText('filter-filt')).toBeInTheDocument();
    expect(screen.getByText('page-0')).toBeInTheDocument();
  });

  it('updateDate snapshot view', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updateDate' onClick={() => value.api.updateDate(new Date('2023-01-02'))}>{'date-' + formatDate(value.params.date)}</Button>
        <div>page-{value.params.page}</div>
        <div>sortIndex-{value.params.sortIndex}</div>
      </>
    ), { props: { page: 1, date: new Date('2023-01-01'), timelineView: false, sortIndex: 4 } });
    expect(screen.getByText('page-1')).toBeInTheDocument();
    expect(screen.getByText('date-2023-01-01')).toBeInTheDocument();
    expect(screen.getByText('sortIndex-4')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateDate'));
    });
    expect(screen.getByText('page-0')).toBeInTheDocument();
    expect(screen.getByText('date-2023-01-02')).toBeInTheDocument();
    expect(screen.getByText('sortIndex-0')).toBeInTheDocument();
  });

  it('updateDate timeline view', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updateDate' onClick={() => value.api.updateDate(new Date('2023-01-02'))}/>
        <div>sortIndex-{value.params.sortIndex}</div>
      </>
    ), { props: { page: 1, date: new Date('2023-01-01'), timelineView: true, sortIndex: 0 } });
    expect(screen.getByText('sortIndex-0')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateDate'));
    });
    expect(screen.getByText('sortIndex-4')).toBeInTheDocument();
  });

  it('updateTimelineView ', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updateTimeline' onClick={() => value.api.updateTimelineView(false)}/>
        <div>sortIndex-{value.params.sortIndex}</div>
        <div>timelineView-{String(value.params.timelineView)}</div>
      </>
    ), { props: { timelineView: true, sortIndex: 4 } });
    expect(screen.getByText('sortIndex-4')).toBeInTheDocument();
    expect(screen.getByText('timelineView-true')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateTimeline'));
    });
    expect(screen.getByText('sortIndex-0')).toBeInTheDocument();
    expect(screen.getByText('timelineView-false')).toBeInTheDocument();
  });

  it('updatePeriod', async () => {
    await contextRender((value) => (
      <>
        <Button data-testid='updatePeriodToWeek' onClick={() => value.api.updatePeriod(Resources.Period.WEEK)}>{'period-' + value.params.period}</Button>
        <Button data-testid='updatePeriodToDay' onClick={() => value.api.updatePeriod(Resources.Period.DAY)}/>
        <div>date-{formatDate(value.params.date)}</div>
        <div>page-{value.params.page}</div>
      </>
    ), {props: { date: new Date('2023-07-19'), page: 10 }});
    expect(screen.getByText('period-1')).toBeInTheDocument();
    expect(screen.getByText('page-10')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePeriodToDay'));
    });
    expect(screen.getByText('period-0')).toBeInTheDocument();
    expect(screen.getByText('date-2023-07-19')).toBeInTheDocument();
    expect(screen.getByText('page-0')).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePeriodToWeek'));
    });
    expect(screen.getByText('period-1')).toBeInTheDocument();
    expect(screen.getByText('date-2023-07-16')).toBeInTheDocument();
    expect(screen.getByText('page-0')).toBeInTheDocument();
  });
});
