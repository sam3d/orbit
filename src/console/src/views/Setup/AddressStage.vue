<template>
  <div class="flex-center">
    <h2>Network address</h2>
    <p class="subheader">
      The is the IP address that can be used to reach this node from the
      internet. This can't be changed after the node is set up, so make sure
      this node is in the network environment it will remain in.
    </p>

    <input type="text" placeholder="0.0.0.0" v-model="address" ref="input" />

    <a
      href="#"
      @click.prevent="address = urlIP"
      v-if="address !== urlIP && urlIP"
      >Use URL IP ({{ urlIP }})</a
    >

    <Button
      class="green"
      :class="{ disabled: !valid }"
      text="Continue"
      :busy="busy"
      @click="busy = !busy"
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
      address: "",
      busy: false
    };
  },

  mounted() {
    this.$refs.input.focus();
  },

  computed: {
    urlIP() {
      const [ip] = window.location.host.split(":");
      if (validator.isIP(ip)) return ip;
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

input {
  margin-top: 34px;
  font-size: 20px;
}

a {
  margin-top: 15px;
}
</style>
