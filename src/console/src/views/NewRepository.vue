<template>
  <div class="create-new">
    <h1>Create a new repository</h1>
    <p class="description">
      A repository is a highly available location in your cluster for you to
      store your git code.
    </p>

    <label>Repository name</label>
    <input
      type="text"
      class="defined"
      size="30"
      ref="nameField"
      placeholder="Name"
      v-model="name"
      @keypress.enter="newRepo"
      :disabled="busy"
    />

    <label>Deploying code</label>
    <code class="light">{{ path }}</code>

    <code>
      <p class="comment"># Run this script in your code directory.</p>
      git init<br />
      echo "# {{ name || "{name}" }}" > README.md<br />
      git commit -m "Initial commit"<br />
      git remote add deploy {{ path }}<br />
      git push -u deploy master
    </code>

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

  mounted() {
    this.$refs.nameField.focus();
  },

  methods: {
    async newRepo() {
      if (this.busy || !this.valid) return;
      this.busy = true;

      const body = { name: this.safeName, namespace: this.$namespace() };
      const res = await this.$api.post("/repository", body);
      this.busy = false;
      if (res.status !== 201) return alert(res.data);

      this.$reload();
      this.$push(`/repositories`);
    }
  },

  computed: {
    path() {
      const { protocol, host } = window.location;
      const main = `${protocol}//${host}/api/repo`;
      const name = this.paddedName;
      const namespace = this.$store.state.namespaceName;

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
    },

    paddedName() {
      const name = this.safeName || "{name}";
      return name.padEnd(13, "\u00A0");
    }
  }
};
</script>

<style lang="scss" scoped>
code {
  width: 100%;
  max-width: 480px;
}

code:not(.light) {
  font-size: 14px;
  margin-top: 10px;
  max-width: 530px;
}

label {
  margin-top: 30px;
}

.button {
  margin-top: 30px;
}
</style>
