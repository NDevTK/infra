/* Copyright 2019 The Chromium Authors. All Rights Reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file or at
 * https://developers.google.com/open-source/licenses/bsd
 */

/**
 * Form submission function for switching platforms.
 *
 * Allows for #platform_select to optionally populate
 * two fields of the form if '#' is present in the value.
 */
function switchPlatform() {
  const form = document.getElementById('platform_select_form');
  const select = document.getElementById('platform_select');
  const option = select.options[select.selectedIndex];
  if (option.value.indexOf('#') > -1) {
    const platform = option.value.split('#')[0];
    const revision = option.value.split('#')[1];
    option.value = platform;
    const revisionField = document.createElement('input');
    revisionField.setAttribute('type', 'hidden');
    revisionField.setAttribute('name', 'revision');
    revisionField.setAttribute('value', revision);
    form.appendChild(revisionField);
  }
  form.submit();
}
