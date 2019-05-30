<template>
  <div>
    <h1>Routers</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!routers.length"
      subject="routers"
      description="Routers combine with certificates and route external web traffic to your deployments"
      action="Add Router"
      target="/routers/new"
    />

    <template v-else>
      <div class="list">
        <h2>Routers ({{ routers.length }})</h2>
        <div
          class="item"
          v-for="router in routers"
          @click="$push(`/routers/${router.id}`)"
        >
          <span>{{ router.domain }} &rarr; {{ router.app_id }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Routers" },

  data() {
    return {
      loading: true,
      routers: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/routers");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.routers = data.filter(d => d.namespace_id === this.$namespace());
    }
  }
};
</script>
