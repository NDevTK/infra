// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {css} from 'lit-element';

export const SHARED_STYLES = css`
  :host {
    --mr-edit-field-padding: 0.125em 4px;
    --mr-edit-field-width: 90%;
    --mr-input-grid-gap: 6px;
    font-family: var(--chops-font-family);
    color: var(--chops-primary-font-color);
    font-size: var(--chops-main-font-size);
  }
  /** Converts a <button> to look like an <a> tag. */
  .linkify {
    display: inline;
    padding: 0;
    margin: 0;
    border: 0;
    background: 0;
    cursor: pointer;
  }
  h1, h2, h3, h4 {
    background: none;
  }
  a, chops-button, a.button, .button, .linkify {
    color: var(--chops-link-color);
    text-decoration: none;
    font-weight: var(--chops-link-font-weight);
    font-family: var(--chops-font-family);
  }
  a:hover, .linkify:hover {
    text-decoration: underline;
  }
  a.button, .button {
    /* Links that look like buttons. */
    display: inline-flex;
    align-items: center;
    justify-content: center;
    text-decoration: none;
    transition: filter 0.3s ease-in-out, box-shadow 0.3s ease-in-out;
  }
  a.button:hover, .button:hover {
    filter: brightness(95%);
  }
  chops-button, a.button, .button {
    box-sizing: border-box;
    font-size: var(--chops-main-font-size);
    background: var(--chops-white);
    border-radius: 6px;
    --chops-button-padding: 0.25em 8px;
    margin: 0;
    margin-left: auto;
  }
  a.button, .button {
    padding: var(--chops-button-padding);
  }
  chops-button i.material-icons, a.button i.material-icons, .button i.material-icons {
    display: block;
    margin-right: 4px;
  }
  chops-button.emphasized, a.button.emphasized, .button.emphasized {
    background: var(--chops-primary-button-bg);
    color: var(--chops-primary-button-color);
    text-shadow: 1px 1px 3px hsla(0, 0%, 0%, 0.25);
  }
  textarea, select, input {
    box-sizing: border-box;
    font-size: var(--chops-main-font-size);
  }
  /* Note: decoupling heading levels from styles is useful for
  * accessibility because styles will not always line up with semantically
  * appropriate heading levels.
  */
  .medium-heading {
    font-size: var(--chops-large-font-size);
    font-weight: normal;
    line-height: 1;
    padding: 0.25em 0;
    color: var(--chops-link-color);
    margin: 0;
    margin-top: 0.25em;
    border-bottom: var(--chops-normal-border);
  }
  .medium-heading chops-button {
    line-height: 1.6;
  }
  .input-grid {
    padding: 0.5em 0;
    display: grid;
    max-width: 100%;
    grid-gap: var(--mr-input-grid-gap);
    grid-template-columns: minmax(120px, max-content) 1fr;
    align-items: flex-start;
  }
  .input-grid label {
    font-weight: bold;
    text-align: right;
    word-wrap: break-word;
  }
  @media (max-width: 600px) {
    .input-grid label {
      margin-top: var(--mr-input-grid-gap);
      text-align: left;
    }
    .input-grid {
      grid-gap: var(--mr-input-grid-gap);
      grid-template-columns: 100%;
    }
  }
`;

/**
 * Markdown specific styling:
 * * render link destination on hover as a tooltip
 * @type {CSSResult}
 */
export const MD_STYLES = css`
  .markdown .annotated-link {
    position: relative;
  }
  .markdown .annotated-link:hover .tooltip {
    display: block
  }
  .markdown .tooltip {
    display: none;
    position: absolute;
    width: auto;
    white-space: nowrap;
    box-shadow: rgb(170 170 170) 1px 1px 5px;
    box-shadow: 0 4px 8px 3px rgb(0 0 0 / 10%);
    border-radius: 8px;
    background-color: rgb(255, 255, 255);
    top: -32px;
    left: 0px;
    border: 1px solid #dadce0;
    padding: 6px 10px;
  }
  .markdown .material-icons {
    font-size: 18px;
    vertical-align: middle;
  }
  .markdown .material-icons.link {
    color: var(--chops-link-color);
  }
  .markdown .material-icons.link_off {
    color: var(--chops-field-error-color);
  }
  .markdown table {
    -webkit-font-smoothing: antialiased;
    box-sizing: inherit;
    border-collapse: collapse;
    margin: 8px 0 8px 0;
    box-shadow: 0 2px 2px 0 hsla(315, 3%, 26%, 0.30);
    border: 1px solid var(--chops-gray-300);
    line-height: 1.4;
  }
  .markdown th {
      border-bottom: 1px solid var(--chops-gray-300);
      border-right: 1px solid var(--chops-gray-300);
      padding: 1px;
      text-align: left;
      font-weight: 500;
      color: var(--chops-gray-900);
      background-color: var(--chops-gray-50);
  }
  .markdown td {
      border-bottom: 1px solid var(--chops-gray-300);
      border-right: 1px solid var(--chops-gray-300);
      padding: 1px;
  }
  .markdown pre {
    -webkit-font-smoothing: antialiased;
    line-height: 1.6;
    box-sizing: inherit;
    background-color: hsla(0, 0%, 0%, 0.05);
    border: 2px solid hsla(0, 0%, 0%, 0.10);
    border-radius: 2px;
    overflow-x: auto;
    padding: 4px;
  }
`;

export const MD_PREVIEW_STYLES = css`
  ${MD_STYLES}
  .markdown-preview {
    padding: 0.25em 1em;
    color: var(--chops-gray-800);
    background-color: var(--chops-gray-200);
    border-radius: 10px;
    margin: 0px 0px 10px;
    overflow: auto;
  }
  .preview-height-description {
    max-height: 40vh;
  }
  .preview-height-comment {
    min-height: 5vh;
    max-height: 15vh;
  }
`;
