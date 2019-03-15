<template>
  <div class="button" :class="{ disabled: busy || disabled }" @click="click">
    <div v-if="busy" class="overlay">
      <Spinner />
    </div>

    <span :class="{ hidden: busy }">{{ text }}</span>
  </div>
</template>

<script>
import Spinner from "@/components/Spinner";

export default {
  props: ["text", "busy", "disabled"],

  components: {
    Spinner
  },

  methods: {
    click() {
      if (this.busy || this.disabled) return;
      this.$emit("click");
    }
  }
};
</script>

<style lang="scss" scoped>
.button {
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
}
</style>
