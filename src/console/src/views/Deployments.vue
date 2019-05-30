<template>
  <div>
    <h1>Deployments</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!deployments.length"
      subject="deployments"
      description="Deployments are where you build and deploy the code from your repositories."
      action="Add Deployment"
      target="/deployments/new"
    />

    <template v-else>
      <div class="list">
        <h2>Deployments ({{ deployments.length }})</h2>
        <div
          class="item"
          v-for="deployment in deployments"
          @click="$push(`/deployments/${deployment.id}`)"
        >
          <div class="status" :class="{ green: deployment.build_logs }"></div>
          <span class="name">{{ deployment.name }}</span>
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

<style lang="scss" scoped>
.status {
  width: 8px;
  height: 8px;
  border-radius: 8px;
  margin-right: 8px;
  background-color: #feca57;

  &.green {
    background-color: #1dd1a1;
  }
}
</style>
