// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { LitElement, html, css } from 'lit-element';
import 'elements/framework/mr-comment-content/mr-comment-content.js';

import { connectStore } from 'reducers/base.js';
import * as projectV0 from 'reducers/projectV0.js';
import * as userV0 from 'reducers/userV0.js';

// URL where announcements are fetched from.
const ANNOUNCEMENT_SERVICE =
  'https://chopsdash.appspot.com/prpc/dashboard.ChopsAnnouncements/SearchAnnouncements';

// Prefix prepended to responses for security reasons.
export const XSSI_PREFIX = ')]}\'';

const FETCH_HEADERS = Object.freeze({
  'accept': 'application/json',
  'content-type': 'application/json',
});

// How often to refresh announcements.
export const REFRESH_TIME_MS = 5 * 60 * 1000;

/**
 * @type {Array<Announcement>} A list of hardcodded announcements for Monorail.
 */
export const HARDCODED_ANNOUNCEMENTS = [{
  "messageContent": "The Chromium project will be migrating to Buganizer in " +
    "early 2024 (go/chrome-buganizer). Please test your workflows for this " +
    "transition with these instructions: go/cob-buv-quick-start",
  "projects": ["chromium"],
  "groups": ["everyone@google.com", "googlers@chromium.org"],
}];

/**
 * @typedef {Object} Announcement
 * @property {string=} id
 * @property {string} messageContent
 * @property {Array<string>=} projects Monorail extension for hard-coded
 *    announcements. Specifies the names of projects the announcement will
 *    occur in.
 * @property {Array<string>=} groups Monorail extension for hard-coded
 *    announcements. Specifies email groups the announces will show up in.
 */

/**
 * @typedef {Object} AnnouncementResponse
 * @property {Array<Announcement>} announcements
 */

/**
 * `<chops-announcement>` displays a ChopsDash message when there's an outage
 * or other important announcement.
 *
 * @customElement chops-announcement
 */
class _ChopsAnnouncement extends LitElement {
  /** @override */
  static get styles() {
    return css`
      :host {
        display: block;
        width: 100%;
      }
      mr-comment-content {
        display: block;
        color: #222;
        font-size: 13px;
        background: #FFCDD2; /* Material design red */
        width: 100%;
        text-align: center;
        padding: 0.5em 16px;
        box-sizing: border-box;
        margin: 0;
        /* Using a red-tinted grey border makes hues feel harmonious. */
        border-bottom: 1px solid #D6B3B6;
      }
    `;
  }
  /** @override */
  render() {
    if (this._error) {
      return html`<p><strong>Error: </strong>${this._error}</p>`;
    }
    return html`
      ${this._processedAnnouncements().map(
      ({ messageContent }) => html`
          <mr-comment-content
            .content=${messageContent}>
          </mr-comment-content>`)}
    `;
  }

  /** @override */
  static get properties() {
    return {
      service: { type: String },
      additionalAnnouncements: { type: Array },

      // Properties from the currently logged in user, usually feched through
      // Redux.
      currentUserName: { type: String },
      userGroups: { type: Array },
      currentProject: { type: String },

      // Private properties managing state from requests to Chops Dash.
      _error: { type: String },
      _announcements: { type: Array },
    };
  }

  /** @override */
  constructor() {
    super();

    /** @type {string} */
    this.service = undefined;
    /** @type {Array<Announcement>} */
    this.additionalAnnouncements = HARDCODED_ANNOUNCEMENTS;

    this.currentUserName = '';
    this.userGroups = [];
    this.currentProject = '';

    /** @type {string} */
    this._error = undefined;
    /** @type {Array<Announcement>} */
    this._announcements = [];

    /** @type {number} Interval ID returned by window.setInterval. */
    this._interval = undefined;
  }

  /** @override */
  updated(changedProperties) {
    if (changedProperties.has('service')) {
      if (this.service) {
        this.startRefresh();
      } else {
        this.stopRefresh();
      }
    }
  }

  /** @override */
  disconnectedCallback() {
    super.disconnectedCallback();

    this.stopRefresh();
  }

  /**
   * Set up autorefreshing logic or announcement information.
   */
  startRefresh() {
    this.stopRefresh();
    this.refresh();
    this._interval = window.setInterval(() => this.refresh(), REFRESH_TIME_MS);
  }

  /**
   * Logic for clearing refresh behavior.
   */
  stopRefresh() {
    if (this._interval) {
      window.clearInterval(this._interval);
    }
  }

  /**
   * Refresh the announcement banner.
   */
  async refresh() {
    try {
      const { announcements = [] } = await this.fetch(this.service);
      this._error = undefined;
      this._announcements = announcements;
    } catch (e) {
      this._error = e.message;
      this._announcements = HARDCODED_ANNOUNCEMENTS;
    }
  }

  /**
   * Fetches the announcement for a given service.
   * @param {string} service Name of the service to fetch from ChopsDash.
   *   ie: "monorail"
   * @return {Promise<AnnouncementResponse>} ChopsDash response JSON.
   * @throws {Error} If something went wrong while fetching.
   */
  async fetch(service) {
    const message = {
      retired: false,
      platformName: service,
    };

    const response = await window.fetch(ANNOUNCEMENT_SERVICE, {
      method: 'POST',
      headers: FETCH_HEADERS,
      body: JSON.stringify(message),
    });

    if (!response.ok) {
      throw new Error('Something went wrong while fetching announcements');
    }

    // We can't use response.json() because of the XSSI prefix.
    const text = await response.text();

    if (!text.startsWith(XSSI_PREFIX)) {
      throw new Error(`No XSSI prefix in announce response: ${XSSI_PREFIX}`);
    }

    return JSON.parse(text.substr(XSSI_PREFIX.length));
  }

  _processedAnnouncements() {
    const announcements = [...this.additionalAnnouncements, ...this._announcements];

    // Only show announcements relevant to the project the user is viewing and
    // the group the user is part of, if applicable.
    return announcements.filter(({ groups, projects }) => {
      if (groups && groups.length && !this._isUserInGroups(groups,
        this.userGroups, this.currentUserName)) {
        return false;
      }
      if (projects && projects.length && !this._isViewingProject(projects,
        this.currentProject)) {
        return false;
      }
      return true;
    });
  }

  /**
   * Helper to check if the user is a member of the allowed groups.
   * @param {Array<string>} allowedGroups
   * @param {Array<{{userId: string, displayName: string}}>} userGroups
   * @param {string} userEmail
   */
  _isUserInGroups(allowedGroups, userGroups, userEmail) {
    const userGroupSet = new Set(userGroups.map(
        ({ displayName }) => displayName.toLowerCase()));
    return allowedGroups.find((group) => {
      group = group.toLowerCase();

      // Handle custom groups in Monorail like everyone@google.com
      if (group.startsWith('everyone@')) {
        let [_, suffix] = group.split('@');
        suffix = '@' + suffix;
        return userEmail.endsWith(suffix);
      }

      return userGroupSet.has(group);
    });
  }

  _isViewingProject(projects, currentProject) {
    return projects.find((project = "") => project.toLowerCase() === currentProject.toLowerCase());
  }
}

/** Redux-connected version of _ChopsAnnouncement. */
export class ChopsAnnouncement extends connectStore(_ChopsAnnouncement) {
  /** @override */
  stateChanged(state) {
    const { displayName, groups } = userV0.currentUser(state);
    this.currentUserName = displayName;
    this.userGroups = groups;

    this.currentProject = projectV0.viewedProjectName(state);
  }
}

customElements.define('chops-announcement-base', _ChopsAnnouncement);
customElements.define('chops-announcement', ChopsAnnouncement);
