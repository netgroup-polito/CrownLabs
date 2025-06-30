<template>
  <div id="overview">
    <AppBar
      title="Management Cluster"
      :scale="Math.round(scale*100) + '%'"
      @togglePathStyle="linkHandler"
      @reload="fetchOverview(forceRedraw=true)"
      @zoomIn="() => { $refs.overviewTree.$refs.tree.zoomIn() }"
      @zoomOut="() => { $refs.overviewTree.$refs.tree.zoomOut() }"
      @showSettings="() => { showSettingsOverlay = true }"
    />
    <ManagementClusterTree
      ref="overviewTree"
      :treeConfig="treeConfig"
      :treeData="treeData"
      :treeIsReady="treeIsReady"
      @scale="(val) => { scale = val }"
    />

    <v-overlay
      absolute
      :value="showSettingsOverlay"
      z-index="99999"
    >
      <SettingsCard
        @close="() => { showSettingsOverlay = !showSettingsOverlay }"
        class="settingsCard"
        :version="gitVersion"
      />
    </v-overlay>
    <AlertMessage
      type="error"
      v-model=alert
      :message="errorMessage"
    />
  </div>
</template>

<script>
import Vue from "vue";

import ManagementClusterTree from "../components/ManagementClusterTree.vue";
import SettingsCard from "../components/SettingsCard.vue";
import AppBar from "../components/AppBar.vue";
import AlertMessage from "../components/AlertMessage.vue";

import { useSettingsStore } from "../stores/settings.js";
import { setVersion } from "../mixins/setVersion.js";

export default {
  name: "ManagementCluster",
  components: {
    ManagementClusterTree,
    SettingsCard,
    AlertMessage,
    AppBar,
  },
  mixins: [setVersion],
  data() {
    return {
      showSettingsOverlay: false,
      treeConfig: { nodeWidth: 300, nodeHeight: 140, levelHeight: 275 },
      treeData: {},
      cachedTreeString: "",
      treeIsReady: false,
      scale: 1,
      alert: false,
      errorMessage: "",
    };
  },
  setup() {
    const store = useSettingsStore();
    return { store };
  },
  async beforeMount() {
    await this.fetchOverview();
  },
  computed: {
    theme() {
      return this.$vuetify.theme.dark ? "dark" : "light";
    },
  },
  mounted() {
    document.title = "Management Cluster";
    this.intervalHandler(this.store.selectedInterval);
  },
  beforeDestroy() {
    this.selected = {};
    clearInterval(this.polling);
  },
  watch: {
    "store.selectedInterval": function (val) {
      console.log("Overview store.selectedInterval: " + val);
      this.intervalHandler(val);
    },
  },
  methods: {
    intervalHandler(val) {
      console.log("Setting polling interval to " + val);
      clearInterval(this.polling);
      if (val === "Off") return;

      let totalSeconds = 0;

      let seconds = val.match(/(\d+)\s*s/);
      let minutes = val.match(/(\d+)\s*m/);

      if (seconds) {
        totalSeconds += parseInt(seconds[1]);
      }
      if (minutes) {
        totalSeconds += parseInt(minutes[1]) * 60;
      }

      console.log("Setting interval to " + totalSeconds + " seconds");
      this.polling = setInterval(
        function () {
          this.fetchOverview();
        }.bind(this),
        totalSeconds * 1000
      );
    },
    linkHandler(val) {
      this.isStraight = val;
    },
    async fetchOverview(forceRedraw = false) {
      try {
        const response = await Vue.axios.get("/management-cluster");

        if (response.data == null) {
          this.errorMessage = "Couldn't find a management cluster from default kubeconfig";
          return;
        }

        console.log("Cluster overview data:", response.data);
        if (
          forceRedraw ||
          this.cachedTreeString !== JSON.stringify(response.data)
        ) {
          this.treeData = response.data;
          this.cachedTreeString = JSON.stringify(response.data);
          this.treeIsReady = true;
        }
      } catch (error) {
        console.log("Error:", error.toJSON());
        this.alert = true;
        if (error.response) {
          if (error.response.status == 404) {
            this.errorMessage =
              "Management cluster not found, is the kubeconfig set?";
          } else {
            this.errorMessage =
              "Unable to load management cluster and workload clusters";
          }
        } else if (error.request) {
          this.errorMessage = "No server response received";
        } else {
          this.errorMessage = "Unable to create request";
        }
      }
    },
  },
};
</script>

<style lang="less" scoped>
#overview {
  height: 100%;
}
</style>