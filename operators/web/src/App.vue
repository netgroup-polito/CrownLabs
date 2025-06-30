<template>
  <v-app
    id="app"
    :style="{background: $vuetify.theme.themes[theme].background}"
  >
    <v-main id="main">
      <router-view />
    </v-main>
  </v-app>
</template>

<script>
import { useSettingsStore } from "./stores/settings.js";

export default {
  name: "App",
  computed: {
    theme() {
      return this.$vuetify.theme.dark ? "dark" : "light";
    },
  },
  setup() {
    const store = useSettingsStore();
    return { store };
  },
  mounted() {
    // Initialize dark theme from store
    this.$vuetify.theme.dark = this.store.darkTheme;
  },
};
</script>

<style>
html,
body {
  height: 100%;
  width: 100%;
  padding: 0;
  margin: 0;
  /* background-color: #f8f3f2; */
  background-color: var(--v-background-base);
}

#app {
  font-family: Roboto, serif;
  /* text-align: center; */
  height: 100%;
  /* height: -webkit-fill-available; */
}

#main {
  height: 100%;
  /* background-color: #f8f3f2; */
}
</style>