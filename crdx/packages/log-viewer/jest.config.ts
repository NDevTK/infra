// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import type { Config } from 'jest';

const config: Config = {
  preset: 'ts-jest/presets/js-with-babel',
  testEnvironment: 'jsdom',
  testMatch: ['**/__tests__/**/*.[jt]s?(x)', '**/*.test.[jt]s?(x)'],
  // Some modules use `es6` modules, which is not compatible with jest, so we
  // need to transform them.
  transformIgnorePatterns: ['/node_modules/?!(lodash-es|lit)'],
  setupFilesAfterEnv: ['./src/test_utils/test_setup.ts'],
  moduleNameMapper: {
    '\\.(css|less)$': 'identity-obj-proxy',
    '^@/(.*)': '<rootDir>/src/$1',
    '^@root/(.*)': '<rootDir>/$1',
  },
};

export default config;
