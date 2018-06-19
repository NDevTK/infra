'use strict';

import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import '@polymer/polymer/lib/elements/dom-if.js';
import '@polymer/app-route/app-route.js'
import '@polymer/app-route/app-location.js'
import '@polymer/app-layout/app-header/app-header.js'
import '@polymer/app-layout/app-toolbar/app-toolbar.js'

import './tricium-feedback.js'
import './tricium-run.js'

class TriciumApp extends PolymerElement {

  static get properties() {
    return {
      route: Object,
      _mainPageActive: Boolean,
      _feedbackPageActive: Boolean,
      _feedbackPageData: Object,
      _runPageActive: Boolean,
      _runPageData: Object,
    };
  }

  static get template() {
    return html`
      <style>
        app-toolbar {
          background-color: hsl(221, 67%, 92%);
        }
        .link {
          font-size: 75%;
          margin-right: 10px;
        }
        .title {
          text-decoration: none;
          font-size: 115%;
        }
        [main-title] {
          pointer-events: auto;
          margin-right: 100px;
        }
      </style>
      <app-header reveals slot="header">
        <app-toolbar>
          <div main-title><a href="/" class="title">Tricium</a></div>
          <a href="/rpcexplorer/" class="link" target="_blank">RPC explorer</a>
        </app-toolbar>
      </app-header>

      <app-location route="{{route}}"></app-location>
      <app-route
        route="{{route}}"
        pattern="/"
        active="{{_mainPageActive}}">
      </app-route>
      <app-route
        route="{{route}}"
        pattern="/run/:runID"
        data="{{_runPageData}}"
        active="{{_runPageActive}}">
      </app-route>
      <app-route
        route="{{route}}"
        pattern="/feedback/:category"
        data="{{_feedbackPageData}}"
        active="{{_feedbackPageActive}}">
      </app-route>

      <template is="dom-if" if="[[_mainPageIsActive]]">
        <img id="mascot" src="/static/images/tri.png">
      </template>
      <template is="dom-if" if="[[_runPageActive]]">
        <tricium-run run-id="[[_runPageData.runID]]"></tricium-run>
      </template>
      <template is="dom-if" if="[[_feedbackPageActive]]">
        <tricium-feedback category="[[_feedbackPageData.category]]"></tricium-feedback>
      </template>
    `;
  }
}

customElements.define('tricium-app', TriciumApp);
