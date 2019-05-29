<template>
  <div class="sidebar-screen" v-if="!loading">
    <div class="type">Repository</div>
    <h1>{{ repo.name }}</h1>

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

export default {
  components: { RepoInstructions },

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
