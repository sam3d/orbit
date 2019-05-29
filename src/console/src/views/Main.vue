<template>
  <div class="root">
    <div class="navbar">
      <div
        class="toggle"
        :style="sidebarToggleStyle"
        @click="showSidebar = !showSidebar"
      ></div>

      <div class="logo" @click="$push('/')"></div>

      <transition name="slide" mode="out-in">
        <div class="page" :key="$store.state.title">
          {{ $store.state.title }}
        </div>
      </transition>

      <input
        type="text"
        class="search"
        placeholder="Search for apps, namespaces, volumes, and domains"
        ref="search"
      />

      <div class="actions">
        <div
          class="create"
          :class="{ open: showCreateMenu }"
          @click="showCreateMenu = !showCreateMenu"
        >
          <span>Create</span>
          <img src="@/assets/icon/dropdown-white.svg" />

          <div class="menu-background" v-if="showCreateMenu"></div>
          <transition name="slide-down">
            <div class="menu" v-if="showCreateMenu">
              <div class="item" @click="$push('/security')">
                <h4>Node</h4>
                <p>Add a new node to the cluster</p>
              </div>
              <div class="item" @click="$push('/namespaces/new')">
                <h4>Namespace</h4>
                <p>Create a new namespace</p>
              </div>
              <div class="item" @click="$push('/users/new')">
                <h4>User</h4>
                <p>Sign up a new user</p>
              </div>
              <div class="separator"></div>
              <div class="item" @click="$push('/repositories/new')">
                <h4>Repository</h4>
                <p>Create a new git repository</p>
              </div>
              <div class="item" @click="$push('/deployments/new')">
                <h4>Deployment</h4>
                <p>Deploy a new service</p>
              </div>
              <div class="item" @click="$push('/routers/new')">
                <h4>Router</h4>
                <p>Create a new ingress router</p>
              </div>
              <div class="item" @click="$push('/certificates/new')">
                <h4>Certificate</h4>
                <p>Provision a new SSL certificate</p>
              </div>
              <div class="item" @click="$push('/volumes/new')">
                <h4>Volume</h4>
                <p>Create a new block storage volume</p>
              </div>
            </div>
          </transition>
        </div>
      </div>

      <div
        class="user"
        @click="showUserMenu = !showUserMenu"
        :class="{ open: showUserMenu }"
      >
        <div class="meta">
          <div class="name">{{ this.$store.state.user.name }}</div>
          <div class="username">@{{ this.$store.state.user.username }}</div>
        </div>

        <div class="profile" :style="profileStyle"></div>
        <div class="arrow"></div>

        <div class="menu-background" v-if="showUserMenu"></div>
        <transition name="slide-down">
          <div class="menu" v-if="showUserMenu">
            User menu
          </div>
        </transition>
      </div>
    </div>

    <div class="container">
      <transition name="slide">
        <div class="sidebar" v-if="showSidebar">
          <div class="category">Cluster</div>

          <div class="item" @click="$push('/nodes')">Nodes</div>
          <div class="item" @click="$push('/namespaces')">Namespaces</div>
          <div class="item" @click="$push('/users')">Users</div>
          <div class="item" @click="$push('/security')">Security</div>

          <div class="category">Namespace</div>

          <select
            class="namespace"
            v-model="namespace"
            @click="fetchNamespaces"
          >
            <option>default</option>
            <option v-for="namespace in namespaces" :value="namespace.id">
              {{ namespace.name }}
            </option>
          </select>

          <div class="item" @click="$push('/')">Overview</div>
          <div class="item" @click="$push('/repositories')">Repositories</div>
          <div class="item" @click="$push('/deployments')">Deployments</div>
          <div class="item" @click="$push('/routers')">Routers</div>
          <div class="item" @click="$push('/certificates')">Certificates</div>
          <div class="item" @click="$push('/volumes')">Volumes</div>
        </div>
      </transition>

      <div class="content">
        <transition mode="out-in" name="fade">
          <router-view :key="namespace + $reloadKey.current"></router-view>
        </transition>
      </div>
    </div>

    <div class="footer">
      <div class="version-container">
        <div class="version">Orbit v0.1.0</div>
        <div class="date">Up to date</div>
      </div>

      <div class="status-container">
        <div class="status">Cluster is healthy</div>
        <div class="dot green"></div>
      </div>
    </div>

    <transition name="slow-slide">
      <div class="slider" v-if="showSlider">
        <div class="content-container">
          <div class="content">
            <router-view name="slider"></router-view>
          </div>
          <div class="close" @click="up"></div>
        </div>
        <div class="background" @click="up"></div>
      </div>
    </transition>
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
      showUserMenu: false,
      showCreateMenu: false,
      hasProfile: false,
      namespace: "default", // Keep track of the selected namespace

      namespaces: [] // Keep track of the current namespaces
    };
  },

  mounted() {
    this.checkProfile();
    this.fetchNamespaces();
    this.namespace = this.$route.query.namespace || "default";

    // Handle global keydown listener.
    window.addEventListener("keydown", this.keydownHandler);
  },

  beforeDestroy() {
    window.removeEventListener("keydown", this.keydownHandler);
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

    // Listener function for keydown events.
    keydownHandler(e) {
      const { search } = this.$refs;

      // Escape key.
      if (e.keyCode == 27) {
        // Hide the slider.
        if (this.showSlider) {
          e.preventDefault();
          this.up();
          return;
        }

        // Exit the search bar.
        if (search && document.activeElement === search) {
          e.preventDefault();
          search.blur();
        }
      }

      // Forward slash.
      if (e.keyCode == 191) {
        return; // TODO: Re-implement this correctly.
        if (!search || document.activeElement === search) return;
        e.preventDefault();
        search.focus();
      }
    },

    // Get a list of the namespaces.
    async fetchNamespaces() {
      const res = await this.$api.get("/namespaces");
      this.namespaces = res.data;
    },

    // Navigate up a path element in the URL.
    up() {
      const url = this.$route.path;
      const elements = url.split("/");
      this.$push(
        this.$route.path
          .split("/")
          .slice(0, -1)
          .join("/")
      );
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
    },

    showSlider() {
      return this.$route.matched.some(route =>
        route.components.hasOwnProperty("slider")
      );
    }
  },

  watch: {
    // Watch the namespace property and update the URL if it changes.
    namespace(namespace) {
      // Update the store with the selected namespace.
      const found = this.namespaces.find(ns => ns.id === this.namespace);
      const name = found ? found.name : "";
      this.$store.commit("namespace", { id: namespace, name });

      // Remove the query parameter completely if it is default.
      if (namespace === "default") namespace = undefined;

      this.$router.push({
        query: {
          ...this.$route.query,
          namespace
        }
      });
    },

    // Keep the namespace name updated.
    namespaces(namespaces) {
      const namespace = this.namespaces.find(ns => ns.id === this.namespace);
      if (!namespace) return;
      this.$store.commit("namespace", {
        id: this.$namespace(),
        name: namespace.name
      });
    }
  }
};
</script>

