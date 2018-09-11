import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-ajax/iron-ajax.js';
import '@polymer/iron-collapse/iron-collapse.js';
import '@polymer/iron-icon/iron-icon.js';
import '@polymer/iron-icons/iron-icons.js';
import '@polymer/paper-button/paper-button.js';
import '@polymer/paper-checkbox/paper-checkbox.js';
import '@polymer/paper-dropdown-menu/paper-dropdown-menu.js';
import '@polymer/paper-input/paper-input.js';
import '@polymer/paper-item/paper-item.js';
import '@polymer/paper-listbox/paper-listbox.js';
import '@polymer/paper-slider/paper-slider.js';
import '@polymer/paper-spinner/paper-spinner.js';
import './shared-styles.js';
import Croppr from 'croppr/src/croppr.js';

class Setup extends PolymerElement {
  static get template() {
    return html`
      <link rel="stylesheet" href="../node_modules/croppr/src/css/croppr.css">
      <style include="shared-styles">
        :host {
          display: block;

          padding: 10px;
        }
        .constrain-width {
          max-width: 800px;
        }
        .short-input {
          max-width: 200px;
        }
        .medium-input {
          max-width: 300px;
        }
        .slider paper-slider {
          --paper-slider-input: {
             width: 650px;
          }
          width: 650px;
        }
        .slider {
          display: flex;
          align-items: center;
        }
        .slider > span {
          width: 85px;

        }
        .helptext {
          color: gray;
          font-size: small;
        }
        .startbutton {
          padding-top: 20px;
        }
        .error {
          color: red;
          padding-bottom: 15px;
        }
        .stack-options {
          padding-left: 20px;
        }
        .stack-option {
          padding-top: 20px;
        }
        .infobox {
          padding: 3px;
          background-color: #f4f4f4;
          display: inline-block;
        }
      </style>

      <iron-ajax
          id="convertajax"
          url="/convert"
          method="POST"
          handle-as="text"
          on-response="onConvertSuccess_"
          on-error="onConvertError_"
          ></iron-ajax>
      <iron-ajax
          id="timelapseajax"
          url="/timelapse"
          handle-as="json"
          on-response="onTimelapseAjax_"
          ></iron-ajax>

      <div class="card">
        <div class="circle">2</div>
        <h1>Configure Timelapse Job</h1>
        <p>
          <div>Input Timelapse</div>
          <div class="helptext infobox">
             <div>[[timelapse.Name]]</div>
             <div>[[timelapse.Count]] frames</div>
             <div>[[timelapse.DurationString]] (at 60fps)</div>
          </div>
        </p>

        <p class="medium-input">
          <paper-input
                  label="Output Filename"
                  value="{{filename_}}"
                  always-float-label 
                  auto-validate
                  pattern="[a-zA-Z0-9-_ ]+"
                  error-message="Not a valid filename"
                  autofocus
                  >
            <span slot="suffix">.mp4</span>
          </paper-input>
        </p>

        <p>
          <div class="helptext">
           <div>The output MP4 framerate can be adjusted.</div>
           <div>60fps produces smooth video.</div>
           <div>30fps can be used to extend the length of the timelapse.</div>
          </div>
          <paper-dropdown-menu label="Output Framerate" no-animations>
            <paper-listbox attr-for-selected="value" selected="{{fps_}}" slot="dropdown-content">
              <paper-item value="30">30 fps</paper-item>
              <paper-item value="60">60 fps</paper-item>
            </paper-listbox>
          </paper-dropdown-menu>
        </p>


        <p>
          <div>Output Video File</div>
          <div class="helptext infobox">
            <div>MP4 1920x1080 [[fps_]] fps</div>
            <div hidden$="[[!filename_]]">[[timelapse.ParentPath]][[filename_]].mp4</div>
          </div>
        </p>

        <p>
        <div class="helptext">
          Adjust the start and end frame positions to control the length of the timelapse.
        </div>
        <div class="slider">
          <span>Start Frame</span>
         <paper-slider min="0" max="[[getLastFrame_(timelapse)]]" value="{{startFrame_}}" pin></paper-slider>
          <paper-input
                type="number"
                min="0"
                max="[[getLastFrame_(timelapse)]]"
                value="{{startFrame_}}"
                no-label-float
            ></paper-input> 
        </div>
        <div class="slider">
          <span>End Frame</span>
          <paper-slider min="0" max="[[getLastFrame_(timelapse)]]" value="{{endFrame_}}" pin></paper-slider>
          <paper-input
                type="number"
                min="0"
                max="[[getLastFrame_(timelapse)]]"
                value="{{endFrame_}}"
                no-label-float
            ></paper-input> 
        </div>
        </p>

        <p>
          <div>Select Image Region</div>
          <div hidden$="[[!loading_]]">
              <paper-spinner active="[[loading_]]"></paper-spinner>
          </div>
          <div class="constrain-width" id="container">
          </div>
          <div class="helptext">
            <span>x=[[crop.x]] y=[[crop.y]]</span>
            <span>Size [[crop.width]]x[[crop.height]]</span>
          </div>
        </p>

        <p>
          <div>
                  <paper-checkbox checked="{{stack_}}">
                    Photo Stacking
                  </paper-checkbox>
          </div>
          <div>
                  <iron-collapse opened="[[stack_]]">
                  <div class="stack-options">
                        <div class="stack-option">
                          <div class="helptext">
                            <div>The frame stack count controls how long the stacking tail is.</div>
                            <div>At 60fps, a value of "60" will give one second of history.</div>
                            <div>Larger values will give a longer tail, though will take longer to process.</div>
                          </div>
                          <paper-input
                                class="short-input"
                                label="Frames to Stack"
                                type="number"
                                min="1"
                                max="[[timelapse.Count]]"
                                value="{{stackWindow_}}"
                                disabled="[[stackAll_]]"
                                always-float-label></paper-input>
                          <paper-checkbox checked="{{stackAll_}}">
                            Stack All
                          </paper-checkbox>
                       </div>
                       <div class="stack-option">
                          <div class="helptext">
                            <div>Frame skipping can be used to make the stacking effect more obvious.</div>
                            <div>Increase to create larger gaps between stacked images.</div>
                          </div>
                          <paper-checkbox checked="{{skip_}}" disabled="[[stackAll_]]">
                            Frame Skip
                          </paper-checkbox>
                          <paper-input
                                class="short-input"
                                label="Frame Skip"
                                type="number"
                                min="1"
                                max="[[min_(stackWindow_, timelapse.Count)]]"
                                value="{{skipCount_}}"
                                disabled="[[!skip_]]"
                                always-float-label></paper-input>
                       </div>
                       <div class="stack-option">
                          <div class="helptext">
                            <div>The blending mode can be changed to achieve different stacking effects.</div>
                          </div>
                          <!-- animations broken in polymer 3.0, disabled for now -->
                          <paper-dropdown-menu label="Blending Mode" no-animations>
                            <paper-listbox attr-for-selected="value" selected="{{stackMode_}}" slot="dropdown-content">
                              <paper-item value="lighten">Lighten</paper-item>
                              <paper-item value="darken">Darken</paper-item>
                            </paper-listbox>
                          </paper-dropdown-menu>
                       </div>
                  </div>
                  </iron-collapse>
          </div>
        </p>

        <p>
          <div>Advanced Options</div>
          <div class="helptext">
            These options may be useful for debugging performance issues.
          </div>
          <div>
                  <paper-checkbox id="profilecpu">
                    CPU Profiling
                  </paper-checkbox>
          </div>
          <div>
                  <paper-checkbox id="profilemem">
                    Memory Profiling
                  </paper-checkbox>
          </div>
        </p>

        <div class="startbutton">
            <div class="error" hidden$="[[!error_]]">
              <iron-icon icon="error"></iron-icon>
              [[error_]]
            </div>
            <paper-button on-tap="onConvert_" raised>
                   <iron-icon icon="schedule"></iron-icon>
                  Add Timelapse Job to Queue
            </paper-button>
        </div>
      </div>
    `;
  }

