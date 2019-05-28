<template>
  <div>
    <h1>Volumes</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!volumes.length"
      subject="volumes"
      description="Highly available and distributed block storage volumes for your deployments"
      action="Add Volume"
    />

    <template v-else>
      <div class="list">
        <h2>Volumes ({{ volumes.length }})</h2>
        <div class="item" v-for="volume in volumes">
          <span>{{ volume.name }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Volumes" },

  data() {
    return {
      loading: true,
      volumes: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/volumes");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.volumes = data.filter(d => d.namespace_id === this.$namespace());
    }
  }
};
</script>
