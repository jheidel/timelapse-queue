import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import './shared-styles.js';

class Setup extends PolymerElement {
  static get template() {
    return html`
      <style include="shared-styles">
        :host {
          display: block;

          padding: 10px;
        }
      </style>

      <div class="card">
        <div class="circle">2</div>
        <h1>Configure Timelapse Job</h1>
        <p>TODO TODO TODO<p>
      </div>
    `;
  }
}

window.customElements.define('tq-setup', Setup);
