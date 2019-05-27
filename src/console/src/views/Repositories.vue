<template>
  <div>
    <h1>Repositories</h1>

    <LoadingList v-if="loading" header />
    <template v-else>
      <div class="list">
        <h2>Repositories (3)</h2>
        <div class="item">Hello</div>
        <div class="item">Hello</div>
        <div class="item">Hello</div>
      </div>

      <div class="list">
        <div class="item">Hello</div>
        <div class="item">Hello</div>
        <div class="item">Hello</div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Repositories" },
  data() {
    return {
      loading: true
    };
  }
};
</script>
