<template>
  <div class="wrapper">
    <!-- :title="'Cluster Resources: ' + this.$route.query.namespace + '/' + this.$route.query.name" -->
    <AppBar
      :title="'Cluster: ' + getTitle()"
      :showBack="true"
      :scale="Math.round(scale*100) + '%'"
      @togglePathStyle="linkHandler"
      @reload="() => { fetchCluster(forceRedraw=true); fetchCRD(selected, true) }"
      @zoomIn="() => { $refs.targetTree.$refs.tree.zoomIn() }"
      @zoomOut="() => { $refs.targetTree.$refs.tree.zoomOut() }"
      @showSettings="() => { showSettingsOverlay = true }"
    />
    <div
      id="chartLoadWrapper"
      v-if="treeIsReady"
    >
      <DescribeClusterTree
        id="targetTree"
        ref="targetTree"
        :treeConfig="treeConfig"
        :treeData="treeData"
        :isStraight="store.straightLinks"
        :legend="legend"
        @selectNode="fetchCRD"
        @scale="(val) => { scale = val }"
        :style="{
          height: Object.keys(selected).length == 0 ? '100%' : (hasConditions() ? 'calc(100% - 84px - 40px)' : 'calc(100% - 84px)') 
          // height of card title + card subtitle/chip group
        }"
      />

      <div
        class="resourceView"
        v-if="resourceIsReady && this.selected.name"
      >
        <CustomResourceDefinition
          :items="treeviewResource"
          :jsonItems="resource"
          :name="selected.kind + '/' + selected.name"
          :color="$vuetify.theme.themes[theme].legend[selected.provider]"
          @unselectNode="(val) => { this.selected=val; }"
        />
      </div>
    </div>
    <div
      id="resourceTree"
      class="spinner"
      v-else
    >
      <v-progress-circular
        :size="50"
        :width="5"
        indeterminate
        color="primary"
      ></v-progress-circular>
    </div>
    <AlertMessage
      type="error"
      v-model="alert"
      :message="errorMessage"
    />

    <v-overlay
      class="overlay"
      :value="showSettingsOverlay"
      z-index="99999"
      light
    >
      <SettingsCard
        @close="() => { showSettingsOverlay = !showSettingsOverlay }"
        class="settingsCard"
        :version="gitVersion"
      />
    </v-overlay>
    <ScrollButton />
  </div>
</template>

<script>
/* eslint-disable */
import Vue from "vue";
import DescribeClusterTree from "../components/DescribeClusterTree.vue";
import AppBar from "../components/AppBar.vue";
import CustomResourceDefinition from "../components/CustomResourceDefinition.vue";
import AlertMessage from "../components/AlertMessage.vue";
import SettingsCard from "../components/SettingsCard.vue";
import ScrollButton from "../components/ScrollButton.vue";

import { useSettingsStore } from "../stores/settings.js";
import { setVersion } from "../mixins/setVersion.js";

import _ from "lodash";
import colors from "vuetify/lib/util/colors";

