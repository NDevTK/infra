// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { resolve } from 'path';

import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [
    react(),
    dts({
      include: ['src'],
      insertTypesEntry: true,
      exclude: ['./src/**/*.test.tsx', './src/test_utils/**'],
    }),
    tsconfigPaths(),
  ],
  build: {
    lib: {
      // eslint-disable-next-line no-undef
      entry: resolve(__dirname, 'src/index.ts'),
      formats: ['es'],
      fileName: (format) => `log-viewer.${format}.js`,
      name: 'log-viewer',
    },
    rollupOptions: {
      external: [
        'react',
        'react/jsx-runtime',
        '@mui/material',
        /^@emotion\/\S+$/,
        /^@mui\/\S+$/,
      ],
    },
  },
});
