// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { Provider } from 'react-redux';
import { store } from './app/store';
import App from './App';

test('renders chromium text', () => {
  const { getByText } = render(
      <Provider store={store}>
        <App />
      </Provider>,
  );

  expect(getByText(/chromium/i)).toBeInTheDocument();
});
