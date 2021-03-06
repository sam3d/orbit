<template>
  <div class="flex-center">
    <h2>Network address</h2>
    <p class="subheader">
      The is the IP address that can be used to reach this node from the
      internet. This can't be changed after the node is set up, so make sure
      this node is in the network environment it will remain in.
    </p>

    <input
      name="address"
      class="code"
      type="text"
      placeholder="0.0.0.0"
      v-model="address"
      ref="input"
      size="15"
      maxlength="15"
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      :disabled="busy"
      @input="error = ''"
      @keyup.enter="bootstrap"
    />

    <div class="error" v-if="error">{{ error }}</div>

    <a
      href="#"
      @click.prevent="address = urlIP"
      v-if="address !== urlIP && urlIP && !busy"
      >Use URL IP ({{ urlIP }})</a
    >

    <a
      href="#"
      @click.prevent="address = publicIP"
      v-if="address !== publicIP && publicIP && !busy"
      >Use Node's Public IP ({{ publicIP }})</a
    >

    <Button
      class="green"
      :disabled="!valid"
      :busy="busy"
      text="Continue"
      @click="bootstrap"
    />
  </div>
</template>

<script>
import validator from "validator";

export default {
  data() {
    return {
      address: "",
      error: "",
      busy: false
    };
  },

  mounted() {
    this.focus();
  },

  methods: {
    // Perform the bootstrap operation.
    async bootstrap() {
      if (!this.valid) return;
      this.error = "";

      this.busy = true;
      const res = await this.$api.post(
        "/cluster/bootstrap",
        { advertise_address: this.address },
        { redirect: false }
      );
      this.busy = false;

      if (res.status === 200) {
        this.$emit("complete");
        return;
      }

      this.error = res.data;
      this.focus();
    },

    // Focus on the main input element.
    async focus() {
      await this.$nextTick();
      this.$refs.input.focus();
    }
  },

  computed: {
    urlIP() {
      const [ip] = window.location.host.split(":");
      if (validator.isIP(ip)) return ip;
    },

    publicIP() {
      return this.$store.state.ip;
    },

    valid() {
      return validator.isIP(this.address);
    }
  }
};
</script>

<style lang="scss" scoped>
.button {
  margin-top: 34px;
}

input[name="address"] {
  margin-top: 34px;
}

a {
  margin-top: 15px;
}

.error {
  margin: 20px 0;
  max-width: 400px;
}
</style>
