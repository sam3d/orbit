<template>
  <div>
    <h1>Nodes</h1>

    <LoadingList v-if="loading" header />

    <template v-else>
      <div class="list">
        <h2>Nodes ({{ nodes.length }})</h2>
        <div class="item" v-for="node in nodes">
          <span>{{ node.address }}</span>
          <span>{{ node.node_roles }}</span>
          <span>{{ node.state }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Nodes" },

  data() {
    return {
      loading: true,
      nodes: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/nodes");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.nodes = data;
    }
  }
};
</script>
