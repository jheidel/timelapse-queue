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
        .timelapses paper-button {
          color: green;
        }
        .timelapse {
          padding: 5px;
          margin: 5px;
          background: #f5f5f5;
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
        <h1>Select a Timelapse</h1>
        <div class="files parents">
          <template is="dom-repeat" items="[[response.Parents]]">
          <div>
           <paper-button on-tap="_onDir">
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
           <paper-button on-tap="_onDir">
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
             <div><span>[[item.Count]]</span> images (<span>[[item.DurationString]] @60fps</span>)</div>
            <paper-button on-tap="_onSelectTimelapse" raised>New Timelapse Job</paper-button>
          </div>
          </template>
          <template is="dom-if" if="[[!response.Timelapses]]">
          <div class="emptystate">
            <span>No timelapses in this directory.</span>
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

  _onSelectTimelapse(e) {
      const timelapse = e.model.item;

      console.log(timelapse.Path);

      window.location.href = '/?path=' + timelapse.Path + '#/setup';
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

window.customElements.define('tq-browse', Browse);
