/**
 * @license
 * Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
 * This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
 * The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
 * The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
 * Code distributed by Google as part of the polymer project is also
 * subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
 */

import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import './shared-styles.js';
import '@polymer/iron-ajax/iron-ajax.js';
import '@polymer/paper-button/paper-button.js';

class MyView1 extends PolymerElement {
  static get template() {
    return html`
      <style include="shared-styles">
        :host {
          display: block;

          padding: 10px;
        }
        .files {
          display: flex;
          flex-wrap: wrap;
        }
        .parents {
          color: red;
        }
        .dirs {
          color: green;
        }
        .timelapses {
          color: blue;
        }
      </style>

      <iron-ajax
          auto
          url="/filebrowser"
          params="[[_buildParams(path)]]"
          handle-as="json"
          last-response="{{response}}"
          ></iron-ajax>

      <div class="card">
        <div class="circle">1</div>
        <h1>Select a timelapse</h1>
        <div class="files parents">
          <template is="dom-repeat" items="[[response.Parents]]">
          <div>
           <paper-button on-tap="_onDir">[[item.Name]]</paper-button>
          </div>
          </template>
        </div>
        <hr>
        <div class="files dirs">
          <template is="dom-repeat" items="[[response.Dirs]]">
          <div>
           <paper-button on-tap="_onDir">[[item.Name]]</paper-button>
          </div>
          </template>
        </div>
        <hr>
        <div class="files timelapses">
          <template is="dom-repeat" items="[[response.Timelapses]]">
          <div>
           <paper-button>[[item.Name]]</paper-button>
          </div>
          </template>
        </div>
      </div>
    `;
  }

  _onDir(e) {
          this.path = e.model.item.Path;
  }

  _buildParams(path) {
      return {'path': path};
  }

  static get properties() {
    return {
      path: {
        type: String,
        value: '',
      },
      response: {
        type: Object,
      },
    };
  }

}

window.customElements.define('my-view1', MyView1);
