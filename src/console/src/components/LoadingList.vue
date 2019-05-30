<template>
  <div class="list">
    <h2 class="placeholder" v-if="header">LOADING</h2>
    <div class="item placeholder" v-for="item in items">&nbsp;</div>
  </div>
</template>

<script>
export default {
  props: {
    header: { type: Boolean },
    items: { type: Number, default: 3 }
  }
};
</script>
