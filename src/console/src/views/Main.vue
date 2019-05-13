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

      <input type="text" class="search" placeholder="Search..." />

      <div class="actions">
        The action icons
      </div>
      <div class="user">
        The user data
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
      margin-left: 20px;
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
      flex-grow: 1;

      transition: background-color 0.2s;

      &:focus {
        background-color: #fff;
      }
    }

    .actions {
    }

    .user {
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
