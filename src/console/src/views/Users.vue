<template>
  <div>
    <h1>Users</h1>

    <LoadingList v-if="loading" header />
    <Empty
      v-else-if="!users.length"
      subject="users"
      description="Users are different individuals that have access to the Orbit system and console."
      action="Add User"
      target="/users/new"
    />

    <template v-else>
      <div class="list">
        <h2>Users ({{ users.length }})</h2>
        <div
          class="item"
          v-for="user in users"
          @click="$push(`/users/${user.id}`)"
        >
          <div class="profile" :style="user.profileStyle"></div>
          <span class="name">{{ user.name }}</span>
          <span class="username">@{{ user.username }}</span>
          <span class="email">{{ user.email }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script>
import defaultProfile from "@/assets/icon/blank-profile.svg";

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
      if (!Array.isArray(data)) return (this.loading = false);

      for (let user of data) {
        let url = `/api/user/${user.id}/profile`;
        const { status } = await this.$api.get(url);
        url = status === 200 ? url : defaultProfile;
        user.profileStyle = { backgroundImage: `url("${url}")` };
      }

      this.users = data;
      this.loading = false;
    }
  }
};
</script>

<style lang="scss" scoped>
.list {
  .item {
    padding: 10px !important;
  }
}

.profile {
  width: 38px;
  height: 38px;
  border-radius: 1000px;
  margin-right: 10px;
  background-size: cover;
  background-repeat: no-repeat;
  background-position: center;
  background-color: #ddd;
}

.name {
  font-weight: bold;
  margin-right: 20px;
}

.email {
  font-size: 14px;
  margin-left: 20px;
}
</style>
