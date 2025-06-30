<template>
  <v-avatar
    :size="size"
    min-width="0"
    min-height="0"
    :color="getColor(type)"
    :left="left"
    v-if="circle"
  >
    <v-icon
      v-if="type==='success'"
      class="readyIcon"
      :color="$vuetify.theme.dark ? 'black' : 'white'"
      :size="size-2"
    > mdi-check</v-icon>
    <v-icon
      v-else-if="type==='error'"
      class="readyIcon"
      :color="$vuetify.theme.dark ? 'black' : 'white'"
      :size="size-2"
    > mdi-exclamation</v-icon>
    <v-progress-circular
      v-else-if="spinner && (type==='warning' || type==='info')"
      class="readySpinner"
      indeterminate
      :size="spinnerSize"
      :width="spinnerWidth"
      :color="$vuetify.theme.dark ? 'black' : 'white'"
    >
    </v-progress-circular>
    <v-icon
      v-else-if="type==='warning'"
      class="readyIcon"
      :color="$vuetify.theme.dark ? 'black' : 'white'"
      :size="size-2"
    > mdi-exclamation</v-icon>
    <v-icon
      v-else-if="type==='info'"
      class="readyIcon"
      :color="$vuetify.theme.dark ? 'black' : 'white'"
      :size="size-2"
    > mdi-information-variant</v-icon>
    <v-icon
      v-else-if="type==='unknown' || !type"
      class="readyIcon"
      :color="$vuetify.theme.dark ? 'black' : 'white'"
      :size="size-2"
    > mdi-help</v-icon>
  </v-avatar>
  <v-avatar
    :size="size"
    tile
    min-width="0"
    min-height="0"
    :color="getColor(type)"
    :left="left"
    class="mr-1"
    v-else
  >
    <v-icon
      v-if="type==='success'"
      class="readyIcon"
      :color="getColor(type)"
      :size="size"
    > mdi-check-circle</v-icon>
    <v-icon
      v-else-if="type==='error'"
      class="readyIcon"
      :color="getColor(type)"
      :size="size"
    > mdi-alert-circle</v-icon>
    <v-progress-circular
      v-else-if="spinner && (type==='warning' || type==='info')"
      class="readySpinner"
      indeterminate
      :size="spinnerSize"
      :width="spinnerWidth"
      :color="getColor(type)"
    >
    </v-progress-circular>
    <v-icon
      v-else-if="type==='warning'"
      class="readyIcon"
      :color="getColor(type)"
      :size="size"
    > mdi-alert-circle</v-icon>
    <v-icon
      v-else-if="type==='info'"
      class="readyIcon"
      :color="getColor(type)"
      :size="size"
    > mdi-information</v-icon>
    <v-icon
      v-else-if="type==='unknown' || !type"
      class="readyIcon"
      :color="getColor(type)"
      :size="size"
    > mdi-help-circle</v-icon>
  </v-avatar>
</template>

<script>
export default {
  name: "StatusIcon",
  props: {
    type: {
      default: "success",
      type: String,
    },
    size: Number,
    spinnerWidth: {
      default: 1.5,
      type: Number,
    },
    circle: Boolean, 
    // If circle is true, the icon will be displayed in a circle with the color in the background. This is meant for the top right corner of a CRD.
    // If circle is false, the icon will be displayed in a square with the color as the border. This is meant for the condition chips in the CRD views.
    spinner: Boolean,
    left: Boolean,
  },
  data() {
    return {
      spinnerSize: this.size ? this.size * 0.75 : 12,
    };
  },
  methods: {
    getColor(type) {
      if (!this.circle) return "";

      if (type === "unknown") return "error";

      if (this.spinner && (type === "warning" || type === "info"))
        return "warning";

      // Remove this if we want info to be blue.
      if (type === "info") return "warning";

      return type;
    },
  },
};
</script>