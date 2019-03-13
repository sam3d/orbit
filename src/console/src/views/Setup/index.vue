<template>
  <section class="setup">
    <nav>
      <div class="logo" @click="page = 0">
        <img src="@/assets/logo/gradient-thick.svg" />
        <span>Orbit</span>
      </div>

      <a href="https://docs.orbit.sh">Docs</a>
    </nav>

    <div class="body">
      <div class="inner">
        <template v-if="page === 0">
          <h1 class="large">Welcome to Orbit</h1>

          <p class="large">
            This node still needs to be configured before it is able to join or
            start a cluster. Please configure it as soon as possible, as this
            page is publicly accessible over the web to anybody with the link.
          </p>

          <div class="button purple" @click="page = 1">Start setup</div>
        </template>
      </div>
    </div>

    <ProgressView :stages="stages" :hidden="page === 0" />
  </section>
</template>

<script>
import ProgressView from "./Progress";

export default {
  meta: { title: "Setup" },

  components: { ProgressView },

  data() {
    return {
      page: 0,
      stages: [
        { name: "Welcome", state: "complete" },
        { name: "Mode", state: "active" },
        { name: "Domain", state: "incomplete" },
        { name: "User", state: "incomplete" },
        { name: "complete", state: "incomplete" }
      ]
    };
  }
};
</script>

<style lang="scss">
section.setup {
  width: 100%;
  height: 100%;
  left: 0;
  top: 0;
  position: fixed;

  display: flex;
  flex-direction: column;

  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px;
    flex-shrink: 0;

    .logo {
      cursor: pointer;

      display: flex;
      align-items: center;

      img {
        height: 40px;
      }

      span {
        margin-left: 10px;
        font-family: "Cabin", sans-serif;
        font-size: 24px;
        margin-top: 3px; // The text is a bit weird, so this centers it
      }
    }
  }

  .body {
    text-align: center;
    flex-grow: 1;
    padding: 20px;

    display: flex;
    text-align: center;
    overflow-y: scroll;

    .inner {
      margin: auto;
    }
  }

  h1.large {
    font-size: 45px;
  }

  p.large {
    max-width: 600px;
    margin: 50px 0;
    font-size: 18px;
    line-height: 2rem;

    animation-delay: 0.3s;
  }

  .button {
    animation-delay: 0.6s;
  }
}
</style>
