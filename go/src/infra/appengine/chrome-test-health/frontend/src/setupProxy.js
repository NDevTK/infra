// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/* eslint-disable */
const { createProxyMiddleware } = require('http-proxy-middleware');

const target = process.env.PROXY || 'http://127.0.0.1:8800'

// Proxy all /api methods
module.exports = function (app) {
  app.use(
    '/prpc',
    createProxyMiddleware({
      target: target,
      changeOrigin: true,
    })
  );
  app.use(
    '/auth',
    createProxyMiddleware({
      target: target,
      changeOrigin: true,
    })
  );
};
