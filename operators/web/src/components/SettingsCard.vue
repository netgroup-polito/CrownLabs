<template>
  <v-card
    class="py-1"
    min-width="500px"
    elevation="6"
    :light="!$vuetify.theme.dark"
  >

    <v-card-title class="ml-2">
      Settings
      <v-spacer></v-spacer>
      <v-btn
        icon
        @click="() => { this.$emit('close', {}); }"
      >
        <v-icon>mdi-window-close</v-icon>
      </v-btn>

    </v-card-title>

    <v-card-text class="py-0 ml-2">
      <div class="text-subtitle-2">About</div>
    </v-card-text>
    <v-list rounded>
      <v-list-item>
        <v-list-item-icon>
          <v-icon>mdi-information</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Version</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          {{ version }}
        </v-list-item-action>
      </v-list-item>
      <v-list-item
        href="https://github.com/Jont828/cluster-api-visualizer"
        target="_blank"
      >
        <v-list-item-icon>
          <v-icon>mdi-github</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Source code</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          <v-icon>mdi-open-in-new</v-icon>
        </v-list-item-action>
      </v-list-item>
    </v-list>

    <v-card-text class="py-0 ml-2">
      <div class="text-subtitle-2">Appearance</div>
    </v-card-text>
    <v-list rounded>
      <v-list-item>
        <v-list-item-icon>
          <v-icon>mdi-brightness-6</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Dark theme</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          <v-switch
            v-model="store.darkTheme"
            @change="toggleDarkTheme"
          >
          </v-switch>
        </v-list-item-action>
      </v-list-item>

      <v-list-item>
        <v-list-item-icon>
          <v-icon>mdi-sitemap-outline</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Straighten links</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          <v-switch v-model="store.straightLinks">
          </v-switch>
        </v-list-item-action>
      </v-list-item>
    </v-list>

    <v-card-text class="py-0 ml-2">
      <div class="text-subtitle-2">General</div>
    </v-card-text>
    <v-list rounded>
      <v-list-item>
        <v-list-item-icon>
          <v-icon>mdi-file-download</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Download Format</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          <v-select
            class="selectBox"
            v-model="store.selectedFileType"
            :items="fileTypes"
            dense
            hide-details
          >
          </v-select>
        </v-list-item-action>
      </v-list-item>

      <v-list-item>
        <v-list-item-icon>
          <v-icon>mdi-timer-sync</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Polling period</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          <v-select
            class="selectBox"
            :items="pollingInterval"
            v-model="store.selectedInterval"
            dense
            hide-details
          >
          </v-select>
        </v-list-item-action>
      </v-list-item>

      <v-list-item class="listRow">
        <v-list-item-icon>
          <v-icon>mdi-file-document</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Max Log Lines</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action>
          <div class="relWrap">
            <v-text-field
              :rules="rules"
              v-model="maxLogLines"
              height="25"
              class="textInput"
            ></v-text-field>
          </div>
        </v-list-item-action>
      </v-list-item>
    </v-list>
  </v-card>

</template>

<script>
import Vue from "vue";

import { useSettingsStore } from "../stores/settings.js";

export default {
  name: "SettingsCard",
  components: {},
  props: {
    version: String,
  },
  watch: {
    maxLogLines: function() {
      if (this.positiveInteger.test(this.maxLogLines)) {
        this.store.maxLogLines = this.maxLogLines;
      }
    }
  },
  methods: {
    toggleDarkTheme(val) {
      this.$vuetify.theme.dark = val;
    },
    // TODO: Do this once on page load so this doesn't take a minute and lag.
    ensurePositiveInt(val) {
      if (this.positiveInteger.test(val)) {
        return true;
      } else {
        return "Must be a positive integer.";
      }
    },
  },
  setup() {
    const store = useSettingsStore();

    return { store };
  },
  data() {
    return {
      positiveInteger: /^[1-9]\d*$/,
      maxLogLines: this.store.maxLogLines,
      fileTypes: ["YAML", "JSON"],
      pollingInterval: ["1s", "5s", "10s", "30s", "1m", "5m", "Off"],
      rules: [
        value => {
          const pattern = /^[1-9]\d*$/
          return pattern.test(value) || 'Must be positive integer'
        },
      ],
    };
  },
};
</script>

<style lang="less" scoped>
.selectBox {
  width: 100px;
}

.textInput {
  position: absolute;
  top: 0;
  width: 100px;
}

.relWrap {
  width: 100px;
  height: 30px;
  position: relative;
}
</style>