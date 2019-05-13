<template>
  <div class="root">
    <div class="navbar">
      <div
        class="toggle"
        :style="sidebarToggleStyle"
        @click="showSidebar = !showSidebar"
      ></div>

      <div class="logo" @click="$router.push('/')"></div>

      <div class="page">Overview</div>

      <input
        type="text"
        class="search"
        placeholder="Search for apps, namespaces, volumes, and domains"
      />

      <div class="actions"></div>

      <div class="user" @click="logout()">
        <div class="meta">
          <div class="name">{{ this.$store.state.user.name }}</div>
          <div class="username">@{{ this.$store.state.user.username }}</div>
        </div>

        <div class="profile" :style="profileStyle"></div>
      </div>
    </div>

    <div class="container">
      <div class="sidebar" v-if="showSidebar">
        <div class="category">Cluster</div>

        <div class="item">Nodes</div>
        <div class="item">Namespaces</div>
        <div class="item">Users</div>
        <div class="item">Security</div>

        <div class="category">Namespace</div>

        <select class="namespace">
          <option>default</option>
          <option>orbit-system</option>
        </select>

        <div class="item">Overview</div>
        <div class="item">Repositories</div>
        <div class="item">Deployments</div>
        <div class="item">Routers</div>
        <div class="item">Certificates</div>
        <div class="item">Volumes</div>
      </div>

      <div class="content">
        This is the content.
      </div>
    </div>
    <div class="footer">Orbit Version</div>
  </div>
</template>

<script>
import hamburgerIcon from "@/assets/icon/hamburger.svg";
import exitIcon from "@/assets/icon/exit.svg";
import defaultProfile from "@/assets/icon/blank-profile.svg";

export default {
  data() {
    return {
      showSidebar: true,
      hasProfile: false
    };
  },

  mounted() {
    this.checkProfile();
  },

  methods: {
    /**
     * This will update the use profile image to ensure that we have the latest
     * version.
     */
    async checkProfile() {
      const id = this.$store.state.user.id;
      const res = await this.$api.get(`/user/${id}/profile`);
      this.hasProfile = res.status === 200;
    },

    async logout() {
      const { user, token } = this.$store.state;
      await this.$api.delete(`/user/${user.id}/session/${token}`);
      await this.$store.dispatch("updateUser");
    }
  },

  computed: {
    sidebarToggleStyle() {
      const url = this.showSidebar ? exitIcon : hamburgerIcon;
      return { backgroundImage: `url("${url}")` };
    },

    profileStyle() {
      const id = this.$store.state.user.id;
      const url = this.hasProfile ? `/api/user/${id}/profile` : defaultProfile;
      return { backgroundImage: `url("${url}")` };
    }
  }
};
</script>

<style lang="scss" scoped>
$backgroundColor: #f5f6fa;
$borderColor: darken($backgroundColor, 5%);

.root {
  display: flex;
  flex-direction: column;
  position: absolute;
  width: 100vw;
  height: 100vh;
  left: 0;
  top: 0;

  .navbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-shrink: 0;

    background-color: #fff;
    border-bottom: solid 1px $borderColor;
    padding: 12px 20px;

    overflow-y: scroll;

    & > * {
      flex-shrink: 0;
      margin-left: 20px;
    }

    .toggle {
      margin-left: 0;
      width: 40px;
      height: 40px;
      opacity: 0.7;
      background-size: 14px 14px;
      background-position: center;
      background-repeat: no-repeat;
      cursor: pointer;

      transition: transform 0.2s;

      &:hover {
        transform: scale(1.2);
      }

      &:active {
        transform: scale(0.8);
      }
    }

    .logo {
      margin-left: 14px;
      width: 38px;
      height: 38px;
      background-size: cover;
      background-position: center;
      background-image: url("~@/assets/logo/gradient.svg");
      cursor: pointer;
    }

    .page {
      margin: 0 20px;
      font-size: 15px;
      font-weight: bold;
      opacity: 0.8;

      cursor: default;
    }

    .search {
      border: solid 1px $borderColor;
      background-color: $backgroundColor;
      padding: 10px;
      font-size: 14px;

      background-image: url("~@/assets/icon/search.svg");
      background-repeat: no-repeat;
      background-size: 16px;
      background-position: center left 10px;
      padding-left: 36px;

      width: 378px;
      flex-shrink: 1;
      margin: 0 auto;
      transition: background-color 0.2s;

      &:focus {
        background-color: #fff;
      }
    }

    .actions {
      display: flex;
      justify-content: flex-end;
      align-items: center;
    }

    .user {
      display: flex;
      align-items: center;
      padding: 5px;
      border-radius: 1000px;
      padding-left: 15px;

      cursor: pointer;

      transition: border 0.2s, background-color 0.2s;
      border: solid 1px transparent;
      &:hover {
        border: solid 1px $borderColor;
      }
      &:active {
        background-color: transparentize($borderColor, 0.6);
      }

      .profile {
        width: 38px;
        height: 38px;
        border-radius: 34px;
        background-color: #ddd;
        margin-left: 10px;

        background-size: cover;
        background-position: center;
        background-repeat: no-repeat;
      }

      .meta {
        display: flex;
        flex-direction: column;
        text-align: right;
        justify-content: center;
        align-items: flex-end;

        .name {
          font-size: 15px;
        }

        .username {
          font-size: 13px;
          font-weight: bold;
          margin-top: 2px;
        }
      }
    }
  }

  .container {
    display: flex;
    flex-grow: 1;
  }

  .sidebar {
    background-color: #fff;
    border-right: solid 1px $borderColor;
    flex-shrink: 0;
    overflow: scroll;
    width: 250px;
    padding: 20px;

    .category {
      font-size: 13px;
      margin: 10px 0;
      color: rgba(0, 0, 0, 0.5);
      font-weight: bold;
      text-transform: uppercase;

      &:not(:first-of-type) {
        margin-top: 40px;
      }
    }

    select.namespace {
      -webkit-appearance: none;
      -moz-appearance: none;
      -ms-appearance: none;
      -o-appearance: none;
      appearance: none;

      background: none;

      background-image: url("~@/assets/icon/dropdown.svg");
      background-size: 10px;
      background-position: center right 10px;
      background-repeat: no-repeat;

      font-family: "Montserrat", sans-serif;
      font-size: 14px;
      font-weight: bold;

      padding: 10px 20px;
      border: solid 1px #ddd;
      width: 100%;

      margin-bottom: 10px;

      &:focus {
        outline: none;
      }
    }

    .item {
      padding: 14px;
      border-radius: 4px;
      cursor: pointer;

      transition: background-color 0.2s;

      &:hover {
        background-color: $backgroundColor;
      }

      &:active {
        background-color: $borderColor;
      }
    }
  }

  .content {
    flex-grow: 2;
    padding: 20px;
    overflow: scroll;
    max-height: 100%;
  }

  .footer {
    background-color: #fff;
    border-top: solid 1px $borderColor;
    display: flex;
    padding: 10px 20px;
    flex-shrink: 0;
  }
}
</style>
