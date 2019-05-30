<template>
  <div class="sidebar-screen">
    <h1>Add a new namespace</h1>
    <p class="description">
      Namespaces are logical groups where you can keep isolated dependencies and
      code.
    </p>

    <label>Name</label>
    <input
      type="text"
      class="defined"
      size="30"
      placeholder="Name"
      :disabled="busy"
      v-model="name"
      ref="nameField"
      @keypress.enter="create"
    />

    <Button
      text="Create namespace"
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
      name: ""
    };
  },

  mounted() {
    this.$refs.nameField.focus();
  },

  methods: {
    async create() {
      if (!this.valid || this.busy) return;
      this.busy = true;

      const body = { name: this.name };
      const res = await this.$api.post("/namespace", body);
      this.busy = false;
      if (res.status !== 201) return alert(res.data);

      this.$reload();
      this.$push("/namespaces");
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

.button {
  margin-top: 30px;
}
</style>
