<template>
  <div>
    <h2>Domain &amp; security</h2>
    <p class="subheader">
      The domain name that you use to access Orbit. This includes the main
      dashboard and also where any git repositories will be hosted. This should
      be secured before adding your first user account.
    </p>

    <div class="domain-group">
      <input
        name="domain"
        type="text"
        placeholder="orbit.example.com"
        v-model="domain"
        ref="input"
        size="30"
      />

      <span>checking</span>
    </div>

    <a
      href="#"
      @click.prevent="domain = urlDomain"
      v-if="urlDomain && urlDomain !== domain"
      >Use URL Domain ({{ urlDomain }})</a
    >

    <Button
      class="green"
      :class="{ disabled: !validDomain }"
      text="Continue"
      @click="processing = !processing"
    />
  </div>
</template>

<script>
import validator from "validator";

import Button from "@/components/Button";
import Spinner from "@/components/Spinner";

export default {
  components: {
    Button,
    Spinner
  },

  data() {
    return {
      domain: "",
      processing: false
    };
  },

  mounted() {
    this.$refs.input.focus(); // Focus the domain input on page entry
  },

  computed: {
    // urlDomain returns the domain name of the current page, if there is
    // one.
    urlDomain() {
      const [domain] = document.location.host.split(":"); // Strip the port
      if (validator.isFQDN(domain)) return domain;
    },

    // This checks whether the domain provided is actually valid.
    validDomain() {
      return validator.isFQDN(this.domain);
    }
  }
};
</script>

<style lang="scss" scoped>
.domain-group {
  margin-top: 30px;
  align-items: center;
  display: inline-flex;
}

input[name="domain"] {
  margin: 0 14px;
}

a {
  margin-top: 15px;
  display: block;
}

.button {
  margin-top: 30px;
}
</style>
