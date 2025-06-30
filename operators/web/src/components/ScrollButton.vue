<template>
  <div id="scrollButtonWrap">
    <v-fab-transition>
      <v-btn
        v-show="scrollY > 20"
        ref="button"
        color="accent"
        medium
        fab
        @click="$vuetify.goTo(0, {
          duration: 500,
          offset: 0,
          easing: 'easeInOutCubic',
        })"
      >
        <v-icon :color="$vuetify.theme.dark ? 'black' : 'white'">
          mdi-chevron-up
        </v-icon>
      </v-btn>
    </v-fab-transition>
  </div>
</template>

<script>
export default {
  name: "ScrollButton",
  data() {
    return {
      scrollY: 0,
    };
  },
  mounted() {
    window.addEventListener("scroll", this.onScroll);
  },
  methods: {
    onScroll(e) {
      this.scrollY = window.scrollY;
    }
  },
};
</script>

<style lang="less" scoped>
#scrollButtonWrap {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
}
</style>