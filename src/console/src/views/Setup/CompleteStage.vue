<template>
  <div>
    <h2>All done!</h2>
    <template v-if="mode === 'bootstrap'">
      <p class="subheader">
        Congratulations, the set up for this cluster is now complete! You may
        now start using your cluster by continuing to the log in page by
        pressing the button below.
      </p>

      <Button
        class="purple"
        text="Log in to your cluster"
        @click="$router.push('/login')"
      />
    </template>
    <template v-else>
      <p class="subheader">
        Congratulations, the set up for this node is now complete! You may now
        continue using your cluster by pressing the link below to be redirected
        to the dashboard.
      </p>

      <Button
        class="purple"
        text="Return to cluster"
        :busy="busy"
        @click="navigateToCluster"
      />
    </template>
  </div>
</template>

<script>
export default {
  props: ["mode"],

  data() {
    return {
      busy: false
    };
  },

  methods: {
    // This method will retrieve the orbit domain name from the API and navigate
    // to the correct URL for this node that has just been created.
    async navigateToCluster() {
      this.busy = true;

      try {
        const url = await this.getConsoleURL();
        const id = await this.getNodeID();
        const target = `${url}/node/${id}`;

        window.location.href = target;
      } catch (err) {
        this.busy = false;
        console.log(err);
      }
    },

    // Retrieve the orbit console URL from the API.
    async getConsoleURL() {
      // Retrieve the list of routers.
      const res = await this.$api.get("/routers", { redirect: false });
      if (res.status !== 200) throw "Could not retrieve routers.";

      // Retrieve the correct router.
      const router = res.data.find(router => router.app_id === "console");
      if (!router) throw "No router for the orbit console exists.";

      // Derive the protocol and domain into a URL.
      const protocol = router.certificate_id ? "https" : "http";
      const url = `${protocol}://${router.domain}`;

      return url;
    },

    // Get the ID of the current node.
    async getNodeID() {
      // Get the current node details.
      const res = await this.$api.get("/node/current", { redirect: false });
      if (res.status !== 200)
        throw "Could not retrieve current node information.";
      const node = res.data;

      // Return the ID.
      return node.id;
    }
  }
};
</script>

<style lang="scss" scoped>
.button {
  margin-top: 30px;
}

.button {
  display: inline-block !important;
}
</style>
