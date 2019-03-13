<template>
  <section class="setup">
    <nav>
      <div class="logo">
        <img src="@/assets/logo/gradient.svg" />
        <span class="name">Orbit</span>
      </div>

      <ul>
        <li><a href="https://orbit.sh/docs">Read the docs</a></li>
        <li><a href="https://orbit.sh/support">Support</a></li>
      </ul>
    </nav>

    <div class="body">
      <div class="inner">
        <transition name="fade" mode="out-in">
          <div key="welcome" v-if="stage === 'welcome'">
            <h1 class="large">Welcome to Orbit</h1>

            <p class="large">
              This node still needs to be configured before it is able to join
              or start a cluster. Please configure it as soon as possible, as
              this page is publicly accessible over the web to anybody with the
              link.
            </p>

            <div class="button purple" @click="stage = 'mode'">
              Start setup
            </div>
          </div>

          <div key="mode" v-if="stage === 'mode'"></div>
        </transition>
      </div>
    </div>

    <ProgressView
      v-model="stage"
      :stages="stages"
      :hidden="stage === 'welcome'"
    />
  </section>
</template>

<script>
import ProgressView from "./Progress";

export default {
  meta: { title: "Setup" },
  components: { ProgressView },

  data() {
    return {
      stage: "welcome",
      mode: "bootstrap" // bootstrap || join
    };
  },

  computed: {
    stageNames() {
      let names = ["welcome", "mode"]; // Will always use these first two stages.

      if (this.mode === "bootstrap") names.push("address", "domain", "user");
      else if (this.mode === "join") names.push("cluster");

      names.push("complete"); // Will always have a last stage.
      return names;
    },

    stages() {
      const stages = [];

      let markAsIncomplete = false;
      for (name of this.stageNames) {
        let state = "complete"; // Assume all stages complete by default.

        // If this is the current stage, everything that comes after must now be
        // marked as incomplete, except for the current stage, which is marked
        // as "active".
        if (name === this.stage) {
          state = "active";
          markAsIncomplete = true;
        } else if (markAsIncomplete) {
          state = "incomplete";
        }

        stages.push({ name, state });
      }

      return stages;
    }
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

  .fade-enter-active,
  .fade-leave-active {
    transition: opacity 0.3s;
  }

  .fade-enter,
  .fade-leave-active {
    opacity: 0;
  }

  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px;
    flex-shrink: 0;

    .logo {
      cursor: default;

      display: flex;
      align-items: center;

      img {
        height: 50px;
      }

      span.name {
        margin-left: 15px;
        font-family: "Cabin", sans-serif;
        font-weight: 500;
        opacity: 0.8;
        font-size: 34px;
        margin-top: 3px; // The text is too high, so this centers it
      }
    }

    ul li {
      display: inline-block;
      &:not(:last-of-type) {
        margin-right: 24px;
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
