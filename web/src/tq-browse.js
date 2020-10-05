import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import './shared-styles.js';
import '@polymer/iron-ajax/iron-ajax.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/iron-icon/iron-icon.js';
import '@polymer/iron-icons/iron-icons.js';

class Browse extends PolymerElement {
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
        .parents paper-button {
          color: purple;
          text-transform: none;
        }
        .dirs paper-button {
          color: blue;
          text-transform: none;
        }
        .add {
          color: green;
        }
        .remove {
          color: darkred;
        }
        .finalize {
          color: darkblue;
        }
        .timelapse {
          padding: 5px;
          margin: 5px;
          background: #f5f5f5;
        }
        .final {
          padding-top: 5px;
          padding-bottom: 5px;
        }
      </style>

      <iron-ajax
          auto
          url="/filebrowser"
          params="[[buildParams_(path)]]"
          handle-as="json"
          last-response="{{response}}"
          ></iron-ajax>
      <iron-ajax
          auto="[[parts.length]]"
          id="timelapseajax"
          url="/timelapse"
          params="[[toTimelapseParams_(parts.length)]]"
          handle-as="json"
          last-response="{{timelapse_}}"
          ></iron-ajax>
      <div class="card">
        <div class="circle">1</div>
        <h1>Select a Timelapse Part</h1>
        <div class="files parents">
          <template is="dom-repeat" items="[[response.Parents]]">
          <div>
           <paper-button on-tap="onDir_">
             <iron-icon icon="arrow-back"></iron-icon>
             <iron-icon icon="folder"></iron-icon>
             [[item.Name]]
           </paper-button>
          </div>
          </template>
        </div>
        <hr>
        <div class="files dirs">
          <template is="dom-repeat" items="[[response.Dirs]]">
          <div>
           <paper-button on-tap="onDir_">
             <iron-icon icon="folder"></iron-icon>
             [[item.Name]]
           </paper-button>
          </div>
          </template>
          <template is="dom-if" if="[[!response.Dirs]]">
          <div class="emptystate">
            <span>No more directories.</span>
          </div>
          </template>
        </div>
        <hr>
        <div class="files timelapses">
          <template is="dom-repeat" items="[[response.Timelapses]]">
          <div class="timelapse">
             <div>
              <a href="/image?path=[[item.Path]]" target="_blank">
               <img src="/image?path=[[item.Path]]&thumb=true" alt="[[item.Name]]">
              </a>
             </div>
             <div>[[item.Name]]</div>
             <div><span>[[item.Count]]</span> images</div>
             <paper-button class="add" on-tap="onSelectTimelapse_" raised>
               <iron-icon icon="add"></iron-icon>
               Add
             </paper-button>
          </div>
          </template>
          <template is="dom-if" if="[[!response.Timelapses]]">
          <div class="emptystate">
            <span>No timelapses in this directory.</span>
          </div>
          </template>
        </div>
      </div>

      <div class="card" hidden$="[[!hasValues_(parts.length)]]">
        <h3>Selected Timelapse Parts</h3>
        <p>
          A timelapse may contain multiple image sequences.  This can be used
    to combine multiple folders.  You can either select more timelapse parts
    above, or continue with the new timelapse job below.
        </p>

        <div class="files timelapses">
          <template is="dom-repeat" items="[[parts]]">
          <div class="timelapse">
             <div>
              <a href="/image?path=[[item.Path]]" target="_blank">
               <img src="/image?path=[[item.Path]]&thumb=true" alt="[[item.Name]]">
              </a>
             </div>
             <div>[[item.Name]]</div>
             <div><span>[[item.Count]]</span> images</div>
             <paper-button class="remove" on-tap="onRemoveTimelapse_" raised>
               <iron-icon icon="remove"></iron-icon>
               Remove
             </paper-button>
          </div>
          </template>
        </div>

        <div class="final">
             <div><b>New Job Contains</b></div>
             <div><span>[[timelapse_.Count]]</span> images</div>
             <div><span>[[parts.length]]</span> image sequence(s)</div>
             <div><span>[[timelapse_.DurationString]]</span> (at 60fps)</div>
        </div>

        <div>
          <paper-button class="finalize" on-tap="onFinalizeTimelapse_" raised>New Timelapse Job</paper-button>
        </div>
      </div>
    `;
  }

  onDir_(e) {
          this.path = e.model.item.Path;
  }

  buildParams_(path) {
      return {'path': path};
  }

  onSelectTimelapse_(e) {
    const timelapse = e.model.item;
    this.push('parts', timelapse);
  }

  onRemoveTimelapse_(e) {
    const timelapse = e.model.item;
    const index = this.parts.findIndex((t) => t.Path === timelapse.Path);
    this.splice('parts', index, 1);
  }

  onFinalizeTimelapse_(e) {
    const path = '/?path=' + this.parts.map((p) => p.Path).join(",") + '#/setup';

    // Reset internal state.
    this.set('parts', []);

    window.history.pushState({}, null, path);
    window.dispatchEvent(new CustomEvent('location-changed'));
  }

  toTimelapseParams_(l) {
    return {
      'path': this.parts.map((p) => p.Path).join(","),
    };
  }

  hasValues_(v) {
    return !!v;
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
      parts: {
        type: Array,
        value: function() {
          return [];
        },
      },
      timelapse_: {
        type: Object,
      },
    };
  }

}

window.customElements.define('tq-browse', Browse);
