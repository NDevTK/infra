import React from 'react';
import { render, screen } from '@testing-library/react';
import { AssetInstance } from './AssetInstance';
import { Provider } from 'react-redux';
import { store } from '../../app/store';
import { setRightSideDrawerOpen } from '../utility/utilitySlice';
import userEvent from '@testing-library/user-event';
import { AssetInstanceState, setState } from './assetInstanceSlice';
import { AssetInstanceModel } from '../../api/asset_instance_service';
import { setDefaultState } from '../resource/resourceSlice';

const testState: AssetInstanceState = {
  assetInstances: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 10,
  fetchStatus: 'idle',
  assets: [],
  savingStatus: 'idle',
  deleteAtBuffer: '',
  record: AssetInstanceModel.defaultEntity(),
};

beforeAll(() => {
  store.dispatch(setState(testState));
});

afterAll(() => {
  store.dispatch(setState(testState));
});

test('Renders asset instance creation form', () => {
  render(
    <Provider store={store}>
      <AssetInstance />
    </Provider>
  );

  const formHeading = screen.getByTestId('form-heading');
  expect(formHeading).toBeInTheDocument();

  const assetInstanceIdField = screen.getByTestId('asset-instance-id');
  expect(assetInstanceIdField).toBeInTheDocument();

  const assetIdField = screen.getByTestId('asset-id');
  expect(assetIdField).toBeInTheDocument();

  const statusField = screen.getByTestId('status');
  expect(statusField).toBeInTheDocument();

  const deleteTimeField = screen.getByTestId('delete-time');
  expect(deleteTimeField).toBeInTheDocument();

  const cancelButton = screen.getByTestId('cancel-button');
  expect(cancelButton).toBeInTheDocument();

  const saveButton = screen.getByTestId('save-button');
  expect(saveButton).toBeInTheDocument();
});

test('assetInstanceId field should be disabled', () => {
  render(
    <Provider store={store}>
      <AssetInstance />
    </Provider>
  );

  const assetInstanceIdField = screen.getByTestId('asset-instance-id');
  expect(assetInstanceIdField).toBeDisabled();
});

test('assetId field should be disabled', () => {
  render(
    <Provider store={store}>
      <AssetInstance />
    </Provider>
  );

  const assetIdField = screen.getByTestId('asset-id');
  expect(assetIdField).toBeDisabled();
});

test('status field should be disabled', () => {
  render(
    <Provider store={store}>
      <AssetInstance />
    </Provider>
  );

  const statusField = screen.getByTestId('status');
  expect(statusField).toBeDisabled();
});

test('Clicking on cancel button alters the state', () => {
  render(
    <Provider store={store}>
      <AssetInstance />
    </Provider>
  );

  store.dispatch(setRightSideDrawerOpen());
  expect(store.getState().utility.rightSideDrawerOpen).toBe(true);
  const cancelButton = screen.getByTestId('cancel-button');
  userEvent.click(cancelButton);
  expect(store.getState().utility.rightSideDrawerOpen).toBe(false);
});

test('Inputting delete time alters State', () => {
  render(
    <Provider store={store}>
      <AssetInstance />
    </Provider>
  );

  const deleteAtField = screen.getByTestId('delete-time');
  userEvent.type(deleteAtField, 'invalid timestamp');
  expect(store.getState().assetInstance.record.deleteAt).toBe(undefined);
  userEvent.type(deleteAtField, '2022-07-01T00:00');
  expect(store.getState().assetInstance.record.deleteAt?.toISOString()).toBe(
    new Date('2022-07-01T00:00:00-07:00').toISOString()
  );
});
