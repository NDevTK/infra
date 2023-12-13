// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom/extend-expect';

// Material UI charts need to be mocked since they error
// out due to an export which jest doesn't like.
// Take a look at this github issue:
// https://github.com/mui/material-ui/issues/35465
jest.mock('@mui/x-charts/LineChart', () => (
  { LineChart: jest.fn().mockImplementation(({ children }) => children) }
));
