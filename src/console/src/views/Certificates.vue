<template>
  <div>
    <h1>Certificates</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!certificates.length"
      subject="certificates"
      description="These are SSL certificates that attach to routers to secure traffic coming into your application"
      action="Add Certificate"
      target="/certificates/new"
    />

    <template v-else>
      <div class="list">
        <h2>Certificates ({{ certificates.length }})</h2>
        <div
          class="item"
          v-for="certificate in certificates"
          @click="$push(`/certificates/${certificate.id}`)"
        >
          <span>{{ certificate.domains }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Certificates" },

  data() {
    return {
      loading: true,
      certificates: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/certificates");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.certificates = data.filter(
        d => d.namespace_id === this.$namespace()
      );
    }
  }
};
</script>
