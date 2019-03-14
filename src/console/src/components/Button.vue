<template>
  <div class="button" :class="{ disabled: busy }" @click="$emit('click')">
    <div v-if="busy" class="overlay">
      <div class="spinner"></div>
    </div>

    <span :class="{ hidden: busy }">{{ text }}</span>
  </div>
</template>

<script>
export default {
  props: ["text", "busy"]
};
</script>

<style lang="scss" scoped>
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

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

    .spinner {
      display: inline-block;
      width: 20px;
      height: 20px;
      border: 3px solid transparent;
      border-radius: 50%;
      border-top-color: #02101c;
      border-bottom-color: #02101c;
      animation: spin 1s linear infinite;
    }
  }
}
</style>
