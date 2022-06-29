import React from 'react';
import { render, screen, waitForElement } from '@testing-library/react';
import { Resource } from './Resource';
import { Provider } from 'react-redux';
import { store } from '../../app/store';
import { setRightSideDrawerOpen } from '../utility/utilitySlice';
import userEvent from '@testing-library/user-event';
import { setDefaultState } from './resourceSlice';

beforeAll(() => {
  store.dispatch(setDefaultState());
});

afterAll(() => {
  store.dispatch(setDefaultState());
});

test('Renders resource creation form', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  const formHeading = screen.getByTestId('form-heading');
  expect(formHeading).toBeInTheDocument();

  const nameField = screen.getByTestId('name');
  expect(nameField).toBeInTheDocument();

  const descriptionField = screen.getByTestId('description');
  expect(descriptionField).toBeInTheDocument();

  const imageProjectField = screen.getByTestId('image-project');
  expect(imageProjectField).toBeInTheDocument();

  const imageFamilyField = screen.getByTestId('image-family');
  expect(imageFamilyField).toBeInTheDocument();

  const operatingSystemField = screen.getByTestId('operating-system');
  expect(operatingSystemField).toBeInTheDocument();

  const resourceIdField = screen.getByTestId('resource-id');
  expect(resourceIdField).toBeInTheDocument();

  const cancelButton = screen.getByTestId('cancel-button');
  expect(cancelButton).toBeInTheDocument();

  const saveButton = screen.getByTestId('save-button');
  expect(saveButton).toBeInTheDocument();
});

test('ResourceId field should be disabled', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  const resourceIdField = screen.getByTestId('resource-id');
  expect(resourceIdField).toBeDisabled();
});

test('Clicking on cancel button alters the state', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  store.dispatch(setRightSideDrawerOpen());
  expect(store.getState().utility.rightSideDrawerOpen).toBe(true);

  const cancelButton = screen.getByTestId('cancel-button');
  userEvent.click(cancelButton);
  expect(store.getState().utility.rightSideDrawerOpen).toBe(false);
});

test('Adding text to Name Field Alters State', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  expect(store.getState().resource.record.name).toBe('');

  const nameField = screen.getByTestId('name');
  userEvent.type(nameField, 'Test Resource 1');
  expect(store.getState().resource.record.name).toBe('Test Resource 1');
});

test('Adding text to Description Field Alters State', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  expect(store.getState().resource.record.description).toBe('');

  const descriptionField = screen.getByTestId('description');
  userEvent.type(descriptionField, 'Test Resource Description');
  expect(store.getState().resource.record.description).toBe(
    'Test Resource Description'
  );
});

test('Adding text to Image Project Field Alters State', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  expect(store.getState().resource.record.imageProject).toBe('');

  const imageProjectField = screen.getByTestId('image-project');
  userEvent.type(imageProjectField, 'Test Image Project');
  expect(store.getState().resource.record.imageProject).toBe(
    'Test Image Project'
  );
});

test('Adding text to Image Family Field Alters State', () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  expect(store.getState().resource.record.imageFamily).toBe('');

  const imageFamilyField = screen.getByTestId('image-family');
  userEvent.type(imageFamilyField, 'Test Image Family');
  expect(store.getState().resource.record.imageFamily).toBe(
    'Test Image Family'
  );
});

test('Selecting an Operating System Alters State', async () => {
  render(
    <Provider store={store}>
      <Resource />
    </Provider>
  );

  expect(store.getState().resource.record.operatingSystem).toBe('');

  const osSelectField = document.getElementById('operating-system')!;
  userEvent.click(osSelectField);
  const optionField = await waitForElement(() =>
    screen.getByTestId('os-option')
  );
  userEvent.click(optionField);
  expect(store.getState().resource.record.operatingSystem).toBe(
    'windows_machine'
  );
});
