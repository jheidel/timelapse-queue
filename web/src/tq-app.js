import { PolymerElement, html } from '@polymer/polymer/polymer-element.js';
import { setPassiveTouchGestures, setRootPath } from '@polymer/polymer/lib/utils/settings.js';
import '@polymer/app-layout/app-drawer-layout/app-drawer-layout.js';
import '@polymer/app-layout/app-drawer/app-drawer.js';
import '@polymer/app-layout/app-header-layout/app-header-layout.js';
import '@polymer/app-layout/app-header/app-header.js';
import '@polymer/app-layout/app-scroll-effects/app-scroll-effects.js';
import '@polymer/app-layout/app-toolbar/app-toolbar.js';
import '@polymer/app-route/app-location.js';
import '@polymer/app-route/app-route.js';
import '@polymer/iron-ajax/iron-ajax.js';
import '@polymer/iron-icon/iron-icon.js';
import '@polymer/iron-icons/iron-icons.js';
import '@polymer/iron-pages/iron-pages.js';
import '@polymer/iron-selector/iron-selector.js';
import '@polymer/paper-icon-button/paper-icon-button.js';
import '@polymer/paper-item/paper-icon-item.js';
import '@polymer/paper-toast/paper-toast.js';
import './tq-icons.js';

import('./tq-queue.js');

// Gesture events like tap and track generated from touch will not be
// preventable, allowing for better scrolling performance.
setPassiveTouchGestures(true);

class TimelapseQueueApp extends PolymerElement {
  static get template() {
    return html`
      <style>
        :host {
          --app-primary-color: #4285f4;
          --app-secondary-color: black;

          display: block;
        }

        app-drawer-layout:not([narrow]) [drawer-toggle] {
          display: none;
        }

        app-header {
          color: #fff;
          background-color: var(--app-primary-color);
        }

        app-header paper-icon-button {
          --paper-icon-button-ink-color: white;
        }

        .drawer-list {
          margin: 0 20px;
        }

        .drawer-list a {
          display: block;
          padding: 0 16px;
          text-decoration: none;
          color: var(--app-secondary-color);
          line-height: 40px;
        }

        .drawer-list a.iron-selected {
          color: black;
          font-weight: bold;
        }

        .buildinfo {
          background-color: white;
          position: absolute;
          bottom: 0px;
          left: 0px;
          padding: 2px;
          font-size: 8pt;
          z-index: 10;
        }
      </style>

      <app-location route="{{route}}" use-hash-as-path></app-location>

      <app-route route="{{route}}" pattern="/:page" data="{{routeData}}" tail="{{subroute}}" query-params="{{queryParams}}">
      </app-route>

      <iron-ajax
          id="buildajax"
          url="/build"
          handle-as="text"
          on-response="onBuild_"
          auto></iron-ajax>
      <div class="buildinfo" hidden$="[[!build_]]">
        <div>Last Software Update:</div>
        <div>[[build_]]</div>
      </div>

      <paper-toast id="toast"></paper-toast>

      <app-drawer-layout fullbleed="" narrow="{{narrow}}">
        <!-- Drawer content -->
        <app-drawer id="drawer" slot="drawer" swipe-open="[[narrow]]">
          <app-toolbar>Menu</app-toolbar>
          <iron-selector selected="{{page}}" attr-for-selected="name" class="drawer-list" role="navigation">
            <paper-icon-item name="browse">
              <iron-icon icon="add" slot="item-icon"></iron-icon>
              <span>Add Timelapse</span>
            </paper-icon-item>
            <paper-icon-item name="queue">
              <iron-icon icon="view-list" slot="item-icon"></iron-icon>
              <span>View Queue</span>
            </paper-icon-item>
          </iron-selector>
        </app-drawer>

        <!-- Main content -->
        <app-header-layout has-scrolling-region="">

          <app-header slot="header" condenses="" reveals="" effects="waterfall">
            <app-toolbar>
              <paper-icon-button icon="tq-icons:menu" drawer-toggle=""></paper-icon-button>
              <div main-title="">Timelapse Queue</div>
            </app-toolbar>
          </app-header>

          <iron-pages selected="[[page]]" attr-for-selected="name" role="main">
            <tq-browse name="browse"></tq-browse>
            <tq-setup name="setup" path="[[queryParams.path]]"></tq-setup>
            <tq-queue name="queue"></tq-queue>
            <tq-view404 name="view404"></tq-view404>
          </iron-pages>
        </app-header-layout>
      </app-drawer-layout>
    `;
  }

  static get properties() {
    return {
      page: {
        type: String,
        reflectToAttribute: true,
        observer: '_pageChanged'
      },
      routeData: Object,
      queryParams: Object,
      subroute: Object,
      build_: {
        type: String,
        value: '',
      },
    };
  }

  static get observers() {
    return [
      '_routePageChanged(routeData.page)'
    ];
  }

  ready() {
          super.ready();
          this.addEventListener('toast', this._onToast);
          // Periodically check server build to prevent stale client.
          setInterval(() => this.$.buildajax.generateRequest(), 10000);
  }

  onBuild_(e) {
          const value = e.detail.xhr.response;
          if (!value) {
                  return;
          }
          if (!!this.build_ && this.build_ != value) {
            // Server has a new build version; hard refresh the client.
            location.reload(true);
            return;
          }
          this.build_ = value;
  }

  _onToast(e) {
          this.$.toast.show(e.detail);
  }

  _routePageChanged(page) {
     // Show the corresponding page according to the route.
    if (!page) {
      this.page = 'browse';
      window.history.pushState({}, null, '/#/browse');
      window.dispatchEvent(new CustomEvent('location-changed'));
    } else if (['browse', 'queue', 'setup'].indexOf(page) !== -1) {
      this.page = page;
    } else {
      this.page = 'view404';
    }
    
    // Close a non-persistent drawer when the page & route are changed.
    if (!this.$.drawer.persistent) {
      this.$.drawer.close();
    }
  }

  _pageChanged(page) {
    this.set('routeData.page', page);

    window.dispatchEvent(new CustomEvent('location-changed'));

    // Dynamically import nodes as we navigate to them.
    // This kind of works I guess?
    switch (page) {
      case 'browse':
        import('./tq-browse.js');
        break;
      case 'queue':
        import('./tq-queue.js');
        break;
      case 'setup':
        import('./tq-setup.js');
        break;
      case 'view404':
        import('./tq-view404.js');
        break;
    }
  }
}

window.customElements.define('tq-app', TimelapseQueueApp);
