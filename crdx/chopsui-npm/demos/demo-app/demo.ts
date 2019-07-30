import {LitElement, html, css, customElement, property} from 'lit-element';

import {connect} from 'pwa-helpers/connect-mixin';
import {store, AppState, Todo, User, toggleTodo, updateUser} from './store';

import '@chopsui/chops-header/chops-header.js';

import {PrpcClient} from '@chopsui/prpc-client/prpc-client.js';

import {ListStatusesRequest} from './generated/api/api_proto/projects_pb';

// Note that this demo assumes you are running a Monorail instance locally on 8080.
const monorailHost = 'localhost:8080';

@customElement('chops-demo-app')
export class Demo extends connect(store)(LitElement) {
  @property() _todos: Todo[] = [];
  @property() _prpcResponseText: String = '';
  @property() _prpcResponseError: String = '';
  @property() _user: User = {name: '', id: -1};

  constructor() {
    super();
  }

  static get styles() {
    return css`
      body {
        margin: 0;
        padding: 0;
      }
      chops-header {
        border: 2px solid red;
      }
    `;
  }

  stateChanged(state: AppState) {
    this._todos = state.todos || [];
    this._user = state.user || {name: '', id: -1};
  }

  connectedCallback() {
    super.connectedCallback();
  }

  async _prpcDemo() {
    const req = new ListStatusesRequest();
    req.setProjectName('test');
    const prpcClient = new PrpcClient({host: monorailHost, insecure: true});
    const fixedReq = this._fixArrayNames(req.toObject());
    try {
      // because prpcClient isn't yet typed, we can't do this yet:
      // const resp : ListStatusesResponse = await prpcClient.call('monorail.Projects', 'ListStatuses', fixedReq)
      const resp = await prpcClient.call('monorail.Projects', 'ListStatuses', fixedReq);
      this._prpcResponseText = JSON.stringify(resp);
      this._prpcResponseError = '';
    } catch (e) {
      this._prpcResponseText = '';
      this._prpcResponseError = JSON.stringify(e);
    }

  }

  // This is necessary because the serializer provided by protoc generates array properties
  // named 'fooList' when pRPC servers expect them to be named simply 'foo'.
  _fixArrayNames(obj) {
    const names = Object.getOwnPropertyNames(obj);
    names.forEach((name) => {
      if (name.endsWith('List')) {
        const p = obj[name];
        delete obj[name];
        obj[name.substr(0, name.length - 4)] = p;
      }
    });
    return obj;
  }

  render() {
    return html`
      <chops-header appTitle="ChOps Demo App">
      <div slot="before-header">
        [hamburger menu]
      </div>
      <div slot="subheader">
        [Subheader goes here.]
      </div>
      [Main slot of the element. Searchbars and stuff go here.]
      ${(this._user.name === '') ?
      html`<button @click="${this.loginClicked}">Log in</button>` :
      html`<button @click="${this.logoutClicked}">Log out ${this._user.name}</button>`
      }
    </chops-header>
    <button @click=${this._prpcDemo}>pRPC Demo</button>
    <textarea>Error: ${this._prpcResponseError}</textarea>
    <textarea>Response: ${this._prpcResponseText}</textarea>
    <ul>
      ${this._todos.map((todo, i) => html`<li>item: ${todo.text} ${todo.completed}<button data-index="${i}" @click="${this.doneClicked}">Done</button></li>`)}
    </ul>
    `;
  }

  loginClicked() {
    store.dispatch(updateUser({name: 'User Name', id: 123}));
  }

  logoutClicked() {
    store.dispatch(updateUser({name: '', id: -1}));
  }

  doneClicked(evt) {
    const idx = parseInt(evt.target.getAttribute('data-index'));
    store.dispatch(toggleTodo(idx));
  }
}
