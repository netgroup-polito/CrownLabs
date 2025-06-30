import './assets/fonts.css';

import Vue from 'vue'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'

import axios from "axios";
import VueAxios from "vue-axios";

import { createPinia, PiniaVuePlugin } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

import hljs from 'highlight.js/lib/core';
import json from 'highlight.js/lib/languages/json';
import yaml from 'highlight.js/lib/languages/yaml';
import hljsVuePlugin from "@highlightjs/vue-plugin";

import VueVirtualScroller from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'


Vue.use(VueVirtualScroller)

hljs.registerLanguage('json', json);
hljs.registerLanguage('yaml', yaml);

Vue.use(hljsVuePlugin);

Vue.use(PiniaVuePlugin)
const pinia = createPinia()

pinia.use(piniaPluginPersistedstate)

Vue.config.productionTip = false;

const client = axios.create({
  baseURL: "/api/v1",
});
Vue.use(VueAxios, client);

Vue.mixin({
  methods: {
    // Negative amt is to darken, positive amount to lighten
    adjustColor(col, amt) {
      var usePound = false;
      if (col[0] == "#") {
        col = col.slice(1);
        usePound = true;
      }
      var num = parseInt(col, 16);
      var r = (num >> 16) + amt;
      if (r > 255) r = 255;
      else if (r < 0) r = 0;
      var b = ((num >> 8) & 0x00ff) + amt;
      if (b > 255) b = 255;
      else if (b < 0) b = 0;
      var g = (num & 0x0000ff) + amt;
      if (g > 255) g = 255;
      else if (g < 0) g = 0;

      return (usePound ? "#" : "") + (g | (b << 8) | (r << 16)).toString(16);
    },
  }
});

Vue.config.productionTip = false

new Vue({
  router,
  vuetify,
  pinia,
  render: h => h(App)
}).$mount('#app')
