import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import './shared-styles.js';
import '@polymer/iron-ajax/iron-ajax.js';
import '@polymer/paper-progress/paper-progress.js';
import '@polymer/paper-button/paper-button.js';

class Queue extends PolymerElement {
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
        .item-details {
          display: flex;
          align-items: center;
          flex-wrap: wrap;
        }
        .item-details > div {
          padding: 5px;
        }
        .queue-pending {
          background: #DDDDDD;
        }
        .queue-active {
          background: #FFFFCC;
        }
        .queue-cancel {
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
          id="fetch"
          url="/queue"
          handle-as="json"
          on-response="onResponse_"
          auto></iron-ajax>
      <iron-ajax
          id="op"
          method="POST"
          handle-as="text"
          on-response="onOpSuccess_"
          on-error="onOpError_"
          ></iron-ajax>

      <div class="card">
        <h1>Timelapse Process Queue</h1>
        <div>
          <template is="dom-repeat" items="[[response.Queue]]">
            <div class$="queue-item queue-[[item.State]]">
              <div>
               <img src="/image?path=[[item.Timelapse.Path]]&thumb=true">
              </div>
              <div class="item-details">
                <div>
                    <div>[[item.Config.OutputName]].mp4</div>
                    <div>[[item.Timelapse.Name]]</div>
                    <div>[[getFrames_(item)]] images</div>
                </div>
                <div>
                    <paper-progress value="[[item.Progress]]"></paper-progress>
                </div>
                <div>[[item.Progress]]%</div>
                <div>
                    <div>[[item.State]]</div>
                    <div hidden$="[[!item.ElapsedString]]">[[item.ElapsedString]]</div>
                </div>
                <div hidden$="[[!isState_(item, 'active')]]">
                    <paper-button data-jobid$="[[item.ID]]" data-url="/queue-cancel" data-opname="cancel" on-tap="onOp_" raised>Cancel</paper-button>
                </div>
                <div hidden$="[[isState_(item, 'active', 'cancel')]]">
                    <paper-button data-jobid$="[[item.ID]]" data-url="/queue-remove" data-opname="remove" on-tap="onOp_" raised>Remove</paper-button>
                </div>
                <div hidden$="[[!item.LogPath]]">
                  <a href="/log?path=[[item.LogPath]]" target="_blank">Log</a>
                </div>
              </div>
            </div>
          </template>
          <template is="dom-if" if="[[isEmpty_(response.Queue)]]">
          <div class="emptystate">
            <span>No timelapse jobs queued. Add a timelapse job to start.</span>
          </div>
          </template>
        </div>
      </div>
    `;
  }

  isEmpty_(q) {
    return !q || q.length == 0;
  }

  getFrames_(item) {
    if (!item) {
      return 0;
    }
    return item.Config.EndFrame - item.Config.StartFrame + 1;
  }

  ready() {
    super.ready();
    setInterval(() => this.$.fetch.generateRequest(), 1500);
  }

  onResponse_(e) {
    this.response = e.detail.response;
  }

  isState_(j, ...states) {
    return !!j && states.some(s => j.State === s);
  }

  onOp_(e) {
    const jobid = e.target.dataset.jobid;
    const url = e.target.dataset.url;
    this.opName_ = e.target.dataset.opname;
    this.$.op.url = url;
    this.$.op.headers={'content-type': 'application/x-www-form-urlencoded'};
    this.$.op.body = {'id': jobid};
    this.$.op.generateRequest();

    // Force a state refresh.
    this.$.fetch.generateRequest();
  }

  onOpSuccess_(e) {
    this.toast_("Job " + this.opName_ + " succeeded.");
  }

  onOpError_(e) {
    this.toast_("Failed to " + this.opName_ + " job: " + e.detail.request.xhr.response);
  }

  toast_(msg) {
    this.dispatchEvent(new CustomEvent('toast', {detail: msg, bubbles: true, composed: true}));
  }

  static get properties() {
    return {
      response: {
        type: Object,
      },
      opName_: {
        type: String,
      },
    };
  }
}

window.customElements.define('tq-queue', Queue);
