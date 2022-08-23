// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import * as DOMPurify from 'dompurify';

import createInnerHTMLSanitizingPolicy from './create-policy';

describe('createInnerHTMLSanizingPolicy', () => {
  beforeAll(() => {
    createInnerHTMLSanitizingPolicy();
  });

  it('should create default policy', () => {
    expect(window.trustedTypes?.defaultPolicy).not.toBeNull();
  });

  it('given an html then it should return a sanitized version', () => {
    const dangerousImg = '<img src="nonexistent.png" onerror="alert(\'This restaurant got voted worst in town!\');" />';

    const domPurifiedImg = DOMPurify.sanitize(dangerousImg, { RETURN_TRUSTED_TYPE: true });
    const policyPurifiedImg = window.trustedTypes?.defaultPolicy?.createHTML(dangerousImg).toString();
    expect(domPurifiedImg).toBe(policyPurifiedImg);
  });
});
