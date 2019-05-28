<template>
  <div>
    <h1>Namespaces</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!namespaces.length"
      subject="namespaces"
      description="Namespaces can keep your different projects organised and secure so that you don't forget what is what."
      action="Add Namespace"
      target="/namespaces/new"
    />

    <template v-else>
      <div class="list">
        <h2>Namespaces ({{ namespaces.length }})</h2>
        <div
          class="item"
          v-for="namespace in namespaces"
          @click="$push(`/namespaces/${namespace.id}`)"
        >
          <span>{{ namespace.name }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Namespaces" },

  data() {
    return {
      loading: true,
      namespaces: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/namespaces");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.namespaces = data;
    }
  }
};
</script>
