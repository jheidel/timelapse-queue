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
import '@polymer/paper-progress/paper-progress.js';

class MyView2 extends PolymerElement {
  static get template() {
    return html`
      <style include="shared-styles">
        :host {
          display: block;
          padding: 10px;
        }
        .queue-item {
          padding: 5px;
          margin: 5px;
          display: flex;
          align-items: center;
        }
        .queue-item > div {
          padding: 5px;
        }

        .queue-pending {
          background: #DDDDDD;
        }
        .queue-active {
          background: #FFFFCC;
        }
        .queue-done {
          background: #CCFFCC;
        }
        .queue-failed {
          background: #FFCCCC;
        }
      </style>

      <iron-ajax
          id="ajax"
          url="/queue"
          handle-as="json"
          on-response="onResponse_"
          auto></iron-ajax>

      <div class="card">
        <div class="circle">2</div>
        <h1>Process Queue</h1>
        <div>
          <template is="dom-repeat" items="[[response.Queue]]">
            <div class$="queue-item queue-[[item.State]]">
              <div>
               <img src="/image?path=[[item.Timelapse.Path]]&thumb=true">
              </div>
              <div>[[item.Timelapse.Name]]</div>
              <div>
                  <paper-progress value="[[item.Progress]]"></paper-progress>
              </div>
              <div>[[item.Progress]]%</div>
              <div>[[item.State]]</div>
              <div hidden$="[[!item.LogPath]]">
                <a href="/log?path=[[item.LogPath]]" target="_blank">Log</a>
              </div>
            </div>
          </template>
        </div>
      </div>
    `;
  }

  onResponse_(e) {
    this.response = e.detail.response;
    // Poll again after short delay.
    setTimeout(() => this.$.ajax.generateRequest(), 1500);
  }

  static get properties() {
    return {
      response: {
        type: Object,
      },
    };
  }
}

window.customElements.define('my-view2', MyView2);
