<template>
  <LoadingSlider v-if="loading" />
  <div class="sidebar-screen" v-else>
    <div class="type">Node</div>
    <h1>{{ node.address }}</h1>
    <NodeRoles :roles="node.node_roles" />

    <div class="kvs">
      <div class="kv">
        <div class="key">RPC Port</div>
        <div class="value">:{{ node.rpc_port }}</div>
      </div>

      <div class="kv">
        <div class="key">Raft Port</div>
        <div class="value">:{{ node.raft_port }}</div>
      </div>

      <div class="kv">
        <div class="key">Serf Port</div>
        <div class="value">:{{ node.serf_port }}</div>
      </div>

      <div class="kv">
        <div class="key">WAN Serf Port</div>
        <div class="value">:{{ node.wan_serf_port }}</div>
      </div>

      <div class="kv">
        <div class="key">Swap size</div>
        <div class="value">{{ node.swap_size }} MB</div>
      </div>

      <div class="kv">
        <div class="key">Swappiness</div>
        <div class="value">{{ node.swappiness }}</div>
      </div>
    </div>

    <Button
      text="Remove this node"
      confirm
      class="red"
      @click="deleteRepo"
      :busy="busyDeleting"
    />
  </div>
</template>

<script>
import NodeRoles from "@/components/NodeRoles";

export default {
  components: { NodeRoles },
  data() {
    return {
      node: {
        id: null,
        address: null,
        rpc_port: null,
        raft_port: null,
        serf_port: null,
        wan_serf_port: null,
        node_roles: [],
        swap_size: 2048,
        swappiness: 60
      },
      loading: true,
      busyDeleting: false
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      // Load the data.
      const id = this.$route.params.id;
      const res = await this.$api.get(`/node/${id}`);
      this.loading = false;
      if (res.status !== 200) return alert(res.data);
      this.node = res.data;
    },

    async deleteRepo() {
      this.busyDeleting = true;
    }
  }
};
</script>

<style lang="scss" scoped>
.browser {
  margin-top: 10px;
}

.roles {
  margin-bottom: 30px;
}

.kvs {
  width: 100%;
}

.kv {
  .key {
    font-weight: bold;
  }

  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: 10px 0;
  &:not(:last-of-type) {
    border-bottom: solid 1px #ddd;
  }
}

.button {
  margin-top: 30px;
}
</style>
