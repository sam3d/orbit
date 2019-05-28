<template>
  <div>
    <h1>Nodes</h1>

    <LoadingList v-if="loading" header />

    <template v-else>
      <div class="list">
        <h2>Managers ({{ managerNodes.length }})</h2>
        <NodeListItem v-for="node in managerNodes" :node="node" />

        <h2>Workers ({{ workerNodes.length }})</h2>
        <NodeListItem v-for="node in workerNodes" :node="node" />
      </div>
    </template>
  </div>
</template>

<script>
import NodeListItem from "@/components/NodeListItem";

export default {
  meta: { title: "Nodes" },
  components: { NodeListItem },

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
  },

  computed: {
    managerNodes() {
      return this.nodes.filter(n => n.node_roles.includes("MANAGER"));
    },

    workerNodes() {
      return this.nodes.filter(n => n.node_roles.includes("WORKER"));
    }
  }
};
</script>
