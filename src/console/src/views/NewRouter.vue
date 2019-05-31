<template>
  <LoadingSlider v-if="loading" />
  <div class="sidebar-screen" v-else>
    <h1>Add a new router</h1>
  </div>
</template>

<script>
export default {
  data() {
    return {
      loading: true
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {}
  }
};
</script>

<style lang="scss" scoped></style>
