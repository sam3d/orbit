<template>
  <div>
    <h1>Repositories</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!repos.length"
      subject="repositories"
      description="Repositories are your locations to store your code so that you can create deployments."
      action="Add Repository"
      target="/repositories/new"
    />

    <template v-else>
      <div class="list">
        <h2>Repositories ({{ repos.length }})</h2>
        <div
          class="item"
          v-for="repo in repos"
          @click="$push(`/repositories/${repo.id}`)"
        >
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
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.repos = data.filter(r => r.namespace_id === this.$namespace());
    }
  }
};
</script>