  onTimelapseAjax_(e) {
    const resp = e.detail.xhr.response;
    if (!resp) {
            return;
    }
    this.timelapse = resp;
  }

  onFrame_(frame) {
      if (!this.croppr || !this.enableObservers_ || !this.path || !frame) {
          return;
      }
      this.croppr.setImage('/image?path=' + this.path + '&index=' + frame);
  }
 
  or_(a, b) {
          return a || b;
  }

  min_(a, b) {
    if (!a || !b) {
      return 0;
    }
    return a < b ? a : b;
  }

  getLastFrame_(tl) {
          if (!tl) {
                  return 0;
          }
          return tl.Count - 1;
  }

  onTimelapse_(tl) {
    this.startFrame_ = 0;
    this.endFrame_ = this.getLastFrame_(tl);
  }

  onPath_(path) {
      if (!path) {
              return
      }

      this.$.timelapseajax.params = {'path': path};
      this.$.timelapseajax.generateRequest();

      this.loading_ = true;

      const container = this.$.container;
      // Remove any existing elements left behind by croppr.
      while (container.firstChild) {
              container.removeChild(container.firstChild);
      }

      // Add new image.
      const img = document.createElement('img');
      img.classList.add("constrain-width");
      img.src = '/image?path=' + path;
      container.appendChild(img);

      // TODO set these based on job configuration.
      const width = 1920;
      const height = 1080;

      this.croppr = new Croppr(img, {
              aspectRatio: height / width,
              startSize: [100, 100, '%'],
              // TODO doesn't work when the canvas is scaled.
              //minSize: [width, height, 'px'],
              onCropMove: (value) => {
                this.crop = value;
              },
              onCropEnd: (value) => {
                this.crop = value;
              },
              onInitialize: (instance) => {
                this.crop = instance.getValue();
                this.loading_ = false;
                this.enableObservers_ = true;
              },
      });
  }     
  
