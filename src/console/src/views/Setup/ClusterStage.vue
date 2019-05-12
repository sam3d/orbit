<template>
  <div class="flex-center">
    <h2>Join an existing cluster</h2>
    <p class="subheader">
      Please enter the network address and the joining token for this node. You
      can find both of these under the cluster settings of the existing cluster.
    </p>

    <input
      name="address"
      class="code"
      type="text"
      placeholder="0.0.0.0:6501"
      v-model="address"
      ref="input"
      size="21"
      maxlength="21"
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      :disabled="busy"
      @input="error = ''"
      @keyup.enter="joinCluster"
    />

    <textarea
      placeholder="Paste the join token here"
      v-model="token"
      :disabled="busy"
      @keydown.enter="e => e.preventDefault()"
      @keyup.enter="joinCluster"
    />

    <div class="error" v-if="error">{{ error }}</div>

    <Button
      text="Continue"
      class="green"
      :busy="busy"
      :disabled="!valid"
      @click="joinCluster"
    />
  </div>
</template>

<script>
import validator from "validator";
import Button from "@/components/Button";

export default {
  components: { Button },

  data() {
    return {
      address: "", // The target address
      token: "jointoken", // The join token for the target cluster
      busy: false, // Whether there is any process taking place
      error: null // If there was an error with processing
    };
  },

  methods: {
    async joinCluster() {
      if (this.busy || !this.valid) return;
      this.busy = true;
    }
  },

  computed: {
    // Whether or not the input fields are valid.
    valid() {
      try {
        const [ip, port] = this.address.split(":");

        const validIP = validator.isIP(ip);
        const validPort = validator.isPort(port);
        const validToken = this.token.length >= 3; // TODO: Improve this heuristic

        return validIP && validPort && validToken;
      } catch (e) {
        return false;
      }
    }
  }
};
</script>

<style lang="scss" scoped>
input[name="address"] {
  margin-top: 20px;
}

textarea {
  margin-top: 20px;
  width: 100%;
  height: 80px;
}

.error {
  margin-top: 20px;
}

.button {
  margin-top: 20px;
}
</style>
