<template>
  <LoadingSlider v-if="loading" />
  <div class="sidebar-screen" v-else>
    <div class="type">Deployment</div>
    <h1>{{ deployment.name }}</h1>

    <label>Building from</label>

    <div class="flags">
      <div class="flag">
        <div class="key">Repository</div>
        <div
          class="value link"
          @click="$push(`/repositories/${deployment.repository.id}`)"
        >
          {{ deployment.repository.name }}
        </div>
      </div>

      <div class="flag">
        <div class="key">Branch</div>
        <div class="value">{{ deployment.branch }}</div>
      </div>

      <div class="flag">
        <div class="key">Path</div>
        <div class="value">{{ deployment.path }}</div>
      </div>
    </div>

    <label style="margin-top: 20px;">Build Log</label>
    <select v-model="selectedBuild">
      <option value="" v-if="!busyBuilding">Start a new build</option>
      <option value="" v-else>Current build</option>

      <option v-for="log in logs" :value="log.id">
        {{ log.render }}
      </option>
    </select>

    <code v-if="selectedBuild" ref="selectedBuild">{{
      parseBuildLog(deployment.build_logs[selectedBuild])
    }}</code>
    <code v-else-if="currentBuild" ref="currentBuild">{{ currentBuild }}</code>
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
      selectedBuild: "",
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
      if (!this.deployment.build_logs) this.deployment.build_logs = {};
      await this.$nextTick();

      // Retrieve the repository information for this deployment.
      res = await this.$api.get(`/repository/${this.deployment.repository}`);
      if (res.status !== 200) return alert(res.data);
      this.deployment.repository = res.data;

      this.loading = false;
    },

    async build() {
      this.selectedBuild = ""; // Show the build log for the current build
      this.busyBuilding = true;

      // Start the build process in the background
      const request = this.$api.post(`/deployment/${this.deployment.id}/build`);
      const ticker = setInterval(this.loadCurrentBuild, 2000);
      const res = await request; // Wait for the response to resolve
      clearInterval(ticker); // Clear the interval

      await this.loadCurrentBuild(); // Load the final build lines
      await this.load(); // Load in the final new build data
      this.currentBuild = ""; // Clear the current build data
      this.selectedBuild = this.logs[0].id; // Show the new build data

      // Scroll to the bottom of the new build view.
      await this.$nextTick();
      const ref = this.$refs.selectedBuild;
      ref.scrollTop = ref.scrollHeight;
      this.busyBuilding = false; // We're no longer building
      this.$reload(); // Update the list.
    },

    async loadCurrentBuild() {
      const res = await this.$api.get(`/deployment/${this.deployment.id}`);
      if (res.status !== 200) return;
      const deployment = res.data;
      if (!deployment.build_logs) deployment.build_logs = {};

      // Find the build log that wasn't there before.
      const current = Object.keys(deployment.build_logs).find(
        log => !Object.keys(this.deployment.build_logs).includes(log)
      );

      // If it wasn't there, don't do anything
      if (!current) return;

      // Set the current build data.
      this.currentBuild = this.parseBuildLog(deployment.build_logs[current]);

      // Scroll to the bottom of the build view.
      await this.$nextTick();
      const ref = this.$refs.currentBuild;
      ref.scrollTop = ref.scrollHeight;
    },

    parseBuildLog(log) {
      return log
        .reduce((str, line) => (str += `${line}\n`), "")
        .replace(/\[1G/g, "");
    },

    async deleteDeployment() {
      this.busyDeleting = true;
    }
  },

  computed: {
    logs() {
      return Object.keys(this.deployment.build_logs)
        .map(key => {
          const tokens = key.split("/");

          // Derive the properties from the tokens.
          const o = {
            id: key,
            hash: "",
            time: null,
            path: "",
            render: ""
          };

          if (tokens.length >= 1) o.hash = tokens[0];
          if (tokens.length >= 2) o.time = new Date(tokens[1] / 1000000);
          if (tokens.length >= 3) o.path = tokens[2];

          // Create the display string.
          o.render = "#" + o.hash.padStart(7, "0").substring(0, 7);
          o.render += ` (${o.path || " / "})`;
          if (o.time) o.render += " @ " + o.time.toLocaleString();

          // Return the output;
          return o;
        })
        .sort((a, b) => new Date(b.time) - new Date(a.time));
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

select {
  max-width: 320px;
}

.flags {
  display: flex;

  .flag {
    display: flex;
    border: solid 1px #ddd;
    border-radius: 4px;
    overflow: hidden;
    text-align: left;

    &:not(:last-of-type) {
      margin-right: 10px;
    }

    .key,
    .value {
      padding: 10px;
    }

    .key {
      background-color: #eee;
      border-right: solid 1px #ddd;
    }

    .value {
      min-width: 50px;

      &.link {
        color: #54a0ff;
        cursor: pointer;

        transition: background-color 0.2s;
        &:hover {
          background-color: transparentize(#54a0ff, 0.9);
        }
        &:active {
          background-color: transparentize(#54a0ff, 0.8);
        }
      }
    }
  }
}

.buttons {
  margin-top: 20px;
  display: flex;
  align-items: center;
  height: 60px;
  flex-shrink: 0;

  .button:not(:last-of-type) {
    margin-right: 20px;
  }
}
</style>
