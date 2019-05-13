<template>
  <div class="root">
    <div class="navbar">Navbar</div>
    <div class="container">
      <div class="sidebar">The sidebar</div>
      <div class="content">The content</div>
    </div>
    <div class="footer">The footer</div>
  </div>
</template>

<script>
export default {};
</script>

<style lang="scss" scoped>
$borderColor: darken(#f5f6fa, 5%);

.root {
  display: flex;
  flex-direction: column;
  position: absolute;
  width: 100vw;
  height: 100vh;
  left: 0;
  top: 0;

  .navbar {
    display: flex;
    background-color: #fff;
    border-bottom: solid 1px $borderColor;
    padding: 20px;
  }

  .container {
    display: flex;
    flex-grow: 2;
  }

  .sidebar {
    flex-shrink: 0;
    background-color: #fff;
    border-right: solid 1px $borderColor;
    padding: 20px;
  }

  .content {
    flex-grow: 2;
    flex-shrink: 0;
  }

  .footer {
    background-color: #fff;
    border-top: solid 1px $borderColor;
    display: flex;
    padding: 10px 20px;
  }
}
</style>
