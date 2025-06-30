<template>
  <v-card class="resource-card mx-auto">
    <link
      v-if="$vuetify.theme.dark"
      rel="stylesheet" 
      href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/styles/stackoverflow-dark.min.css"
    >
    <link
      v-else
      rel="stylesheet" 
      href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/styles/stackoverflow-light.min.css"
    >

    <v-sheet
      :color="($vuetify.theme.dark) ? '#272727' : color"
      class="resourceSheet pa-4"
    >
      <v-card-title
        class="text-h5"
        :style="{
          color: ($vuetify.theme.dark) ? color : 'white'
        }"
      >
        {{ name }}
        <v-tooltip bottom>
          <template v-slot:activator="{ on, attrs }">
            <router-link
              :to="url"
            >
              <v-btn
                icon
                color="white"
                v-bind="attrs"
                v-on="on"
                class="ml-1"
              >
                <v-icon>mdi-open-in-new</v-icon>
              </v-btn>
            </router-link>
          </template>
          <span>Open logs for CRD</span>
        </v-tooltip>
        <v-spacer></v-spacer>
        <v-tooltip bottom>
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              icon
              color="white"
              v-bind="attrs"
              v-on="on"
              @click="downloadCRD"
            >
              <v-icon>mdi-download</v-icon>
            </v-btn>
            </template>
          <span>Download CRD</span>
        </v-tooltip>
        <v-btn
          icon
          color="white"
          @click="() => { this.$emit('unselectNode', {}); }"
        >
          <v-icon>mdi-window-close</v-icon>
        </v-btn>

      </v-card-title>
      <v-card-subtitle>
        <div class="conditionChipListWrapper my-2">
          <div
            v-for="(condition, index) in conditions"
            :key="index"
          >
            <v-tooltip
              :top="scrollY <= 20"
              :bottom="scrollY > 20"
              :disabled="condition.status === 'True'"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-chip
                  active
                  link
                  :class="{
                    'conditionChip': true,
                  }"
                  :outlined="$vuetify.theme.dark"
                  :color="($vuetify.theme.dark) ? getType(condition) : 'white'"
                  :text-color="($vuetify.theme.dark) ? '' : getType(condition)"
                  @click="selectCondition(index)"
                  v-bind="attrs"
                  v-on="on"
                >
                  <StatusIcon
                    :type="condition.status === 'True' ? 'success' : (condition.Status === 'Unknown' ? 'unknown' : (condition.severity ? condition.severity.toLowerCase() : 'unknown' ))"
                    :spinnerWidth="2"
                    left
                  >
                  <!-- TODO: verify that StatusIcon works when passing in undefined as type, i.e. if condition.severity is undefined -->
                  </StatusIcon>
                  {{ condition.type }}
                </v-chip>
              </template>
              <span v-if="condition.severity && condition.reason">{{ condition.severity }}: {{ condition.reason }}</span>
              <span v-else-if="condition.severity">{{ condition.severity }}</span>
              <span v-else-if="condition.reason">{{ condition.reason }}</span>
              <span v-else>Unknown</span>
            </v-tooltip>
          </div>
        </div>

        <div class="mt-4">
          <v-text-field
            v-model="search"
            label="Search Custom Resource Fields"
            dark
            flat
            :solo-inverted="!$vuetify.theme.dark"
            :solo="$vuetify.theme.dark"
            hide-details
            clearable
            clear-icon="mdi-close-circle-outline"
            :color="($vuetify.theme.dark) ? 'white' : color"
          ></v-text-field>
          <v-checkbox
            v-model="caseSensitive"
            dark
            hide-details
            label="Case sensitive search"
            :color="($vuetify.theme.dark) ? 'white' : color"
          ></v-checkbox>
        </div>

      </v-card-subtitle>
    </v-sheet>
    <v-card-text>
      <v-treeview
        hoverable
        :items="items"
        :search="search"
        :filter="filter"
        :open.sync="open"
        :open-all="true"
        :active.sync="active"
        activatable
        rounded
        class="text-wrap"
      >
        <template v-slot:label="{ item }">
          <highlightjs 
            language="yaml" 
            :ref="item.id"
            :code="item.name" 
            class="text-wrap yaml-code" 
          />
        </template>
      </v-treeview>
    </v-card-text>
  </v-card>
