<template>
  <div class="progress" ref="container" :style="style">
    <div class="stages">
      <div
        v-for="stage in stages"
        class="stage"
        :class="stage.state"
        :style="{ minWidth: originalHeight + 'px' }"
        @click="$emit('input', stage.name)"
      >
        <div class="icon">
          <div v-if="stage.state === 'complete'" class="tick">✔</div>
          <div v-if="stage.state === 'active'" class="dot"></div>
          <div v-if="stage.state === 'incomplete'" class="dot incomplete"></div>
        </div>
        <div class="text">{{ stage.name }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: ["stages", "hidden", "value"],

  data() {
    return {
      originalHeight: 0
    };
  },

  mounted() {
    const { container } = this.$refs;
    this.originalHeight = container.scrollHeight;
  },

  computed: {
    style() {
      return {
        height: (this.hidden ? 0 : this.originalHeight) + "px",
        margin: this.hidden ? `0 20px` : `0 20px 20px 20px`
      };
    }
  }
};
</script>

<style lang="scss" scoped>
.progress {
  margin: 20px;

  background-color: #fff;
  border-radius: 10px;
  box-shadow: 0 3px 10px 0 rgba(0, 0, 0, 0.025);

  overflow-x: auto;
  overflow-y: hidden;
  flex-shrink: 0;

  transition: height 0.5s, margin-top 0.5s, margin-bottom 0.5s;

  .stages {
    cursor: default;

    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-shrink: 0;

    padding: 6px;
  }

  .stage {
    display: flex;
    flex-direction: column;
    align-items: center;
    flex-shrink: 0;

    padding: 14px;
    border-radius: 7px;

    .icon {
      width: 32px;
      height: 32px;
      border-radius: 26px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      font-size: 14px;

      .dot {
        background-color: #fff;
        width: 10px;
        height: 10px;
        border-radius: 10px;
      }
    }

    &.incomplete {
      color: #888;
      .dot {
        background-color: #888;
      }
      &:hover {
        background-color: transparentize(#888, 0.95);
      }
    }

    &.complete {
      color: #1dd1a1;
      .icon {
        background-color: #1dd1a1;
      }
      &:hover {
        background-color: transparentize(#1dd1a1, 0.95);
      }
    }

    &.active {
      color: #8959ea;
      .icon {
        background-color: #8959ea;
      }
      &:hover {
        background-color: transparentize(#8959ea, 0.95);
      }
    }

    .text {
      margin-top: 10px;
      font-size: 13px;
      font-family: "Cabin", sans-serif;
      text-transform: uppercase;
      letter-spacing: 0.1rem;
    }
  }
}
</style>
