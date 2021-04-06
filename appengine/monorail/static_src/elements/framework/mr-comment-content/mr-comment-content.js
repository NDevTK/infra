// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {LitElement, html, css} from 'lit-element';
import {unsafeHTML} from 'lit-html/directives/unsafe-html.js';
import {ifDefined} from 'lit-html/directives/if-defined';
import {autolink} from 'autolink.js';
import {connectStore} from 'reducers/base.js';
import * as issueV0 from 'reducers/issueV0.js';
import * as projectV0 from 'reducers/projectV0.js';
import {SHARED_STYLES} from 'shared/shared-styles.js';
import {renderMarkdown, sanitizeMarkdown} from 'shared/md-helper.js';

/**
 * `<mr-comment-content>`
 *
 * Displays text for a comment.
 *
 */
export class MrCommentContent extends connectStore(LitElement) {
  /** @override */
  constructor() {
    super();

    this.content = '';
    this.commentReferences = new Map();
    this.isDeleted = false;
    this.projectName = '';
  }

  /** @override */
  static get properties() {
    return {
      content: {type: String},
      commentReferences: {type: Object},
      revisionUrlFormat: {type: String},
      isDeleted: {
        type: Boolean,
        reflect: true,
      },
      projectName: {type: String},
    };
  }

  /** @override */
  static get styles() {
    return [
      SHARED_STYLES,
      css`
        :host {
          word-break: break-word;
          font-size: var(--chops-main-font-size);
          line-height: 130%;
          font-family: var(--mr-toggled-font-family);
        }
        :host([isDeleted]) {
          color: #888;
          font-style: italic;
        }
        .line {
          white-space: pre-wrap;
        }
        .strike-through {
          text-decoration: line-through;
        }
      `,
    ];
  }

  /** @override */
  render() {
    // const options = { mdLibrary: 'showdown', sanitize: false, xssLibrary: 'js-xss' };
    // const options = { mdLibrary: 'marked', sanitize: false };
    // const options = { mdLibrary: 'showdown', sanitize: false };

    // const showdownHTML = renderMarkdown(this.content, { mdLibrary: 'showdown', sanitize: true, xssLibrary: 'DOMPurify' });
    const markedHTML = renderMarkdown(this.content, { mdLibrary: 'marked', sanitize: true, xssLibrary: 'DOMPurify' });
    // const markdownItHTML = renderMarkdown(this.content, { mdLibrary: 'markdown-it', sanitize: true, xssLibrary: 'DOMPurify' });

    // const intermediary = renderMarkdown(this.content, { mdLibrary: 'marked', sanitize: false });
    // const filteredHTML = sanitizeMarkdown(intermediary, { xssLibrary: 'js-xss' });

    const runs = autolink.markupAutolinks(
        this.content, this.commentReferences, this.projectName,
        this.revisionUrlFormat);

    const templates = runs.map((run) => {
      switch (run.tag) {
        case 'b':
          return html`<b class="line">${run.content}</b>`;
        case 'br':
          return html`<br>`;
        case 'a':
          return html`<a
            class="line"
            target="_blank"
            href=${run.href}
            class=${run.css}
            title=${ifDefined(run.title)}
          >${run.content}</a>`;
        default:
          return html`<span class="line">${run.content}</span>`;
      }
    });

    // suppose there is only one text run, the entirey of the Markdown content
    const mdAutolinkChunks = autolink.autolinkChunk(markedHTML, this.commentReferences, this.projectName, this.revisionUrlFormat)
    console.log('autolinkChunk(markedHTML)');
    console.log(mdAutolinkChunks);
    
    const mdAutolinkContent = mdAutolinkChunks.reduce((acc, chunk) => acc + chunk.content, '')
    // const mdAutolinkContent = mdAutolinkChunks.reduce((acc, chunk) => { return acc + chunk.content; }, '')
    console.log('mdAutolinkContent')
    console.log(mdAutolinkContent)




    // const rawAutoMark = runs.map((run) => {
    //   switch (run.tag) {
    //     case 'b':
    //       return html`<b class="line">${run.content}</b>`;
    //     case 'br':
    //       return html`<br>`;
    //     case 'a':
    //       return html`<a
    //         class="line"
    //         target="_blank"
    //         href=${run.href}
    //         class=${run.css}
    //         title=${ifDefined(run.title)}
    //       >${run.content}</a>`;
    //     default:
    //       const mdRunContent = renderMarkdown(run.content, { mdLibrary: 'marked', sanitize: true, xssLibrary: 'DOMPurify' });
    //       // console.log(mdRunContent)
    //       // return unsafeHTML`<span class="line">${mdRunContent}</span>`;
    //       return html`${unsafeHTML(mdRunContent)}`;
    //   }
    // });

    // const rawMarkAutoRuns = autolink.markupAutolinks(markedHTML, this.commentReferences, this.projectName, this.revisionUrlFormat)
    // const rawMarkAuto = rawMarkAutoRuns.map((run) => {
    //   switch (run.tag) {
    //     case 'b':
    //       return html`<b class="line">${run.content}</b>`;
    //     case 'br':
    //       return html`<br>`;
    //     case 'a':
    //       return html`<a
    //         class="line"
    //         target="_blank"
    //         href=${run.href}
    //         class=${run.css}
    //         title=${ifDefined(run.title)}
    //       >${run.content}</a>`;
    //     default:
    //       return html`${unsafeHTML(run.content)}`;
    //   }
    // });

    // const shouldRenderMarkdown = someHelpers.shouldRenderMarkdown({ project: '', issueToggle: true })
    // if (shouldRenderMarkdown) {
    //   try {
    //     return html`${mdRenderHelper(content)}`
    // } else {
    //   return html`${templates}`;
    // }

    // return html`<div>
    //   <h1>Original content</h1>${templates}
    //   <hr>
    //   <h1>Unfiltered content</h1>${unsafeHTML(markdownHTML)}
    //   <hr>
    //   <h1>Markdown content</h1>${unsafeHTML(sanitizedHTML)}
    //  </div>`

    return html`<div>
      <h2>Raw</h2>${this.content}
      <hr>

      <h2>Classic</h2>${templates}
      <hr>

      <h2>raw --> Markdown (no autolink)</h2>
      ${unsafeHTML(markedHTML)}
      <hr>


     </div>`
  }

  /** @override */
  stateChanged(state) {
    this.commentReferences = issueV0.commentReferences(state);
    this.projectName = issueV0.viewedIssueRef(state).projectName;
    this.revisionUrlFormat =
      projectV0.viewedPresentationConfig(state).revisionUrlFormat;
  }
}
customElements.define('mr-comment-content', MrCommentContent);
