<template>
  <div class="flex-center">
    <h2>Address</h2>
    <p class="subheader">
      The is the IP address that can be used to reach this node from the
      internet. This can't be changed after the node is set up, so make sure
      this node is in the network environment it will remain in.
    </p>

    <input
      type="text"
      placeholder="192.168.56.11"
      v-model="address"
      autofocus
    />

    <div class="button" :class="{ disabled: !address }">Continue</div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      address: ""
    };
  }
};
</script>

<style lang="scss" scoped>
input {
  margin: 34px 0;
}
</style>
