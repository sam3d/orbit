<template>
  <div>
    <h1>Page not found</h1>
    <p>Sorry, this page you are visiting does not exist.</p>
    <router-link to="/">&larr; Go to the overview</router-link>
  </div>
</template>

<script>
export default {
  meta: { title: "Not found" }
};
</script>

<style lang="scss" scoped>
p {
  font-size: 20px;
}

a {
  margin-top: 30px;
  display: inline-block;
  font-size: 20px;
}
</style>
