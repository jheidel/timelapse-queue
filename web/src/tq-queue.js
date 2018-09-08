import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import './shared-styles.js';
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
        <div class="circle">2</div>
        <h1>Process Queue</h1>
        <div>
          <template is="dom-repeat" items="[[response.Queue]]">
            <div class$="queue-item queue-[[item.State]]">
              <div>
               <img src="/image?path=[[item.Timelapse.Path]]&thumb=true">
              </div>
              <div>
                  <div>[[item.Timelapse.Name]]</div>
                  <div>[[item.Timelapse.Count]] images</div>
                  <div>[[item.Timelapse.DurationString]]</div>
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
                  <paper-button data-jobid$="[[item.ID]]" data-url="/queue-cancel" on-tap="onOp_" raised>Cancel</paper-button>
              </div>
              <div hidden$="[[isState_(item, 'active')]]">
                  <paper-button data-jobid$="[[item.ID]]" data-url="/queue-remove" on-tap="onOp_" raised>Remove</paper-button>
              </div>
              <div hidden$="[[!item.LogPath]]">
                <a href="/log?path=[[item.LogPath]]" target="_blank">Log</a>
              </div>
            </div>
          </template>
          <template is="dom-if" if="[[!response.Queue]]">
          <div class="emptystate">
            <span>No timelapse jobs queued. Add a timelapse job to start.</span>
          </div>
          </template>
        </div>
      </div>
    `;
  }

  onResponse_(e) {
    this.response = e.detail.response;
    // Poll again after short delay.
    setTimeout(() => this.$.fetch.generateRequest(), 1500);
  }

  isState_(j, state) {
    return !!j && j.State === state;
  }

  onOp_(e) {
    const jobid = e.target.dataset.jobid;
    const url = e.target.dataset.url;
    this.$.op.url = url;
    this.$.op.headers={'content-type': 'application/x-www-form-urlencoded'};
    this.$.op.body = {'id': jobid};
    this.$.op.generateRequest();
  }

  onOpSuccess_(e) {
    this.toast_("Job change request succeeded.");
  }

  onOpError_(e) {
    this.toast_("Job change request failed: " + e.detail.request.xhr.response);
  }

  toast_(msg) {
    this.dispatchEvent(new CustomEvent('toast', {detail: msg, bubbles: true, composed: true}));
  }

  static get properties() {
    return {
      response: {
        type: Object,
      },
    };
  }
}

window.customElements.define('tq-queue', Queue);
