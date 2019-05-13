<template>
  <router-view></router-view>
</template>

<script>
export default {
  meta: {
    titleTemplate(titleChunk) {
      const name = "Orbit";
      if (titleChunk) return `${titleChunk} | ${name}`;
      else return name;
    },

    /**
     * When the page changes, keep track of that new page in the store. This is
     * so that the current page can be shown on the main view.
     */
    changed({ titleChunk }) {
      this.$store.commit("title", titleChunk);
    }
  }
};
</script>
