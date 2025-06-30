<template>
  <div
    class="badge"
    :style="{
      'border-color': $vuetify.theme.themes[theme].background,
    }"
  >
    <div
      class="topRight"
      :style="{
        'top': -(size/2) + 'px',
        'right': -(size/2) + 'px',
        'height': size + 'px',
        'width': size + 'px',
      }"
    >
      <div
        class="border"
        :style="{
          'border-color': $vuetify.theme.themes[theme].background,
          'height': (size-4) + 'px',
          'width': (size-4) + 'px',
        }"
      >

        <StatusIcon
          circle
          class="readyWrap"
          spinner
          :type="type"
          :size="size-4"
        >
        </StatusIcon>
      </div>
    </div>

  </div>

</template>

<script>
import colors from "vuetify/lib/util/colors";
import StatusIcon from "./StatusIcon.vue";

export default {
  name: "StatusBadge",
  components: {
    StatusIcon,
  },
  props: {
    type: String,
    blinking: Boolean,
    size: {
      default: 16,
      type: Number,
    },
  },
  computed: {
    theme() {
      return this.$vuetify.theme.dark ? "dark" : "light";
    },
  },
  methods: {
    getColor() {
      switch (this.type) {
        case "ready":
          return colors.green.base;
        case "error":
          return colors.red.accent2;
        case "loading":
          return colors.orange.darken1;
        default:
          return colors.grey;
      }
    },
  },
};
</script>

<style lang="less" scoped>
.blink-transition {
  transition: all 1s ease;
  opacity: 0%;
}

.blink-enter,
.blink-leave {
  opacity: 100%;
}

.badge .topRight {
  position: absolute;
  display: inline-block;
}

.border {
  position: relative;
  border-radius: 50%;
  border-width: 2px !important;
  border-style: solid !important;
}

.badge .readyWrap {
  position: absolute;
  // border: 2px solid !important;

  // box-shadow: 0px 0px 10px rgba(0, 0, 0, 1);
}

.readyIcon {
  display: inline-block;
  vertical-align: middle;
  text-align: center;
  // padding: 2px;
}

@-webkit-keyframes Blinking {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.2;
  }
  100% {
    opacity: 1;
  }
}
@-moz-keyframes Blinking {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.2;
  }
  100% {
    opacity: 1;
  }
}
@keyframes Blinking {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.2;
  }
  100% {
    opacity: 1;
  }
}

.blinking {
  -webkit-animation: Blinking 3s ease-in-out infinite !important;
  -moz-animation: Blinking 3s ease-in-out infinite !important;
  animation: Blinking 3s ease-in-out infinite !important;
}

// .readySpinner {
//   display: inline-block !important;
//   margin-bottom: 1px;
// }
</style>