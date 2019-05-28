<template>
  <div>
    <h1>Deployments</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!deployments.length"
      subject="deployments"
      description="Deployments are where you build and deploy the code from your repositories."
      action="Add Deployment"
    />

    <template v-else>
      <div class="list">
        <h2>Deployments ({{ deployments.length }})</h2>
        <div class="item" v-for="deployment in deployments">
          <span>{{ deployment.name }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Deployments" },

  data() {
    return {
      loading: true,
      deployments: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/deployments");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.deployments = data.filter(d => d.namespace_id === this.$namespace());
    }
  }
};
</script>
