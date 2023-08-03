/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { render } from '@testing-library/react';
import * as ComponentContextP from './features/components/ComponentContext';
import App from './App';
import { COMPONENT } from './features/components/ComponentParams';

describe('when rendering the App', () => {
  // This is needed to allow us to modify window.location
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { assign: jest.fn() },
  });
  it('should pass in default values', async () => {
    const mockComponentContext = jest.fn();
    jest.spyOn(ComponentContextP, 'ComponentContextProvider').mockImplementation((props) => {
      return mockComponentContext(props);
    });
    render(<App/>);
    expect(mockComponentContext).toHaveBeenCalledWith(
        expect.objectContaining({
          components: ['Blink'],
        }),
    );
  });
  it('should pass in url param values', async () => {
    const mockComponentContext = jest.fn();
    jest.spyOn(ComponentContextP, 'ComponentContextProvider').mockImplementation((props) => {
      return mockComponentContext(props);
    });
    window.location.search = 'https://test.com?placeholder=p&comp=Admin';
    render(<App/>);
    expect(mockComponentContext).toHaveBeenCalledWith(
        expect.objectContaining({
          components: ['Admin'],
        }),
    );
  });
  it('should pass in localStorage values', async () => {
    const mockComponentContext = jest.fn();
    window.location.search = '';
    jest.spyOn(ComponentContextP, 'ComponentContextProvider').mockImplementation((props) => {
      return mockComponentContext(props);
    });
    localStorage.setItem(COMPONENT, 'LOCALSTORAGE1,LOCALSTORAGE2');
    render(<App/>);
    expect(mockComponentContext).toHaveBeenCalledWith(
        expect.objectContaining({
          components: ['LOCALSTORAGE1', 'LOCALSTORAGE2'],
        }),
    );
  });
});
