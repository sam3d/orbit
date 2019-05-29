<template>
  <div class="sidebar-screen">
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

    <RepoInstructions :name="name" />

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
import RepoInstructions from "@/components/RepoInstructions";

export default {
  components: { RepoInstructions },

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

      const body = { name: this.name, namespace: this.$namespace() };
      const res = await this.$api.post("/repository", body);
      this.busy = false;
      if (res.status !== 201) return alert(res.data);

      this.$reload();
      this.$push(`/repositories`);
    }
  },

  computed: {
    valid() {
      return this.name.length > 0;
    }
  },

  watch: {
    name(value) {
      this.name = this.$sanitize(value);
    }
  }
};
</script>

<style lang="scss" scoped>
label {
  margin-top: 30px;
}
</style>
