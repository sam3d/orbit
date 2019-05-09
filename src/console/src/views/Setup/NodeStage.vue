<template>
  <div>
    <h2>Configure this node</h2>
    <p class="subheader">
      Decide what purpose this node serves in the cluster, and also configure
      its swap space and swappiness.
    </p>

    <h3>What kind of node should this be?</h3>
    <div class="notice" v-if="mode === 'bootstrap'">
      As this is the first node in the cluster, it must be a manager. This can
      be changed later.
    </div>
    <div class="options">
      <div
        class="option"
        :class="{ selected: type === 'manager', disabled: busy }"
        @click="setType('manager')"
      >
        <h4>Manager</h4>
        <p>
          This node would be responsible for maintaining the state of the
          cluster and performing operations on it.
        </p>
      </div>

      <div
        class="option"
        :class="{
          selected: type === 'worker',
          disabled: mode === 'bootstrap' || busy
        }"
        @click="setType('worker')"
      >
        <h4>Worker</h4>
        <p>
          This would allow the node to perform compute, storage and ingress
          operations without being able to make decisions.
        </p>
      </div>
    </div>

    <h3>What other functions should this node perform?</h3>
    <div class="notice" v-if="mode === 'bootstrap'">
      As this is the first node in the cluster, it must be able to perform all
      three of operations listed below. These can be changed as more nodes are
      added to the cluster.
    </div>
    <div class="options">
      <div
        class="option"
        :class="{
          selected: roles.includes('LOAD_BALANCER'),
          disabled: busy || mode === 'bootstrap'
        }"
        @click="toggleRole('LOAD_BALANCER')"
      >
        <h4>Load balancer</h4>
        <p>
          This node would handle all web traffic that comes into the cluster.
        </p>
      </div>

      <div
        class="option"
        :class="{
          selected: roles.includes('STORAGE'),
          disabled: busy || mode === 'bootstrap'
        }"
        @click="toggleRole('STORAGE')"
      >
        <h4>Storage</h4>
        <p>
          This node would be generally responsible for user and system storage.
        </p>
      </div>

      <div
        class="option"
        :class="{
          selected: roles.includes('BUILDER'),
          disabled: busy || mode === 'bootstrap'
        }"
        @click="toggleRole('BUILDER')"
      >
        <h4>Builder</h4>
        <p>
          This node would be responsible for building the docker images for your
          apps.
        </p>
      </div>
    </div>

    <h3>How much swap space should this node have?</h3>
    <p class="info">
      This node has a built-in amount of hardware
      <a
        href="https://en.wikipedia.org/wiki/Random-access_memory"
        target="_blank"
        >RAM</a
      >. Adding swap increases the
      <a href="https://en.wikipedia.org/wiki/Virtual_memory" target="_blank"
        >virtual memory</a
      >
      of this node. A swappiness setting of zero means that it will only ever be
      used if it is <i>absolutely</i> necessary, where 100 will swap programs to
      the virtual memory almost instantly. Both of these values are already set
      to the recommended values based on the size and type of your disk.
    </p>
  </div>
</template>

<script>
export default {
  props: [
    "mode" // bootstrap / join. If bootstrap, all options are mandatory.
  ],

  data() {
    return {
      type: this.mode === "bootstrap" ? "manager" : "worker",
      roles: ["LOAD_BALANCER", "STORAGE", "BUILDER"],
      busy: false // Whether a process is taking place
    };
  },

  methods: {
    setType(type) {
      if (this.busy) return; // Don't change type if busy
      if (this.mode === "bootstrap") return; // Must leave as manager
      this.type = type;
    },

    /**
     * Toggle the role by adding it or removing it from the role array,
     * depending on the desired action.
     */
    toggleRole(role) {
      if (this.busy) return; // Don't allow for selections if busy
      if (this.mode === "bootstrap") return; // Must leave all selections
      if (!this.roles.includes(role)) this.roles.push(role);
      else this.roles = this.roles.filter(r => r !== role);
    }
  }
};
</script>

<style lang="scss" scoped>
h3 {
  font-size: 20px;
  font-weight: bold;
  margin-top: 50px;
  margin-bottom: 20px;
}

.notice {
  margin-bottom: 20px;
  color: #ff9f43;
  background-color: transparentize(#ff9f43, 0.9);
  display: inline-block;
  font-weight: bold;
  padding: 20px;
  border-radius: 4px;
  max-width: 500px;
  line-height: 1.6rem;
  cursor: default;
}

.info {
  margin: 0 auto;
  max-width: 800px;
  font-size: 16px;
  line-height: 1.4rem;
  opacity: 1;
  cursor: default;
}

.options {
  display: grid;
  grid-gap: 30px;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));

  .option {
    background-color: #fff;
    padding: 30px;
    border-radius: 4px;
    cursor: pointer;

    h4 {
      font-size: 18px;
      font-weight: bold;
      transition: color 0.2s;
    }

    p {
      margin-top: 10px;
      font-size: 16px;
      line-height: 1.4rem;
    }

    transition: transform 0.2s, border 0.2s;
    &:not(.disabled):hover {
      transform: scale(1.05);
    }
    &:not(.disabled):active {
      transform: scale(0.95);
    }

    border: solid 1px transparent;
    &.selected {
      border: solid 1px #1dd1a1;
      h4 {
        color: #1dd1a1;
      }
    }

    &.disabled {
      cursor: not-allowed;
      opacity: 0.5;
    }
  }
}
</style>
