<template>
  <LoadingSlider v-if="loading" />
  <div class="sidebar-screen" v-else>
    <div class="type">Repository</div>
    <h1>{{ repo.name }}</h1>

    <Browser :files="repo.files" class="browser" />

    <RepoInstructions :name="repo.name" />

    <Button
      text="Delete this repository"
      confirm
      class="red"
      @click="deleteRepo"
      :busy="busyDeleting"
    />
  </div>
</template>

<script>
import RepoInstructions from "@/components/RepoInstructions";
import Browser from "@/components/Browser";

export default {
  components: { RepoInstructions, Browser },

  data() {
    return {
      repo: {
        name: "",
        files: {},
        id: ""
      },
      loading: true,
      busyDeleting: false
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      // Load the data.
      const id = this.$route.params.id;
      const res = await this.$api.get(`/repository/${id}`);
      this.loading = false;
      if (res.status !== 200) return alert(res.data);
      this.repo = res.data;
    },

    async deleteRepo() {
      this.busyDeleting = true;
    }
  }
};
</script>

<style lang="scss" scoped>
.browser {
  margin-top: 10px;
}
</style>
