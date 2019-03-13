<template>
  <section class="setup">
    <nav>
      <img class="logo" src="@/assets/logo/gradient-thick.svg" />
      <div class="logo-text">Orbit</div>
    </nav>

    <div class="body">
      <div class="inner">
        <template v-if="page === 0">
          <h1 class="large">Welcome to Orbit</h1>
          <p class="large">
            This node still needs to be configured before it is able to join or
            start a cluster. Please configure it as soon as possible, as this
            page is publicly accessible over the web to anybody with the link.
          </p>
          <div class="button purple">Begin set up</div>
        </template>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  meta: { title: "Setup" },

  data() {
    return {
      page: 0
    };
  }
};
</script>

<style lang="scss">
section.setup {
  width: 100%;
  height: 100%;
  left: 0;
  top: 0;
  position: fixed;

  display: flex;
  flex-direction: column;

  nav {
    display: flex;
    align-items: center;
    padding: 20px;
    flex-shrink: 0;

    .logo {
      height: 40px;
    }

    .logo-text {
      margin-left: 10px;
      font-family: "Cabin", sans-serif;
      font-size: 24px;
    }
  }

  .body {
    text-align: center;
    flex-grow: 1;
    padding: 20px;

    display: flex;
    text-align: center;
    overflow-y: scroll;

    .inner {
      margin: auto;
    }
  }

  h1.large {
    font-size: 45px;
  }

  p.large {
    max-width: 600px;
    margin: 50px 0;
    font-size: 18px;
    line-height: 2rem;
  }
}
</style>
