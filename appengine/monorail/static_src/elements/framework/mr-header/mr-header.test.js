// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {assert} from 'chai';
import {MrHeader} from './mr-header.js';


window.CS_env = {
  token: 'foo-token',
};

let element;

describe('mr-header', () => {
  beforeEach(() => {
    element = document.createElement('mr-header');
    document.body.appendChild(element);
  });

  afterEach(() => {
    document.body.removeChild(element);
  });

  it('initializes', () => {
    assert.instanceOf(element, MrHeader);
  });

  it('presentationConfig renders', async () => {
    element.projectName = 'proj';
    element.issueEntryUrl = 'https://google.com/test/';
    element.projectName = 'best-project';
    element.projectThumbnailUrl = 'http://images.google.com/';
    element.presentationConfig = {
      projectSummary: 'The best project',
    };

    await element.updateComplete;

    assert.equal(element.shadowRoot.querySelector('.project-logo').src,
        'http://images.google.com/');

    assert.equal(element.shadowRoot.querySelector('.new-issue-link').href,
        'https://google.com/test/');

    assert.equal(element.shadowRoot.querySelector('.project-selector').title,
        'The best project');
  });

  it('canAdministerProject is false when user is not logged in', () => {
    element.userDisplayName = '';

    assert.isFalse(element.canAdministerProject);
  });

  it('canAdministerProject is true when user is site admin', () => {
    element.userDisplayName = 'test@example.com';
    element.isSiteAdmin = true;

    assert.isTrue(element.canAdministerProject);

    element.isSiteAdmin = false;

    assert.isFalse(element.canAdministerProject);
  });

  it('canAdministerProject is true when user is owner', () => {
    element.userDisplayName = 'test@example.com';
    element.isSiteAdmin = false;

    element.projectName = 'chromium';
    element.userProjects = {ownerOf: ['chromium']};

    assert.isTrue(element.canAdministerProject);

    element.projectName = 'v8';

    assert.isFalse(element.canAdministerProject);

    element.userProjects = {memberOf: ['v8']};

    assert.isFalse(element.canAdministerProject);
  });

  it('_projectDropdownItems tells user to sign in if not logged in', () => {
    element.userDisplayName = '';
    element.loginUrl = 'http://login';

    const items = element._projectDropdownItems;

    // My Projects
    assert.deepEqual(items[0], {
      text: 'Sign in to see your projects',
      url: 'http://login',
    });
  });

  it('_projectDropdownItems computes projects for user', () => {
    element.userProjects = {
      ownerOf: ['chromium'],
      memberOf: ['v8'],
      contributorTo: ['skia'],
      starredProjects: ['gerrit'],
    };
    element.userDisplayName = 'test@example.com';

    const items = element._projectDropdownItems;

    // TODO(http://crbug.com/monorail/6236): Replace these checks with
    // deepInclude once we upgrade Chai.
    // My Projects
    assert.equal(items[1].text, 'chromium');
    assert.equal(items[1].url, '/p/chromium/issues/list');
    assert.equal(items[2].text, 'skia');
    assert.equal(items[2].url, '/p/skia/issues/list');
    assert.equal(items[3].text, 'v8');
    assert.equal(items[3].url, '/p/v8/issues/list');

    // Starred Projects
    assert.equal(items[5].text, 'gerrit');
    assert.equal(items[5].url, '/p/gerrit/issues/list');
  });
});
