// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { fireEvent, screen } from '@testing-library/react';
import { renderWithComponents } from '../../features/components/testUtils';
import NavBar from './NavBar';

describe('when rendering the navbar', () => {
  it('should set the component', () => {
    const updateMock = jest.fn();
    renderWithComponents(
        <NavBar/>,
        { api: { updateComponents: updateMock } },
    );
    const inputFields = screen.getByTestId('componentsTextField').getElementsByTagName('input');
    expect(inputFields).toHaveLength(1);
    const input = inputFields[0];
    fireEvent.change(input, { target: { value: 'Test' } });
    fireEvent.blur(input);
    expect(updateMock).toBeCalled();
    expect(updateMock).lastCalledWith(['Test']);
  });

  it('should set empty array when no components', () => {
    const updateMock = jest.fn();
    renderWithComponents(
        <NavBar/>,
        { api: { updateComponents: updateMock } },
    );
    const inputFields = screen.getByTestId('componentsTextField').getElementsByTagName('input');
    expect(inputFields).toHaveLength(1);
    const input = inputFields[0];
    fireEvent.change(input, { target: { value: '' } });
    fireEvent.blur(input);
    expect(updateMock).toBeCalled();
    expect(updateMock).lastCalledWith([]);
  });

  it('should set multiple components', () => {
    const updateMock = jest.fn();
    renderWithComponents(
        <NavBar/>,
        { api: { updateComponents: updateMock } },
    );
    const inputFields = screen.getByTestId('componentsTextField').getElementsByTagName('input');
    expect(inputFields).toHaveLength(1);
    const input = inputFields[0];
    fireEvent.change(input, { target: { value: 'A, B' } });
    fireEvent.blur(input);
    expect(updateMock).toBeCalled();
    expect(updateMock).lastCalledWith(['A', 'B']);
  });

  it('should not set duplicate components', () => {
    const updateMock = jest.fn();
    renderWithComponents(
        <NavBar/>,
        { api: { updateComponents: updateMock } },
    );
    const inputFields = screen.getByTestId('componentsTextField').getElementsByTagName('input');
    expect(inputFields).toHaveLength(1);
    const input = inputFields[0];
    fireEvent.change(input, { target: { value: 'A, A' } });
    fireEvent.blur(input);
    expect(updateMock).toBeCalled();
    expect(updateMock).lastCalledWith(['A']);
  });

  it('should not set empty components', () => {
    const updateMock = jest.fn();
    renderWithComponents(
        <NavBar/>,
        { api: { updateComponents: updateMock } },
    );
    const inputFields = screen.getByTestId('componentsTextField').getElementsByTagName('input');
    expect(inputFields).toHaveLength(1);
    const input = inputFields[0];
    fireEvent.change(input, { target: { value: 'A, ,B' } });
    fireEvent.blur(input);
    expect(updateMock).toBeCalled();
    expect(updateMock).lastCalledWith(['A', 'B']);
  });

  it('should handle whitespace well', () => {
    const updateMock = jest.fn();
    renderWithComponents(
        <NavBar/>,
        { api: { updateComponents: updateMock } },
    );
    const inputFields = screen.getByTestId('componentsTextField').getElementsByTagName('input');
    expect(inputFields).toHaveLength(1);
    const input = inputFields[0];
    fireEvent.change(input, { target: { value: ' A ,B,C   ' } });
    fireEvent.blur(input);
    expect(updateMock).toBeCalled();
    expect(updateMock).lastCalledWith(['A', 'B', 'C']);
  });

  it('should render navigation correctly', () => {
    renderWithComponents(
      <NavBar/>,
    );
    expect(screen.getByText('Coverage')).toBeInTheDocument();
    expect(screen.getByText('Resources')).toBeInTheDocument();
  })
});
