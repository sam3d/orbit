<template>
  <div
    class="button"
    :class="{ disabled: busy || disabled, flashing: pendingConfirmation }"
    @click="click"
  >
    <div v-if="busy" class="overlay">
      <Spinner />
    </div>

    <span :class="{ hidden: busy }">
      {{ pendingConfirmation ? `Are you sure? (${countdown})` : text }}
    </span>
    <span class="cancel" v-if="pendingConfirmation">Press to cancel</span>
  </div>
</template>

<script>
import Spinner from "@/components/Spinner";

export default {
  props: {
    text: { type: String },
    busy: { type: Boolean },
    disabled: { type: Boolean },
    confirm: { type: Boolean },
    countdown: { type: Number, default: 5 }
  },

  data() {
    return {
      pendingConfirmation: false
    };
  },

  components: {
    Spinner
  },

  methods: {
    click() {
      if (this.busy || this.disabled) return;

      if (this.confirm) {
        if (!this.pendingConfirmation) {
          this.startTimer(5);
          this.pendingConfirmation = true;
        } else {
          this.stopTimer();
          this.pendingConfirmation = false;
        }
        return;
      }

      this.$emit("click");
    },

    startTimer() {
      this.countdown = 5;
      this.interval = setInterval(this.tick, 1000);
    },

    stopTimer() {
      this.countdown = null;
      clearInterval(this.interval);
    },

    tick() {
      if (--this.countdown == 0) {
        this.stopTimer();
        this.pendingConfirmation = false;
        this.$emit("click");
      }
    }
  },

  watch: {
    disabled() {
      // If the disabled property changes, ensure we can't go ahead with this.
      this.stopTimer();
      this.pendingConfirmation = false;
    }
  }
};
</script>

<style lang="scss" scoped>
@keyframes flashing {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.8;
  }
  100% {
    opacity: 1;
  }
}
.button {
  display: flex;
  flex-direction: column;
  align-items: center;

  &.flashing {
    animation: flashing 1s ease forwards infinite;
  }

  .overlay {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;

    display: flex;
    align-items: center;
    justify-content: center;
  }

  .cancel {
    font-size: 13px;
    margin-top: 5px;
  }
}
</style>