</template>

<script>
import yaml from "js-yaml";
import StatusIcon from "./StatusIcon.vue";
import { useSettingsStore } from "../stores/settings.js";
import colors from "vuetify/lib/util/colors";

export default {
  name: "CustomResourceDefinition",
  components: {
    StatusIcon,
  },
  props: {
    downloadType: String,
    items: Array,
    jsonItems: Object,
    name: String,
    color: String,
  },
  setup() {
    const store = useSettingsStore();

    return { store };
  },
  data() {
    return {
      open: [],
      active: [], // for auto-highlighting statuses
      search: null,
      caseSensitive: false,
      conditions: [],
      url: "",
      scrollY: 0,
    };
  },
  mounted() {
    this.setConditions(this.jsonItems?.status?.conditions);
    window.addEventListener("scroll", this.onScroll);
    console.log("JSON items are", this.jsonItems);
    let kind = this.jsonItems.kind;
    let name = this.jsonItems.metadata.name;
    let namespace = this.jsonItems.metadata.namespace;
    this.url = "/logs?kind=" + kind + "&name=" + name + "&namespace=" + namespace;
    console.log("URL is", this.url);
  },
  methods: {
    getType(condition) {
      if (condition.status === "True") return "success";
      else if (condition.isError || !condition.severity || condition.status === "Unknown") return "error"; // if severity is undefined, we assume it's an error
      else return "warning";
    },
    onScroll(e) {
      this.scrollY = window.scrollY;
      // this.windowTop = window.top.scrollY /* or: e.target.documentElement.scrollTop */
    },
  downloadCRD() {
      const link = document.createElement("a");
      let crdString = "";
      if (this.store.selectedFileType === "JSON")
        crdString = JSON.stringify(this.jsonItems, null, 2);
      else if (this.store.selectedFileType === "YAML") {
        crdString = yaml.dump(this.jsonItems);
      }
      link.href = `data:text/plain;charset=utf-8,${crdString}`;
      link.download =
        this.name + "." + this.store.selectedFileType.toLowerCase();
      link.click();
    },
    setConditions(conditions) {
      this.conditions = [];
      if (conditions !== undefined) {
        conditions.forEach((e, i) => {
          this.conditions.push({
            type: e.type,
            status: e.status,
            isError: e.severity === "Error",
            severity: e.severity,
            reason: e.reason,
          });
        });
      }
      console.log(this.conditions);
    },
    selectCondition(index) {
      this.open.push(".status");
      this.open.push(".status.conditions");
      this.open.push(".status.conditions[" + index + "]");
      this.active.push(".status.conditions[" + index + "].type");
      console.log(this.open);

      let refName = ".status.conditions[" + index + "].type";
      this.$nextTick(() =>
        this.$vuetify.goTo(this.$refs[refName], {
          easing: "easeInOutQuint",
          duration: 1000,
        })
      );
    },
  },
  watch: {
    jsonItems: {
      handler(val, old) {
        console.log("Val is", val);
        this.setConditions(val?.status?.conditions);
      },
    },
    items: {
      handler(val, old) {
        let recurse = function(items, open = []) {
          items.forEach((item) => {
            if (item.children) {
              open = open.concat(recurse(item.children));
            }
            open.push(item.id);
          });
          return open;
        };

        this.open = recurse(val);
      },
    }
  },
  computed: {
    filter() {
      return this.caseSensitive
        ? (item, search, textKey) => {
            // console.log(item, search, textKey);
            return item["name"].indexOf(search) > -1;
          }
        : (item, search, textKey) => {
            // console.log(item, search, textKey);
            return (
              item["name"].toLowerCase().indexOf(search.toLowerCase()) > -1
            );
          };
    },
  },
};
</script>

<style lang="less" scoped>
.conditionChipListWrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>

<style lang="less">
.v-treeview-node__label {
  padding: 10px 0;
}

.resource-card .yaml-code code {
  font-size: 100%;
}
</style>