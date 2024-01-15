// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import dts from 'vite-plugin-dts';

export default defineConfig({
  plugins: [react(), dts({ include: ['src'], insertTypesEntry: true,})],
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      formats: ['es'],
      fileName: (format) => `log-viewer.${format}.js`,
      name: 'log-viewer'
    },
    rollupOptions: {
      external: [
        'react',
        'react/jsx-runtime',
        '@mui/material',
        /^@emotion\/\S+$/,
        /^@mui\/\S+$/,
      ]
    }
  },
})