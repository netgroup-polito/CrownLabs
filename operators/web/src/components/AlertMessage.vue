<template>
  <div class="alert-wrapper">
    <v-alert
      :value="value"
      @input="update"
      class="alert-message"
      :type="type"
      dismissible
      border="left"
      elevation="2"
      colored-border
    >
      {{ message }}
    </v-alert>

  </div>
</template>

<script>
export default {
  name: "AlertMessage",
  props: {
    value: Boolean, // v-model
    message: String,
    type: String
  },
  methods: {
    update(val) {
      // console.log("Emitting 'input' as ", val);
      this.$emit("input", val);
    },
  },
};
</script>

<style lang="less" scoped>
.alert-message {
  min-width: 500px;
  text-align: left;
  display: inline-block;
}

.alert-wrapper {
  text-align: center;
  width: 100%;
  position: fixed;
  bottom: 30px;
  z-index: 9999;
}
</style>