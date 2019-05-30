<template>
  <div class="flex-center">
    <h2>Join an existing cluster</h2>
    <p class="subheader">
      Please enter the network address and the joining token for this node. You
      can find both of these under the cluster security settings of the existing
      cluster.
    </p>

    <input
      name="address"
      class="code"
      type="text"
      placeholder="0.0.0.0"
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

export default {
  data() {
    return {
      address: "", // The target address
      token: "", // The join token for the target cluster
      busy: false, // Whether there is any process taking place
      error: null // If there was an error with processing
    };
  },

  methods: {
    async joinCluster() {
      if (this.busy || !this.valid) return;
      this.error = null;
      this.busy = true;

      // Use a default port if one isn't present.
      let target_address = this.address;
      if (!this.port) {
        this.address = this.ip;
        target_address = this.ip + ":6501";
      }

      // Construct and then send the request.
      const body = { join_token: this.token, target_address };
      const opts = { redirect: false };
      const res = await this.$api.post("/cluster/join", body, opts);

      // Validate the response.
      if (res.status !== 200) {
        this.error = res.data;
        this.busy = false;
        return;
      }

      // Otherwise, we have successfully joined the cluster. Continue to the
      // node configuration set up screen.
      this.$emit("complete");
    }
  },

  computed: {
    // Whether or not the input fields are valid.
    valid() {
      const validToken = this.token.length >= 3; // TODO: Improve this heuristic
      return this.ip && validToken;
    },

    ip() {
      try {
        const [ip] = this.address.split(":");
        if (validator.isIP(ip)) return ip;
        else return "";
      } catch (err) {
        return "";
      }
    },

    port() {
      try {
        const [, port] = this.address.split(":");
        if (validator.isPort(port)) return port;
        else return "";
      } catch (err) {
        return "";
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
  height: 50px;
}

.error {
  margin-top: 20px;
}

.button {
  margin-top: 20px;
}
</style>
