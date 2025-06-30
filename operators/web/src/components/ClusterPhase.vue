<template>
  <v-card-subtitle class="cardSubtitle">
    <!-- TODO: add loading spinner for "-ing" phases, x for failed, and ? for unknown -->
    <!-- <span>Target Cluster</span> -->
    <div class="wrap">
      <v-icon
        v-if="icon != ''"
        class="phase-icon mr-1"
        :color="color"
      > mdi-{{ icon }} </v-icon>
      <v-progress-circular
        v-else
        class="phase-spinner mr-1"
        indeterminate
        :size="14"
        :width="2"
        :color="color"
      ></v-progress-circular>
      <span :class="color + '--text'">
        {{ phase }}
      </span>
    </div>
  </v-card-subtitle>
</template>

<script>
export default {
  name: "ClusterPhase",
  props: {
    phase: String,
  },
  data() {
    return {
      color: "",
      icon: "",
    };
  },
  methods: {
    getColor(phase) {
      switch (phase) {
        case "Provisioned":
          this.color = "success";
          break;
        case "Pending":
        case "Provisioning":
        case "Deleting":
          this.color = "warning";
          break;
        case "Failed":
        case "Unknown":
          this.color = "error";
          break;
        default:
          this.color = "grey";
          break;
      }
    },
    setIcon(phase) {
      switch (phase) {
        case "Provisioned":
          this.icon = "check-circle";
          break;
        case "Pending":
        case "Provisioning":
        case "Deleting":
          this.icon = "";
          break;
        case "Failed":
          this.icon = "alert";
          break;
        case "Unknown":
          this.icon = "help-circle";
          break;
        default:
          this.icon = "";
          break;
      }
    },
  },
  mounted() {
    this.getColor(this.phase);
    this.setIcon(this.phase);
  },
};
</script>

<style lang="less" scoped>
.cardSubtitle {
  padding: 3px 16px !important;
  line-height: 16px;

  .wrap {
    display: flex;
    flex-direction: row;
    align-items: center;

    .phase-icon {
      display: inline-block;
      font-size: 16px;
    }

    .phase-spinner {
      display: inline-block;
    }

    span {
      display: inline-block;
    }
  }
}
</style>