<template>
  <div>
    <h2>Configure this node</h2>
    <p class="subheader">
      Decide what purpose this node serves in the cluster, and also configure
      its swap space and swappiness.
    </p>

    <h3>What kind of node should this be?</h3>
    <div class="notice" v-if="mode === 'bootstrap'">
      As this is the first node in the cluster, it must be a manager.
    </div>
    <div class="options">
      <div
        class="option"
        :class="{ selected: type === 'manager' }"
        @click="setType('manager')"
      >
        <h4>Manager</h4>
        <p>
          Those node is responsible for maintaining the state of the cluster and
          performing operations on it.
        </p>
      </div>

      <div
        class="option"
        :class="{ selected: type === 'worker', disabled: mode === 'bootstrap' }"
        @click="setType('worker')"
      >
        <h4>Worker</h4>
        <p>
          This would allow the node to perform compute, storage and ingress
          operations without being able to make decisions.
        </p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: [
    "mode" // bootstrap / join. If bootstrap, all options are mandatory.
  ],

  data() {
    return {
      type: "manager",
      busy: false // Whether a process is taking place
    };
  },

  methods: {
    setType(type) {
      if (this.busy || (type !== "manager" && this.mode === "bootstrap"))
        return;
      this.type = type;
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
  padding: 20px;
  border: solid 1px #ff9f43;
  border-radius: 4px;
  cursor: default;
}

.options {
  display: grid;
  grid-gap: 30px;
  grid-template-columns: repeat(2, 1fr);

  .option {
    background-color: #fff;
    padding: 30px;
    border-radius: 4px;
    max-width: 350px;
    cursor: pointer;

    h4 {
      font-size: 18px;
      font-weight: bold;
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