<style lang="scss">
@keyframes placeholder {
  from {
    background-position: 200%;
  }
  to {
    background-position: 0%;
  }
}

@keyframes wobble {
  0% {
    opacity: 0.1;
  }
  50% {
    opacity: 0.4;
  }
  100% {
    opacity: 0.1;
  }
}

.root .content {
  h1 {
    font-size: 30px;
    margin: 20px 0;
  }

  h2 {
    font-size: 14px;
    letter-spacing: 0.05rem;
    font-weight: bold;
    opacity: 0.6;
    text-transform: uppercase;
    margin-top: 30px;
    margin-bottom: 10px;

    &.placeholder {
      animation: wobble 1s linear infinite;
    }
  }

  .list {
    margin: 30px 0;

    .item {
      background-color: #fff;
      padding: 20px;
      border-radius: 4px;
      cursor: pointer;
      overflow: hidden;

      & > * {
        flex-shrink: 0;
      }

      display: flex;
      align-items: center;

      transition: all 0.2s;
      box-shadow: 0 2px 5px 0 rgba(0, 0, 0, 0.025);
      &:not(.placeholder):hover {
        box-shadow: 0 3px 8px 0 rgba(0, 0, 0, 0.05);
        transform: translateY(-1px);
      }
      &:not(.placeholder):active {
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
        transform: translateY(1px);
      }

      &:not(:last-of-type) {
        margin-bottom: 10px;
      }

      &.placeholder {
        cursor: default;
        box-shadow: none;
        background: linear-gradient(
          90deg,
          rgba(0, 0, 0, 0),
          rgba(0, 0, 0, 0.1),
          rgba(0, 0, 0, 0)
        );
        background-size: 200%;
        animation: placeholder 1s linear infinite;
      }
    }
  }
}
</style>

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
    padding: 14px 20px;

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
      min-width: 115px;

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

      .create {
        background-color: #8959ea;
        color: #fff;
        padding: 10px;
        border-radius: 4px;
        display: flex;
        align-items: center;
        cursor: pointer;
        transition: opacity 0.2s;

        &:hover {
          opacity: 0.9;
        }

        &.open {
          opacity: 1;
          img {
            transform: rotate(180deg);
          }
        }

        img {
          width: 8px;
          margin-left: 8px;
          transition: transform 0.3s;
        }
      }
    }

    .user {
      display: flex;
      align-items: center;

      cursor: pointer;

      transition: opacity 0.2s;
      &:hover {
        opacity: 0.8;
      }
      &.open {
        opacity: 1;
        .arrow {
          transform: rotate(180deg);
        }
      }

      .arrow {
        width: 10px;
        height: 10px;
        background-image: url("~@/assets/icon/dropdown.svg");
        background-repeat: no-repeat;
        background-position: center;
        background-size: contain;
        margin-left: 10px;

        transition: transform 0.3s;
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
          margin-top: 5px;
        }
      }
    }
  }

  .container {
    display: flex;
    flex-grow: 1;

    .content {
      flex-grow: 2;
      padding: 20px;
      overflow: scroll;
      max-height: 100%;
      max-width: 1200px;
      margin: 0 auto;
    }
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
      margin-bottom: 10px;
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

  .slider {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: flex-end;
    box-shadow: 0 0 15px 0 rgba(0, 0, 0, 0.5);
    z-index: 999;

    .background {
      background-color: rgba(0, 0, 0, 0.2);
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      z-index: 999;
    }

    .content-container {
      background-color: #fff;
      z-index: 1000;
      display: flex;
      max-width: 100vw;

      .close {
        position: absolute;
        width: 30px;
        height: 30px;
        right: 20px;
        top: 20px;
        background-image: url("~@/assets/icon/exit.svg");
        background-size: 15px;
        background-repeat: no-repeat;
        background-position: center;
        cursor: pointer;
        flex-shrink: 0;

        transition: all 0.2s;
        opacity: 0.6;
        &:hover {
          transform: scale(1.1);
          opacity: 0.7;
        }
        &:active {
          transform: scale(0.9);
        }
      }

      .content {
        padding: 70px;
        overflow: scroll;
      }
    }
  }

  .footer {
    background-color: #fff;
    border-top: solid 1px $borderColor;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 20px;
    flex-shrink: 0;
    font-size: 14px;

    & > * {
      display: flex;
      align-items: center;
    }

    .version-container {
      .date {
        margin-left: 10px;
        text-transform: uppercase;
        font-size: 11px;
        font-weight: bold;
        opacity: 0.4;
      }
    }

    .status-container {
      .dot {
        width: 8px;
        height: 8px;
        border-radius: 8px;
        background-color: #aaa;
        margin-left: 8px;

        &.green {
          background-color: #1dd1a1;
        }
      }
    }
  }
}

