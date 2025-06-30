<template>
  <v-app-bar
    id="appBar"
    app
    :color="($vuetify.theme.dark ? null: 'primary')"
    dark
  >
    <v-btn
      icon
      text
      class="ma-2"
      @click="() => { router.back() }"
      v-if="showBack"
    >
      <v-icon color="white">
        mdi-chevron-left
      </v-icon>
    </v-btn>
    <v-app-bar-nav-icon
      class="ma-2"
      v-else
    ></v-app-bar-nav-icon>
    <v-toolbar-title class="text-no-wrap pa-0">{{ title }}</v-toolbar-title>
    <v-tooltip bottom>
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          icon
          text
          class="ma-2"
          @click="$emit('reload', true)"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon color="white">
            {{"mdi-refresh"}}
          </v-icon>
        </v-btn>
      </template>
      <span>Reload resources</span>
    </v-tooltip>

    <v-spacer></v-spacer>
    <v-tooltip bottom>
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          icon
          text
          class="ma-2"
          @click="$emit('zoomOut', true)"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon color="white">
            {{"mdi-minus"}}
          </v-icon>
        </v-btn>
      </template>
      <span>Zoom out</span>
    </v-tooltip>
    <v-icon
      v-if="scaleIcon"
      color="white"
    > mdi-{{ scaleIcon }}</v-icon>
    <span v-else>{{ scale }}</span>
    <v-tooltip bottom>
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          icon
          text
          class="ma-2"
          @click="$emit('zoomIn', true)"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon color="white">
            {{"mdi-plus"}}
          </v-icon>
        </v-btn>
      </template>
      <span>Zoom in</span>
    </v-tooltip>
    <v-tooltip bottom>
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          icon
          text
          class="ma-2"
          @click="$emit('showSettings', true)"
          v-bind="attrs"
          v-on="on"
        >
          <!-- TODO: should it be showSettings or toggleSettings, i.e. should clicking again close the overlay -->
          <v-icon color="white">
            mdi-cog
          </v-icon>
        </v-btn>
      </template>
      <span>Show settings</span>
    </v-tooltip>

  </v-app-bar>
</template>

<script>
import router from '../router';

export default {
  name: "AppBar",
  props: {
    title: String,
    showBack: Boolean,
    isStraight: Boolean,
    scale: String,
    scaleIcon: String,
    backURL: String,
  },
  data() {
    return {
      router: router,
    };
  },
};
</script>

<style lang="less" scoped>
#appBar {
  z-index: 2000;
}

.router-link {
  text-decoration: none;
}
</style>