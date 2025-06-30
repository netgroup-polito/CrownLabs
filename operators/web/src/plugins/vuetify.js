import Vue from 'vue';
import Vuetify from 'vuetify/lib/framework';

import colors from 'vuetify/lib/util/colors'
import '@mdi/font/css/materialdesignicons.css' // Ensure you are using css-loader

Vue.use(Vuetify);

export default new Vuetify({
  theme: {
    options: { customProperties: true },
    themes: {
      light: {
        primary: colors.blue.darken2,
        accent: colors.blue.darken1,
        background: '#f8f3f2',
        legend: {
          cluster: colors.blue.darken1,
          bootstrap: colors.amber.darken2,
          controlplane: colors.purple.darken1,
          infrastructure: colors.green.base,
          addons: colors.red.darken1,
          virtual: colors.grey.darken1,
        },
        // TODO: adjust color for others and key names
        legendTab: {
          cluster: colors.blue.darken2,
          bootstrap: colors.amber.darken3,
          controlplane: colors.purple.darken2,
          infrastructure: colors.green.darken1,
          addons: colors.red.darken2,
          virtual: colors.grey.darken2,
        },
        legendTabHover: {
          cluster: colors.blue.base,
          bootstrap: colors.amber.darken1,
          controlplane: colors.purple.base,
          infrastructure: colors.green.lighten1,
          addons: colors.red.base,
          virtual: colors.grey.base,
        }
      },
      dark: {
        primary: colors.blue.lighten3,
        background: '#121212',
        info: colors.blue.lighten2,
        accent: colors.blue.lighten2,
        success: colors.green.lighten3,
        warning: colors.orange.lighten2, // OG orange darken 1
        error: colors.red.accent1,
        legend: {
          cluster: colors.blue.lighten3,
          bootstrap: colors.amber.lighten3,
          controlplane: colors.purple.lighten3,
          infrastructure: colors.green.lighten3,
          addons: colors.red.lighten2,
          virtual: colors.grey.lighten1,
        }
      }
    }
  },
});