// Styles for a menu that opens from the navbar.
.menu-background {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  width: 100%;
  z-index: 999;
  cursor: default;
}

.menu {
  z-index: 1500;
  position: absolute;
  background-color: #fff;
  color: #151515;
  padding: 5px;
  border-radius: 4px;
  top: 65px;
  right: 20px;
  border: solid 1px $borderColor;
  box-shadow: 0 2px 5px 0 rgba(0, 0, 0, 0.05);

  .separator {
    height: 1px;
    margin: 10px 0;
    width: 100%;
    background-color: $borderColor;
    border-radius: 2px;
  }

  .item {
    padding: 10px;
    border-radius: 4px;
    border: solid 1px transparent;
    transition: all 0.2s;

    h4 {
      font-size: 15px;
      font-weight: bold;
    }

    p {
      margin-top: 4px;
      font-size: 13px;
      opacity: 0.8;
      white-space: nowrap;
    }

    &:hover {
      background-color: #fff;
      transform: scale(1.08);
      border: solid 1px $borderColor;
      box-shadow: 0 2px 5px 0 rgba(0, 0, 0, 0.1);
    }
    &:active {
      transform: scale(0.98);
      border: solid 1px transparent;
      box-shadow: none;
    }
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.fade-enter,
.fade-leave-active {
  opacity: 0;
  transform: scale(0.98);
}

.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.slide-enter,
.slide-leave-active {
  opacity: 0;
  transform: translateX(-5px);
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.slide-down-enter,
.slide-down-leave-active {
  opacity: 0;
  transform: translateY(-5px);
}

.slow-slide-enter-active,
.slow-slide-leave-active {
  transition: opacity 0.5s;
  .content-container {
    transition: transform 0.5s;
  }
}
.slow-slide-enter,
.slow-slide-leave-active {
  opacity: 0;
  .content-container {
    transform: translateX(400px);
  }
}
</style>
