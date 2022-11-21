import React from 'react';
import { render, screen, waitForElement } from '@testing-library/react';
import { Asset } from './Asset';
import { Provider } from 'react-redux';
import { store } from '../../app/store';
import { setRightSideDrawerOpen } from '../utility/utilitySlice';
import userEvent from '@testing-library/user-event';
import { AssetState, AssetRecordValidation, setState } from './assetSlice';
import { AssetResourceModel } from '../../api/asset_resource_service';
import { AssetModel } from '../../api/asset_service';
import { ResourceModel } from '../../api/resource_service';

const testResource: ResourceModel = {
  resourceId: 'test resource id',
  name: 'test name',
  type: 'test type',
  imageFamily: 'test image family',
  imageProject: 'test image project',
  imageSource: 'test image source',
  description: 'test description',
  operatingSystem: 'test os',
  createdAt: undefined,
  createdBy: '',
  modifiedAt: undefined,
  modifiedBy: '',
  deleted: false,
};

const testAssetResource: AssetResourceModel = {
  assetResourceId: 'test asset resource id',
  assetId: 'test asset id',
  resourceId: 'test resource id',
  aliasName: 'test alias',
  createdAt: undefined,
  createdBy: '',
  modifiedAt: undefined,
  modifiedBy: '',
  default: false,
};

const testState: AssetState = {
  assets: [],
  pageToken: undefined,
  pageNumber: 1,
  pageSize: 25,
  fetchAssetStatus: 'idle',
  record: AssetModel.defaultEntity(),
  savingStatus: 'idle',
  deletingStatus: 'idle',
  fetchAssetResourceStatus: 'idle',
  resources: [testResource],
  assetResourcesToSave: [testAssetResource, AssetResourceModel.defaultEntity()],
  assetResourcesToDelete: [],
  assetSpinRecord: '',
  fetchResourceStatus: '',
  defaultAssetResources: [],
  recordValidation: AssetRecordValidation.defaultEntity(),
  showDefaultMachines: false,
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
      <Asset />
    </Provider>
  );

  const formHeading = screen.getByTestId('form-heading');
  expect(formHeading).toBeInTheDocument();

  const name = screen.getByTestId('name');
  expect(name).toBeInTheDocument();

  const description = screen.getByTestId('description');
  expect(description).toBeInTheDocument();

  const type = screen.getByTestId('type');
  expect(type).toBeInTheDocument();

  const assetIdField = screen.getByTestId('asset-id');
  expect(assetIdField).toBeInTheDocument();

  const machinesHeading = screen.getByTestId('machines-heading');
  expect(machinesHeading).toBeInTheDocument();

  const alias = screen.getByTestId('alias-0');
  expect(alias).toBeInTheDocument();

  const resource = screen.getByTestId('resource-0');
  expect(resource).toBeInTheDocument();

  const cancelButton = screen.getByTestId('cancel-button');
  expect(cancelButton).toBeInTheDocument();

  const saveButton = screen.getByTestId('save-button');
  expect(saveButton).toBeInTheDocument();
});

test('assetId field should be disabled', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  const assetIdField = screen.getByTestId('asset-id');
  expect(assetIdField).toBeDisabled();
});

test('Clicking on cancel button alters the state', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  store.dispatch(setRightSideDrawerOpen());
  expect(store.getState().utility.rightSideDrawerOpen).toBe(true);
  const cancelButton = screen.getByTestId('cancel-button');
  userEvent.click(cancelButton);
  expect(store.getState().utility.rightSideDrawerOpen).toBe(false);
});

test('Adding text to name field alters state', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  expect(store.getState().asset.record.name).toBe('');
  const nameField = screen.getByTestId('name');
  userEvent.type(nameField, 'test name');
  expect(store.getState().asset.record.name).toBe('test name');
});

test('Adding description to description field alters state', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  expect(store.getState().asset.record.description).toBe('');
  const nameField = screen.getByTestId('description');
  userEvent.type(nameField, 'test description');
  expect(store.getState().asset.record.description).toBe('test description');
});

test('Asset type is active directory by default', async () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  expect(store.getState().asset.record.assetType).toBe('active_directory');
});

test('Selecting resource alters state', async () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );
  store.dispatch(setState(testState));

  expect(store.getState().asset.resources[0].resourceId).toBe(
    'test resource id'
  );
  const resourceSelectField = document.getElementById('resource-0')!;
  userEvent.click(resourceSelectField);
  const optionField = await waitForElement(() =>
    screen.getByTestId('resource-option-test resource id')
  );
  userEvent.click(optionField);
  expect(store.getState().asset.assetResourcesToSave[0].resourceId).toBe(
    'test resource id'
  );
});

test('Adding text to alias name field alters state', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  expect(store.getState().asset.assetResourcesToSave[0].aliasName).toBe(
    'test alias'
  );
  const aliasField = screen.getByTestId('alias-0');
  userEvent.type(aliasField, 'test alias updated');
  expect(store.getState().asset.assetResourcesToSave[0].aliasName).toBe(
    'test alias updated'
  );
});

test('Deleting entry for associated machines alters state', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  const assetResourcesToSave = store.getState().asset.assetResourcesToSave;

  // Delete an non-empty asset resource, should be added to assetResourcesToDelete list
  const addButonField = screen.getByTestId('delete-button-0')!;
  userEvent.click(addButonField);
  expect(store.getState().asset.assetResourcesToSave).toStrictEqual([
    AssetResourceModel.defaultEntity(),
  ]);
  expect(store.getState().asset.assetResourcesToDelete).toStrictEqual([
    assetResourcesToSave[0],
  ]);

  // Delete an empty asset resource; assetResourcesToDelete should not change
  // Additionally, since there is only one machine entry to display, the user should not be able
  // to delete this entry.
  userEvent.click(addButonField);
  expect(store.getState().asset.assetResourcesToSave).toStrictEqual([
    AssetResourceModel.defaultEntity(),
  ]);
  expect(store.getState().asset.assetResourcesToDelete).toStrictEqual([
    assetResourcesToSave[0],
  ]);
});

test('Adding entry for associated machines alters state', () => {
  render(
    <Provider store={store}>
      <Asset />
    </Provider>
  );

  const assetResourcesToSave = store.getState().asset.assetResourcesToSave;
  const addButonField = screen.getByTestId('add-button-0')!;
  userEvent.click(addButonField);
  expect(store.getState().asset.assetResourcesToSave).toStrictEqual([
    ...assetResourcesToSave,
    AssetResourceModel.defaultEntity(),
  ]);
});
