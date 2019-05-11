<template>
  <div>
    <h2>All done!</h2>
    <template v-if="mode === 'bootstrap'">
      <p class="subheader">
        Congratulations, the set up for this cluster is now complete! You may
        now start using your cluster by continuing to the log in page by
        pressing the button below.
      </p>

      <Button
        class="purple"
        text="Log in to your cluster"
        @click="$router.push('/login')"
      />
    </template>
    <template v-else>
      <p class="subheader">
        Congratulations, the set up for this node is now complete! You may now
        continue using your cluster by pressing the link below to be redirected
        to the dashboard.
      </p>

      <Button
        class="purple"
        text="Return to cluster"
        @click="alert('need to retrieve the main orbit url from the api')"
      />
    </template>
  </div>
</template>

<script>
import Button from "@/components/Button";

export default {
  props: ["mode"],

  components: { Button }
};
</script>

<style lang="scss" scoped>
.button {
  margin-top: 30px;
}
</style>
