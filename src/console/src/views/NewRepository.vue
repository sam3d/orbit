<template>
  <div class="create-new">
    <h1>Create a new repository</h1>
    <p class="description">
      A repository is a highly available location in your cluster for you to
      store your git code.
    </p>
    <code>
      <p class="comment"># Create and deploy your code to this repository.</p>
      git init<br />
      echo "# {{ name || "{name}" }}" > README.md<br />
      git commit -m "Initial commit"<br />
      git remote add deploy {{ path }}<br />
      git push -u deploy master
    </code>
    <label>Repository name</label>
    <input
      type="text"
      class="defined"
      placeholder="name"
      v-model="name"
      @keypress.enter="newRepo"
      :disabled="busy"
    />
    <Button
      text="Create repository"
      class="purple"
      :busy="busy"
      @click="newRepo"
      :disabled="!valid"
    />
  </div>
</template>

<script>
export default {
  data() {
    return {
      busy: false,
      name: ""
    };
  },

  methods: {
    async newRepo() {
      if (this.busy || !this.valid) return;
      this.busy = true;

      const body = { name: this.safeName, namespace: this.$namespace() };
      const res = await this.$api.post("/repository", body);
      if (res.status !== 201) return alert(res.data);
      this.busy = false;
      this.$push(`/repositories/${res.data.id}`);
    }
  },

  computed: {
    path() {
      const { protocol, host } = window.location;
      const main = `${protocol}//${host}/api/repo`;
      const name = this.safeName || "{name}";
      const namespace = this.$namespace();

      if (namespace) return `${main}/${namespace}/${name}`;
      else return `${main}/${name}`;
    },

    valid() {
      return this.safeName.length > 0;
    },

    safeName() {
      return this.name
        .toLowerCase()
        .trim()
        .split(" ")
        .join("-")
        .replace(/[^ a-zA-Z\-]/g, "");
    }
  }
};
</script>

<style lang="scss" scoped>
code {
  margin-top: 30px;
  margin-bottom: 30px;
  font-size: 14px;
  max-width: 500px;
  flex-shrink: 2;
  overflow: scroll;
}

.button {
  margin-top: 30px;
}
</style>
