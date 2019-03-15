<template>
  <div class="button" :class="{ disabled: busy }" @click="$emit('click')">
    <div v-if="busy" class="overlay">
      <Spinner />
    </div>

    <span :class="{ hidden: busy }">{{ text }}</span>
  </div>
</template>

<script>
import Spinner from "@/components/Spinner";

export default {
  props: ["text", "busy"],

  components: {
    Spinner
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