export default {
  name: "DescribeCluster",
  components: {
    DescribeClusterTree,
    AppBar,
    SettingsCard,
    CustomResourceDefinition,
    AlertMessage,
    ScrollButton,
  },
  mixins: [setVersion],
  data() {
    return {
      showSettingsOverlay: false,
      showAboutOverlay: false,
      alert: false,
      errorMessage: "",
      treeIsReady: false,
      resourceIsReady: false,
      resource: {},
      selected: {},
      treeData: {},
      cachedTreeString: "",
      treeConfig: { nodeWidth: 180, nodeHeight: 90, levelHeight: 140 },
      scale: 1,
      legend: {
        cluster: "Cluster API",
        bootstrap: "Bootstrap Provider",
        controlplane: "Control Plane Provider",
        infrastructure: "Infrastructure Provider",
        addons: "Add-ons",
        virtual: "None",
      },
      gitVersion: ""
    };
  },
  setup() {
    const store = useSettingsStore();
    return { store };
  },
  async beforeMount() {
    await this.fetchCluster();
    await this.fetchVersion();
  },
  computed: {
    theme() {
      return this.$vuetify.theme.dark ? "dark" : "light";
    },
  },
  mounted() {
    document.title = "Cluster: " + this.getTitle();
    this.intervalHandler(this.store.selectedInterval);
  },
  beforeDestroy() {
    this.selected = {};
    clearInterval(this.polling);
  },
  watch: {
    "store.selectedInterval": function (val) {
      console.log("DescribeCluster store.selectedInterval: " + val);
      this.intervalHandler(val);
    },
  },
  methods: {
    getTitle() {
      let namespace = this.$route.query.namespace;
      let name = this.$route.query.name;
      if (namespace != "" && namespace != "default") {
        return namespace + "/" + name;
      }
      return name;
    },
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
          this.fetchCluster();
          if (Object.keys(this.selected).length > 0) {
            this.fetchCRD(this.selected, true);
          }
        }.bind(this),
        totalSeconds * 1000
      );
    },
    linkHandler(val) {
      this.isStraight = val;
    },
    async fetchCRD(node, closeOnFailure = false) {
      if (!node.name || !node.kind || !node.group || !node.version) {
        console.log("Node missing required fields:", node);
        if (closeOnFailure) {
          this.resourceIsReady = false;
          this.selected = {}; // TODO: do we want to reset the selected variable or is "this.resourceIsReady = false" enough?
        }
        return;
      }
      try {
        // TODO: refresh selected node view along with cluster tree
        // TODO: fetch tree view using kubectl client instead of clusterctl
        const params = new URLSearchParams();
        params.append("kind", node.kind);
        params.append("apiVersion", node.group + "/" + node.version);
        params.append("name", node.name);
        params.append("namespace", node.namespace);

        const response = await Vue.axios.get("/custom-resource-definition", {
          params: params,
        });
        console.log("Response is", response.data);
        this.resource = response.data;
        this.treeviewResource = this.formatToTreeview(response.data);
        this.selected = node; // Don't select until an error won't pop up
        this.resourceIsReady = true;
      } catch (error) {
        console.log("Error:", error.toJSON());
        this.alert = true;
        if (closeOnFailure) {
          this.resourceIsReady = false;
          this.selected = {}; // TODO: do we want to reset the selected variable or is `this.resourceIsReady = false` enough?
        }

        if (error.response) {
          if (error.response.status == 404) {
            this.errorMessage =
              "Cluster Resource `" +
              node.kind +
              "/" +
              node.name +
              "` not found";
          } else {
            this.errorMessage =
              "Unable to load Cluster Resource `" +
              node.kind +
              "/" +
              node.name +
              "`";
          }
        } else if (error.request) {
          this.errorMessage = "No server response received";
        } else {
          this.errorMessage = "Unable to create request";
        }
      }
    },
    async fetchCluster(forceRedraw = false) {
      console.log("Query params are ", this.$route.query);
      console.log("Other params are ", this.$route.params);
      try {
        // const response = await getCluster(this.$route.params.id);
        const params = new URLSearchParams();
        params.append("name", this.$route.query.name);
        params.append("namespace", this.$route.query.namespace);

        const response = await Vue.axios.get("/describe-cluster", {
          params: params,
        });
        // const response = await Vue.axios.get(
        //   "/cluster-resources/" + this.$route.params.id
        // );

        console.log("Target cluster data:", response.data);
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
              "Cluster `" + this.$route.params.id + "` not found";
          } else {
            this.errorMessage =
              "Failed to construct object tree for cluster `" +
              this.$route.params.id +
              "`";
          }
        } else if (error.request) {
          this.errorMessage = "No server response received";
        } else {
          this.errorMessage = "Unable to create request";
        }
      }
    },
    formatToTreeview(resource, path = "") {
      let result = [];
      if (typeof resource == "string") {
        return [{ name: resource }];
      } else if (Array.isArray(resource)) {
        let children = [];
        resource.forEach((e, i) => {
          result.push({
            id: path + "[" + i + "]",
            name: i.toString() + ":", // Add colon for when we are just showing the index.
            children: this.formatToTreeview(e, path + "[" + i + "]"),
          });
        });
      } else {
        // isObject
        Object.entries(resource).forEach(([key, value]) => {
          let name = "";
          let children = [];
          if (typeof value == "string" || typeof value == "number") {
            name = key + ": " + value;
          } else {
            name = key + ":"; // Add colon for when the value is an object.
            children = this.formatToTreeview(value, path + "." + key);
          }
          result.push({
            id: path + "." + key,
            name: name,
            children: children,
          });
        });
      }

      return result;
    },
    hasConditions() {
      return (
        this.resource?.status?.conditions != undefined &&
        this.resource?.status?.conditions.length > 0
      );
    },
  },
};
</script>

<style lang="less" scoped>
.wrapper {
  height: 100%;
  width: 100%;
  max-width: 100%;
  margin: 0 !important;
}

#chartLoadWrapper {
  height: 100%;

  #treeChartWrapper {
    width: 100%;
    height: 100%;
    position: relative;
    text-align: center;
  }
}

.machine {
  position: absolute;
  transform: translate(0, 65px);
  width: 375px;
  height: 230px;
  border: 3px solid #1e88e5;
  // border: 3px solid #a8c8ff;
  box-shadow: 3px 4px 3px rgba(0, 0, 0, 0.3);
  border-radius: 5px;
  z-index: -10000;

  span {
    position: absolute;
    bottom: 5px;
    right: 10px;
  }
}

.resourceView {
  margin: 0 30px;
  padding-bottom: 30px;
}
</style>

<style lang="less">
.overlay {
  position: fixed;
}

.spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
}
</style>
