<template>
  <div class="root">
    <div class="navbar">
      <div
        class="toggle"
        :style="sidebarToggleStyle"
        @click="showSidebar = !showSidebar"
      ></div>

      <div class="logo" @click="$router.push('/')"></div>

      <div class="page">Applications</div>

      <input
        type="text"
        class="search"
        placeholder="Search for apps, namespaces, volumes, and domains"
      />

      <div class="actions">
        Actions
      </div>

      <div class="user">
        <div class="meta">
          <div class="name">{{ this.$store.state.user.name }}</div>
          <div class="username">@{{ this.$store.state.user.username }}</div>
        </div>

        <div class="profile"></div>
      </div>
    </div>

    <div class="container">
      <div class="sidebar" v-if="showSidebar">
        This is the sidebar.
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

export default {
  data() {
    return {
      showSidebar: true
    };
  },
  computed: {
    sidebarToggleStyle() {
      const url = this.showSidebar ? exitIcon : hamburgerIcon;
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
    padding: 20px;

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

      width: 400px;
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

      .profile {
        width: 38px;
        height: 38px;
        border-radius: 34px;
        background-color: #ddd;
        margin-left: 10px;
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
    padding: 20px;
    flex-shrink: 0;
    overflow: scroll;
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
