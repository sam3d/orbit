<template>
  <LoadingSlider v-if="loading" />
  <div class="sidebar-screen" v-else>
    <h1>Add a new deployment</h1>
    <p class="description">
      A deployment will take the code from an existing repository, build it, and
      then spread it over the cluster based on your settings.
    </p>

    <label v-if="mustCreateRepo">
      You don't yet have any repositories to deploy from.
    </label>
    <template v-else>
      <label>Source repository</label>
      <p class="explain">The repository where your code deploys from.</p>
      <select v-model="selectedRepo">
        <option v-for="repo in repos" :value="repo.id">{{ repo.name }}</option>
      </select>

      <label v-if="!branches.length">
        This repository has not yet been pushed to.
      </label>
      <template v-if="branches.length">
        <label>Git branch</label>
        <p class="explain">The git branch to deploy from.</p>
        <select class="branch" v-model="branch">
          <option v-for="branch in branches">{{ branch }}</option>
        </select>

        <label>Deployment path</label>
        <p class="explain">
          The file path from the source code to deploy from.
        </p>
        <input type="text" class="defined" v-model="path" />

        <label>Deployment name</label>
        <input
          type="text"
          class="defined"
          placeholder="Name"
          v-model="name"
          @keypress.enter="create"
          size="30"
        />
      </template>
    </template>

    <Button
      text="Create deployment"
      class="purple"
      :busy="busy"
      @click="create"
      :disabled="!valid"
    />
  </div>
</template>

<script>
export default {
  data() {
    return {
      busy: false,
      name: "",
      loading: true,
      repos: [],
      selectedRepo: "",
      path: "/",
      branch: "",
      mustCreateRepo: false
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const res = await this.$api.get("/repositories");
      if (res.status !== 200) return alert(res.data);

      if (!Array.isArray(res.data)) {
        this.mustCreateRepo = true;
        this.loading = false;
        return;
      }

      // Filter by namespace.
      const repos = res.data.filter(r => r.namespace_id === this.$namespace());
      if (!repos.length) {
        this.mustCreateRepo = true;
        this.loading = false;
        return;
      }

      // Otherwise, set it to the first repository.
      this.selectedRepo = repos[0].id;

      this.repos = repos;
      this.loading = false;
    },

    async create() {
      if (!this.valid || this.busy) return;
      this.busy = true;

      const body = {
        name: this.name,
        repository_id: this.selectedRepo,
        branch: this.branch,
        path: this.path,
        namespace: this.$namespace()
      };

      const res = await this.$api.post("/deployment", body);
      this.busy = false;
      if (res.status !== 201) return alert(res.data);

      this.$reload();
      this.$push("/deployments");
    },

    useDefaultBranch() {
      if (!this.branches.length) return;
      if (this.branches.includes("master")) this.branch = "master";
      else this.branch = this.branches[0];
    }
  },

  computed: {
    valid() {
      return this.selectedRepo && this.branch && this.path && this.name;
    },

    branches() {
      const repo = this.repos.find(r => r.id === this.selectedRepo);
      if (!repo) return [];
      return Object.keys(repo.files);
    }
  },

  watch: {
    name(value) {
      this.name = this.$sanitize(value);
    },

    // Ensure we don't keep data if we change the repo.
    selectedRepo() {
      this.branch = null;
      this.path = "/";
    },

    branches() {
      this.useDefaultBranch();
    }
  }
};
</script>

<style lang="scss" scoped>
label {
  margin-top: 30px;
}

select {
  max-width: 300px;

  &.branch {
    max-width: 250px;
  }
}

.button {
  margin-top: 30px;
}
</style>
