import React from 'react';
import { render, screen, within } from '@testing-library/react';
import { store } from '../../app/store';
import { Provider } from 'react-redux';
import { ResourceList } from './ResourceList';
import userEvent from '@testing-library/user-event';
import { ResourceModel } from '../../api/resource_service';
import { ResourceRecordValidation, ResourceState, setState } from './resourceSlice';

const testState: ResourceState = {
  resources: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchStatus: 'idle',
  record: ResourceModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  recordValidation: ResourceRecordValidation.defaultEntity(),
};

const testResourceList: ResourceModel[] = [
  {
    ...ResourceModel.defaultEntity(),
    resourceId: 'rid-1',
    name: 'r1',
  },
  {
    ...ResourceModel.defaultEntity(),
    resourceId: 'rid-2',
    name: 'r2',
  },
];

const testStateWithResources: ResourceState = {
  ...testState,
  resources: testResourceList,
};

test('Clicking create button changes state of right side drawer', () => {
  render(
    <Provider store={store}>
      <ResourceList />
    </Provider>
  );

  expect(store.getState().utility.rightSideDrawerOpen).not.toBeTruthy();
  const createButtonField = screen.getByTestId('create-button');
  userEvent.click(createButtonField);
  expect(store.getState().utility.rightSideDrawerOpen).toBeTruthy();
});

test('Should have all the required columns', () => {
  store.dispatch(setState(testStateWithResources));
  render(
    <Provider store={store}>
      <ResourceList />
    </Provider>
  );
  expect(
    within(screen.getAllByRole('row')[0]).getAllByRole('columnheader')
  ).toHaveLength(8);
});

test('should have 2 data rows and a header row', () => {
  store.dispatch(setState(testStateWithResources));
  render(
    <Provider store={store}>
      <ResourceList />
    </Provider>
  );
  expect(screen.getAllByRole('row')).toHaveLength(3);
});

test('Should have just header row when resources are empty', () => {
  store.dispatch(setState(testState));
  render(
    <Provider store={store}>
      <ResourceList />
    </Provider>
  );
  // There is only one row present
  expect(screen.getAllByRole('row')).toHaveLength(1);

  // Header columns are present in that row
  expect(
    within(screen.getAllByRole('row')[0]).getAllByRole('columnheader')
  ).toHaveLength(8);

  // That row should not have cell element which only happens when there is at least 1 resource present
  expect(
    within(screen.getAllByRole('row')[0]).queryAllByRole('cell').length
  ).toBe(0);
});
