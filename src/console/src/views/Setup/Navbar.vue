<template>
  <nav>
    <div class="logo">
      <img src="@/assets/logo/gradient.svg" />
      <span class="name">Orbit</span>
    </div>

    <ul>
      <li><a href="https://orbit.sh/docs">Read the docs</a></li>
      <li><a href="https://orbit.sh/support">Support</a></li>
    </ul>
  </nav>
</template>

<style lang="scss" scoped>
nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  flex-shrink: 0;

  .logo {
    cursor: default;

    display: flex;
    align-items: center;

    img {
      height: 50px;
    }

    span.name {
      margin-left: 15px;
      font-family: "Cabin", sans-serif;
      font-weight: 500;
      opacity: 0.8;
      font-size: 34px;
      margin-top: 3px; // The text is too high, so this centers it
    }
  }

  ul li {
    display: inline-block;
    &:not(:last-of-type) {
      margin-right: 24px;
    }
  }
}
</style>
