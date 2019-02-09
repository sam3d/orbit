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
    }
  }
};
</script>
