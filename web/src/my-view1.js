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
import '@polymer/paper-spinner/paper-spinner.js';
import Croppr from 'croppr/src/croppr.js';

class MyView1 extends PolymerElement {
  static get template() {
    return html`
      <link rel="stylesheet" href="../node_modules/croppr/src/css/croppr.css">
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
        .cropbox {
            width: 640px;
            height: 480px;
        }
      </style>

      <iron-ajax
          auto
          url="/filebrowser"
          params="[[_buildParams(path)]]"
          handle-as="json"
          last-response="{{response}}"
          ></iron-ajax>
      <iron-ajax
          id="convert"
          url="/convert"
          method="POST"
          handle-as="text"
          on-response="onConvertSuccess_"
          on-error="onConvertSuccess_"
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
             <div>
              <a href="/image?path=[[item.Path]]" target="_blank">
               <img src="/image?path=[[item.Path]]&thumb=true" alt="[[item.Name]]">
              </a>
             </div>
             <div><span>[[item.Count]]</span> images (<span>[[item.DurationString]]</span>)</div>
            <paper-button on-tap="_onSelectTimelapse">[[item.Name]]</paper-button>
          </div>
          </template>
        </div>


        <hr>
        <paper-spinner active="[[loading_]]"></paper-spinner>
        <div>
          <div class="cropbox">
            <img id="croppr"/>
          </div>
          <div>
            <span>x=[[crop.x]] y=[[crop.y]]</span>
            <span>Size [[crop.width]]x[[crop.height]]</span>
          </div>
        </div>
        <div>
            <paper-button on-tap="onConvert_">Start Job</paper-button>
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
      console.log(e);
      console.log(e.model.item);
      this.loading_ = true;
      this.timelapse = e.model.item;

      this.$.croppr.src = '/image?path=' + this.timelapse.Path;

      // TODO set these based on job configuration.
      const width = 1920;
      const height = 1080;

      this.croppr = new Croppr(this.$.croppr, {
              aspectRatio: height / width,
              startSize: [100, 100, '%'],
              // TODO would be nice to set this but it's buggy.
              // minSize: [width, height, 'px'],
              onCropMove: (value) => {
                this.crop = value;
              },
              onInitialize: (instance) => {
                this.crop = instance.getValue();
                this.showSelect_ = true;
                this.loading_ = false;
              },
      });
  }

  onConvert_(e) {
    this.$.convert.headers={'content-type': 'application/x-www-form-urlencoded'};
    this.$.convert.body = {
      'path': this.timelapse.Path,
      'x': this.crop.x,
      'y': this.crop.y,
      'width': this.crop.width,
      'height': this.crop.height,
    };
    this.$.convert.generateRequest();
  }

  onConvertSuccess_(e) {
    this.toast_("Job successfully queued.");
  }

  onConvertError_(e) {
    this.toast_("Job creation failed: " + e.detail.request.xhr.response);
  }

  toast_(msg) {
    this.dispatchEvent(new CustomEvent('toast', {detail: msg, bubbles: true, composed: true}));
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
      crop: {
        type: Object,
      },
      timelapse: {
        type: Object,
      },
      showSelect_: {
        type: Boolean,
        value: false,
      },
      loading_: {
        type: Boolean,
        value: false,
      },
    };
  }

}

window.customElements.define('my-view1', MyView1);
