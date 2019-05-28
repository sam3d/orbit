<template>
  <div>
    <h1>Repositories</h1>

    <LoadingList v-if="loading" header />

    <template v-else>
      <div class="list">
        <h2>Repositories ({{ repos.length }})</h2>
        <div class="item" v-for="repo in repos">
          <span>{{ repo.name }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Repositories" },

  data() {
    return {
      loading: true,
      repos: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/repositories");
      this.repos = data.filter(r => r.namespace_id === this.$namespace());
      this.loading = false;
    }
  }
};
</script>
