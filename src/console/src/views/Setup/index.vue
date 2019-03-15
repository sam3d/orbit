<template>
  <section class="setup">
    <Navbar />

    <div class="body">
      <div class="inner">
        <transition name="fade" mode="out-in">
          <!-- The very first stage with a simply progression button -->
          <WelcomeStage v-if="stage === 'welcome'" @complete="stage = 'mode'" />

          <!-- Choose whether to bootstrap a cluster or simply join one -->
          <ModeStage v-if="stage === 'mode'" @complete="changeMode" />

          <!-- Choose the address that this node operates on -->
          <AddressStage v-if="stage === 'address'" @complete="nextStage" />

          <!-- Choose a domain and certificate -->
          <DomainStage v-if="stage === 'domain'" />
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
import Navbar from "./Navbar";

import WelcomeStage from "./WelcomeStage";
import ModeStage from "./ModeStage";
import AddressStage from "./AddressStage";
import DomainStage from "./DomainStage";

export default {
  meta: { title: "Setup" },
  components: {
    ProgressView,
    Navbar,

    WelcomeStage,
    ModeStage,
    AddressStage,
    DomainStage
  },

  data() {
    const state = this.$store.state.init;

    return {
      stage: state.stage,
      mode: state.mode // bootstrap || join
    };
  },

  methods: {
    // changeMode will change whether or not we are bootstrapping this node or
    // joining another cluster.
    changeMode(mode) {
      this.mode = mode;
      this.nextStage();
    },

    // nextStage checks the current stages and navigates to the next logical
    // stage in the process.
    nextStage() {
      let names = this.stageNames;
      let i = names.indexOf(this.stage);
      this.stage = names[++i];
    }
  },

  computed: {
    // stageNames simply returns an array of strings containing the stages that
    // should be present given the selected mode.
    stageNames() {
      let names = ["welcome", "mode"]; // Will always use these first two stages.

      if (this.mode === "bootstrap") names.push("address", "domain", "user");
      else if (this.mode === "join") names.push("cluster");

      names.push("node", "complete"); // Will always have a last stage.
      return names;
    },

    /**
     * Stages returns an object containing the stages that the progress bar
     * needs to show. It figures out what the current page is and therefore the
     * overall progress indicators for each icon. It derives its state from
     * this.stageNames().
     */
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

  .body {
    text-align: center;
    flex-grow: 1;
    padding: 30px;
    padding-bottom: 50px;

    display: flex;
    text-align: center;
    overflow-y: scroll;

    .inner {
      margin: auto;
    }
  }

  //
  // Styles used in all components.
  //

  h1 {
    font-size: 45px;
  }

  h2 {
    font-size: 32px;
  }

  p {
    font-size: 18px;
    line-height: 1.8rem;
  }

  p.large {
    max-width: 600px;
    margin: 50px 0;
    line-height: 2rem;
  }

  p.subheader {
    margin: 0 auto;
    margin-top: 24px;
    max-width: 600px;
  }
}
</style>
