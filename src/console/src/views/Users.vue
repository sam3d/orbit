<template>
  <div>
    <h1>Users</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!users.length"
      subject="users"
      description="Users are different individuals that have access to the Orbit system and console."
      action="Add User"
    />

    <template v-else>
      <div class="list">
        <h2>Users ({{ users.length }})</h2>
        <div class="item" v-for="user in users">
          <span>{{ user.name }}</span>
          <span>{{ user.email }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
export default {
  meta: { title: "Users" },

  data() {
    return {
      loading: true,
      users: []
    };
  },

  mounted() {
    this.load();
  },

  methods: {
    async load() {
      const { data } = await this.$api.get("/users");
      this.loading = false;
      if (!Array.isArray(data)) return;
      this.users = data;
    }
  }
};
</script>
