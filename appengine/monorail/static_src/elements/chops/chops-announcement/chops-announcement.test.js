// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
import {assert} from 'chai';
import {ChopsAnnouncement, REFRESH_TIME_MS,
  XSSI_PREFIX} from './chops-announcement.js';
import sinon from 'sinon';

let element;
let clock;

function assertRendersMessage(message) {
  const messageContainer = element.shadowRoot.querySelector('mr-comment-content');
  assert.include(messageContainer.content, message);
}

function assertDoesNotRender() {
  assert.equal(0, element.shadowRoot.children.length);
}

describe('chops-announcement', () => {
  beforeEach(() => {
    element = document.createElement('chops-announcement-base');
    document.body.appendChild(element);

    element.additionalAnnouncements = [];

    clock = sinon.useFakeTimers({
      now: new Date(0),
      shouldAdvanceTime: false,
    });

    sinon.stub(window, 'fetch');
  });

  afterEach(() => {
    if (document.body.contains(element)) {
      document.body.removeChild(element);
    }

    clock.restore();

    window.fetch.restore();
  });

  it('does not request announcements when no service specified', async () => {
    sinon.stub(element, 'fetch');

    element.service = '';

    await element.updateComplete;

    sinon.assert.notCalled(element.fetch);
  });

  it('requests announcements when service is specified', async () => {
    sinon.stub(element, 'fetch');

    element.service = 'monorail';

    await element.updateComplete;

    sinon.assert.calledOnce(element.fetch);
  });

  it('refreshes announcements regularly', async () => {
    sinon.stub(element, 'fetch');

    element.service = 'monorail';

    await element.updateComplete;

    sinon.assert.calledOnce(element.fetch);

    clock.tick(REFRESH_TIME_MS);

    await element.updateComplete;

    sinon.assert.calledTwice(element.fetch);
  });

  it('stops refreshing when service removed', async () => {
    sinon.stub(element, 'fetch');

    element.service = 'monorail';

    await element.updateComplete;

    sinon.assert.calledOnce(element.fetch);

    element.service = '';

    await element.updateComplete;
    clock.tick(REFRESH_TIME_MS);
    await element.updateComplete;

    sinon.assert.calledOnce(element.fetch);
  });

  it('stops refreshing when element is disconnected', async () => {
    sinon.stub(element, 'fetch');

    element.service = 'monorail';

    await element.updateComplete;

    sinon.assert.calledOnce(element.fetch);

    document.body.removeChild(element);

    await element.updateComplete;
    clock.tick(REFRESH_TIME_MS);
    await element.updateComplete;

    sinon.assert.calledOnce(element.fetch);
  });

  it('renders error when thrown', async () => {
    sinon.stub(element, 'fetch');
    element.fetch.throws(() => Error('Something went wrong'));

    element.service = 'monorail';

    await element.updateComplete;

    // Fetch runs here.

    await element.updateComplete;

    assert.equal(element._error, 'Something went wrong');
    assert.include(element.shadowRoot.textContent, 'Something went wrong');
  });

  it('renders fetched announcement', async () => {
    sinon.stub(element, 'fetch');
    element.fetch.returns(
        {announcements: [{id: '1234', messageContent: 'test thing'}]});

    element.service = 'monorail';

    await element.updateComplete;

    // Fetch runs here.

    await element.updateComplete;

    assert.deepEqual(element._announcements,
        [{id: '1234', messageContent: 'test thing'}]);

    assertRendersMessage('test thing');
  });

  it('renders empty on empty announcement', async () => {
    sinon.stub(element, 'fetch');
    element.fetch.returns({});
    element.service = 'monorail';

    await element.updateComplete;

    // Fetch runs here.

    await element.updateComplete;

    assert.deepEqual(element._announcements, []);
    assertDoesNotRender()
  });

  it('fetch returns response data', async () => {
    const json = {announcements: [{id: '1234', messageContent: 'test thing'}]};
    const fakeResponse = XSSI_PREFIX + JSON.stringify(json);
    window.fetch.returns(new window.Response(fakeResponse));

    const resp = await element.fetch('monorail');

    assert.deepEqual(resp, json);
  });

  it('fetch errors when no XSSI prefix', async () => {
    const json = {announcements: [{id: '1234', messageContent: 'test thing'}]};
    const fakeResponse = JSON.stringify(json);
    window.fetch.returns(new window.Response(fakeResponse));

    try {
      await element.fetch('monorail');
    } catch (e) {
      assert.include(e.message, 'No XSSI prefix in announce response:');
    }
  });

  it('fetch errors when response is not okay', async () => {
    const json = {announcements: [{id: '1234', messageContent: 'test thing'}]};
    const fakeResponse = XSSI_PREFIX + JSON.stringify(json);
    window.fetch.returns(new window.Response(fakeResponse, {status: 500}));

    try {
      await element.fetch('monorail');
    } catch (e) {
      assert.include(e.message,
          'Something went wrong while fetching announcements');
    }
  });

  describe('additional announcement handlings', () => {
    beforeEach(() => {
      sinon.stub(element, 'fetch');
      element.fetch.returns({});
      element.service = 'monorail';
    });

    it('renders additional announcement', async () => {
      element.additionalAnnouncements = [{'messageContent': 'test thing'}];
      await element.updateComplete;

      assertRendersMessage('test thing');
    });

    it('renders when user is in group', async () => {
      element.additionalAnnouncements = [
        {'messageContent': 'test thing', 'groups': ['hello@group.com']}
      ];
      element.userGroups = [
        {"userId": "12344", "displayName": "hello@group.com"}];
      await element.updateComplete;

      assertRendersMessage('test thing');
    });

    it('does not render when user is not in group', async () => {
      element.additionalAnnouncements = [
        {'messageContent': 'test thing', 'groups': ['hello@group.com']}
      ];
      element.userGroups = [
        {"userId": "12344", "displayName": "hello@othergroup.com"}];
      await element.updateComplete;

      assertDoesNotRender();
    });

    it('renders when user is in everyone@ group', async () => {
      element.additionalAnnouncements = [
        {'messageContent': 'test thing', 'groups': ['everyone@world.com']}
      ];
      element.userGroups = [
        {"userId": "12344", "displayName": "hello@group.com"}];
      element.currentUserName = "hello@world.com";
      await element.updateComplete;

      assertRendersMessage('test thing');
    });

    it('does not renders when user is not in everyone@ group', async () => {
      element.additionalAnnouncements = [
        {'messageContent': 'test thing', 'groups': ['everyone@word.com']}
      ];
      element.userGroups = [
        {"userId": "12344", "displayName": "hello@world.com"}];
      element.currentUserName = "hello@world.com";
      await element.updateComplete;

      assertDoesNotRender();
    });

    it('renders when viewing referenced project', async () => {
      element.additionalAnnouncements = [
        {'messageContent': 'test thing', 'projects': ['chromium']}];
      element.currentProject = 'chromium';

      await element.updateComplete;

      assertRendersMessage('test thing');
    });

    it('does not render when not viewing referenced project', async () => {
      element.additionalAnnouncements = [
        {'messageContent': 'test thing', 'projects': ['chromium']}];
      element.currentProject = 'chrome';

      await element.updateComplete;

      assertDoesNotRender();
    });
  });
});
