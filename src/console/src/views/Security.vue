<template>
  <div>
    <h1>Security</h1>
    <div class="header">Add a new node to the cluster</div>
    <p>
      In order to add a new node to the cluster, please go through the following
      steps:
    </p>

    <ol>
      <li>Install Orbit a new node</li>
      <li>
        Follow the instructions from the terminal to navigate to the web
        interface
      </li>
      <li>Start the set up and click on "Join an existing cluster"</li>
      <li>Enter the information shown below in the matching fields</li>
    </ol>

    <div class="kv">
      <div class="key">IP Address</div>
      <div class="value">{{ loading ? "Loading IP address..." : address }}</div>
    </div>
    <div class="kv">
      <div class="key">Join Token</div>
      <div class="value">{{ loading ? "Loading join token..." : token }}</div>
    </div>

    <div class="header">Refresh the join token</div>
    <p>
      If you believe that the join token has been compromised or you would
      simply like peace of mind, you can reset it. This will invalidate the
      current join token.
    </p>
    <Button
      text="Refresh join token"
      class="blue"
      :busy="busy"
      @click="refresh"
    />
  </div>
</template>

<script>
export default {
  meta: { title: "Security" },

  data() {
    return {
      busy: false,
      loading: true,
      address: "", // The cluster IP address
      token: "" // The join token
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      // Load the tokens.
      let res = await this.$api.get("/tokens");
      this.token = res.data.manager;

      // Load the cluster IP address.
      res = await this.$api.get("/nodes");
      this.address = res.data.find(n =>
        n.node_roles.includes("MANAGER")
      ).address;

      this.loading = false;
    },

    async refresh() {
      this.busy = true;
      let res = await this.$api.post("/tokens/refresh");
      this.token = res.data.manager;
      this.busy = false;
    }
  }
};
</script>

<style lang="scss" scoped>
.header {
  font-size: 18px;
  font-weight: bold;
  margin-top: 30px;
}

p {
  font-size: 17px;
  margin: 10px 0;
  line-height: 1.6rem;
}

ol {
  list-style-type: decimal;
  margin-top: 15px;
  line-height: 1.6rem;
  font-size: 16px;
  margin-bottom: 30px;
  li {
    margin-left: 30px;
  }
}

.kv {
  display: flex;
  border: solid 1px #ccc;
  margin: 15px 0;
  border-radius: 4px;
  overflow: hidden;

  .key {
    padding: 10px;
    border-right: solid 1px #ccc;
    background-color: #eee;
    flex-shrink: 0;
  }
  .value {
    padding: 10px;
    overflow-y: scroll;
    white-space: nowrap;
    background-color: #fff;
    flex-grow: 1;
  }
}

.button {
  margin-top: 10px;
}
</style>
