<template>
  <div class="progress" ref="container" :style="style">
    <div class="stages">
      <div class="stage" :class="stage.state" v-for="stage in stages">
        <div class="icon">
          <div v-if="stage.state === 'complete'" class="tick">✔</div>
          <div v-if="stage.state === 'active'" class="dot"></div>
          <div v-if="stage.state === 'incomplete'" class="dot purple"></div>
        </div>
        <div class="text">{{ stage.name }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: ["stages", "hidden"],

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
        margin: this.hidden ? "0 20px" : "20px"
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
    color: #8959ea;
    flex-shrink: 0;

    padding: 14px;
    border-radius: 7px;

    &:hover {
      background-color: transparentize(#8959ea, 0.95);
    }

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

        &.purple {
          background-color: #8959ea;
        }
      }
    }

    &.complete,
    &.active {
      .icon {
        background-color: #8959ea;
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
