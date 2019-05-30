<template>
  <div class="sidebar-screen">
    <Spinner />
  </div>
</template>

<script>
import Spinner from "./Spinner";

export default {
  components: { Spinner }
};
</script>

<style lang="scss" scoped>
.sidebar-screen {
  width: 300px;
  height: 100%;
  justify-content: center;
}

.spinner {
  opacity: 0.6;
}
</style>
