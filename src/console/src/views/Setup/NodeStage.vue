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
    <div class="notice" v-if="swapSize === 0">
      You have disabled swap completely. This is not recommended, but can be
      changed later.
    </div>
    <div class="slider">
      <h4>Swap size</h4>
      <p>
        This is how much virtual memory to allocate for this node. Based on the
        node storage size and class, the recommended is amount of swap space is
        {{ swapSizeLabel(defaultSwapSize) }}.
      </p>
      <Slider
        tooltip="always"
        :data="swapSizes.map(item => item.label)"
        :value="swapSizeLabel(swapSize)"
        @change="
          e => (swapSize = swapSizes.find(item => item.label === e).value)
        "
        marks
        included
        lazy
        :disabled="busy"
      />
      <a
        href="#"
        :style="{
          visibility:
            swapSize !== defaultSwapSize && !busy ? 'visible' : 'hidden'
        }"
        @click.prevent="swapSize = defaultSwapSize"
      >
        Use recommended size of {{ swapSizeLabel(defaultSwapSize) }}
      </a>
    </div>

    <div class="slider">
      <h4>Swappiness</h4>
      <p>
        This is how likely the system is to use the swap space. A value of 0
        means that it will only ever use the swap space if its absolutely
        required. A value of 100 means that it will use the swap all the time. A
        value of {{ defaultSwappiness }} is recommended.
      </p>
      <Slider
        v-model="swappiness"
        tooltip="always"
        lazy
        marks
        :interval="10"
        :disabled="busy || swapSize === 0"
      />
      <a
        href="#"
        :style="{
          visibility:
            swappiness !== defaultSwappiness && !busy ? 'visible' : 'hidden'
        }"
        @click.prevent="swappiness = defaultSwappiness"
      >
        Use recommended swappiness of {{ defaultSwappiness }}
      </a>
    </div>

    <Button class="green" text="Continue" :busy="busy" @click="updateNode" />
  </div>
</template>

<script>
import "vue-slider-component/theme/default.css";
import Slider from "vue-slider-component";

export default {
  components: { Slider },

  props: [
    "mode" // bootstrap / join. If bootstrap, all options are mandatory.
  ],

  data() {
    return {
      type: this.mode === "bootstrap" ? "manager" : "worker",
      roles: ["LOAD_BALANCER", "STORAGE", "BUILDER"],
      swappiness: 0, // Value between 0 and 100
      swapSize: 0, // The swap size in MB

      defaultSwapSize: 2048, // The starting swap size
      defaultSwappiness: 60, // The starting swappiness

      busy: false, // Whether a process is taking place

      swapSizes: [
        { value: 0, label: "None" },
        { value: 128, label: "128MB" },
        { value: 256, label: "256MB" },
        { value: 512, label: "512MB" },
        { value: 1024, label: "1GB" },
        { value: 2048, label: "2GB" },
        { value: 4096, label: "4GB" },
        { value: 8192, label: "8GB" },
        { value: 12288, label: "12GB" },
        { value: 16384, label: "16GB" }
      ]
    };
  },

  mounted() {
    this.swapSize = this.defaultSwapSize;
    this.swappiness = this.defaultSwappiness;
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
    },

    /**
     * This is the actual method that sends the node data to the server.
     */
    async updateNode() {
      if (this.busy) return;
      this.busy = true;

      // Construct the request so that the API can handle the roles.
      const body = {
        node_roles: [...this.roles, this.type.toUpperCase()],
        swappiness: this.swappiness,
        swap_size: this.swapSize
      };

      // Make the request.
      const opts = { redirect: false };
      const res = await this.$api.put("/node/current", body, opts);
      if (res.status !== 200) {
        this.busy = false;
        alert(res.data);
        return;
      }

      // If succeeded, continue on to the last stage.
      this.$emit("complete");
    },

    swapSizeLabel(value) {
      return this.swapSizes.find(item => item.value === value).label;
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

.slider {
  margin: 0 auto;
  margin-bottom: 20px;

  background-color: #fff;
  padding: 40px;
  padding-bottom: 15px;
  border-radius: 4px;
  max-width: 650px;

  h4 {
    font-weight: bold;
    font-size: 17px;
    margin-bottom: 10px;
  }

  p {
    margin-bottom: 34px;
    font-size: 16px;
    line-height: 1.4rem;
  }

  a {
    display: block;
    margin-top: 35px;
  }
}

.button {
  margin-top: 20px;
}
</style>