  onConvert_(e) {
    this.$.convertajax.headers={'content-type': 'application/x-www-form-urlencoded'};
    const config = {
      'Path': this.timelapse.Path,
      'X': this.crop.x,
      'Y': this.crop.y,
      'Width': this.crop.width,
      'Height': this.crop.height,
      'OutputName': this.filename_,
      'FrameRate': parseInt(this.fps_, 10),
      'StartFrame': this.startFrame_,
      'EndFrame': this.endFrame_,
      'Stack': this.stack_,
      'StackWindow': this.stackAll_ ? 0 : parseInt(this.stackWindow_, 10),
      'StackSkipCount': this.skip_ ? parseInt(this.skipCount_, 10) : 0,
      'StackMode': this.stackMode_,
    };
    if (this.$.profilecpu.checked) {
      config['ProfileCPU'] = true;
    }
    if (this.$.profilemem.checked) {
      config['ProfileMem'] = true;
    }

    this.$.convertajax.body = {
        'request': JSON.stringify(config),
    };
    this.$.convertajax.generateRequest();
  }

  onConvertSuccess_(e) {
    // Redirect to queue to see the new job.
    window.history.pushState({}, null, '/#/queue');
    window.dispatchEvent(new CustomEvent('location-changed'));

    this.toast_("Job successfully queued.");
    this.error_ = "";
    this.filename_ = "";
    this.startFrame_ = 0;
    this.endFrame_ = 0;
    this.stack_ = false;
    this.skip_ = false;
    this.croppr.destroy();
  }

  onConvertError_(e) {
    this.toast_("Job creation failed."); 
    this.error_ = e.detail.request.xhr.response;
  }

  toast_(msg) {
    this.dispatchEvent(new CustomEvent('toast', {detail: msg, bubbles: true, composed: true}));
  }


  static get properties() {
    return {
      path: {
        type: String,
        observer: 'onPath_',
        value: '',
      },
      timelapse: {
        type: Object,
        observer: 'onTimelapse_',
      },
      crop: {
        type: Object,
      },
      loading_: {
        type: Boolean,
        value: false,
      },
      stack_: {
        type: Boolean,
        value: false,
      },
      stackMode_: {
        type: String,
        value: "lighten",
      },
      fps_: {
        type: String,
        value: "60",
      },
      stackAll_: {
        type: Boolean,
        value: false,
      },
      skipCount_: {
        type: Number,
        value: 3,
      },
      skip_: {
        type: Boolean,
        value: false,
      },
      filename_: {
        type: String,
      },
      startFrame_: {
        type: Number,
        observer: 'onFrame_',
      },
      endFrame_: {
        type: Number,
        observer: 'onFrame_',
      },
      stackWindow_: {
        type: Number,
        value: 60,
      },
      enableObservers_: {
        type: Boolean,
        observer: false,
      },
      error_: {
        type: String,
        value: "",
      },
    };
  }
}

window.customElements.define('tq-setup', Setup);
