<template>
  <div class="sidebar-screen" v-if="!loading">
    <div class="type">Deployment</div>
    <h1>{{ deployment.name }}</h1>

    <label>Building from</label>

    <table class="build-from">
      <tr>
        <th>Repository</th>
        <td>{{ deployment.repository.name }}</td>
      </tr>
      <tr>
        <th>Branch</th>
        <td>{{ deployment.branch }}</td>
      </tr>
      <tr>
        <th>Path</th>
        <td>{{ deployment.path }}</td>
      </tr>
    </table>

    <code v-if="currentBuild" ref="currentBuild">{{ currentBuild }}</code>
    <code v-else-if="busyBuilding">
      <span class="comment"># Preparing build...</span>
    </code>
    <code v-else>
      <span class="comment"
        ># Press the "Build deployment" button to begin build.</span
      >
    </code>

    <div class="buttons">
      <Button
        text="Build deployment"
        class="purple"
        @click="build"
        :busy="busyBuilding"
        :disabled="busyDeleting"
      />

      <Button
        text="Delete this deployment"
        confirm
        class="red"
        @click="deleteDeployment"
        :busy="busyDeleting"
        :disabled="busyBuilding"
      />
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      loading: true,
      busyDeleting: false,
      busyBuilding: false,
      currentBuild: null,
      deployment: {
        id: "",
        name: "",
        repository: "",
        branch: "",
        path: "",
        build_logs: {}
      }
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      // Load the data.
      const id = this.$route.params.id;
      let res = await this.$api.get(`/deployment/${id}`);
      if (res.status !== 200) return alert(res.data);
      this.deployment = res.data;
      await this.$nextTick();

      // Retrieve the repository information for this deployment.
      res = await this.$api.get(`/repository/${this.deployment.repository}`);
      if (res.status !== 200) return alert(res.data);
      this.deployment.repository = res.data;

      this.loading = false;
    },

    async build() {
      this.busyBuilding = true;

      // Start the build process in the background.
      const request = this.$api.post(`/deployment/${this.deployment.id}/build`);
      const ticker = setInterval(this.loadCurrentBuild, 2000);
      const res = await request; // Wait for the response to resolve.
      clearInterval(ticker); // Clear the interval.
      await this.loadCurrentBuild(); // Load the final build lines
      this.busyBuilding = false; // We're no longer building.
    },

    async loadCurrentBuild() {
      const res = await this.$api.get(`/deployment/${this.deployment.id}`);
      if (res.status !== 200) return;
      const deployment = res.data;

      // Find the build log that wasn't there before.
      const current = Object.keys(deployment.build_logs).find(
        log => !Object.keys(this.deployment.build_logs).includes(log)
      );

      // If it wasn't there, don't do anything
      if (!current) return;

      // Set the current build data.
      this.currentBuild = deployment.build_logs[current]
        .reduce((str, line) => (str += `${line}\n`), "")
        .replace(/\[1G/g, "");

      // Scroll to the bottom of the build view.
      await this.$nextTick();
      const ref = this.$refs.currentBuild;
      ref.scrollTop = ref.scrollHeight;
    },

    async deleteDeployment() {
      this.busyDeleting = true;
    }
  }
};
</script>

<style lang="scss" scoped>
.sidebar-screen {
  height: calc(100vh - 140px);
}

label {
  margin-top: 10px;
}

code {
  max-width: 100%;
  width: 600px;
  white-space: pre;
  margin-top: 30px;
  flex-grow: 1;
  font-size: 14px;
}

.build-from {
  text-align: left;
  width: 100%;
  max-width: 300px;

  &,
  & td,
  & th {
    border: 1px solid #ddd;
    padding: 10px;
  }

  th {
    font-weight: bold;
  }
}

.buttons {
  margin-top: 20px;
  display: flex;

  .button:not(:last-of-type) {
    margin-right: 20px;
  }
}
</style>
