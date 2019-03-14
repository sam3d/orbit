<template>
  <div>
    <h1>Welcome to Orbit</h1>

    <p class="large">
      This node still needs to be configured before it is able to join or start
      a cluster. Please configure it as soon as possible, as this page is
      publicly accessible over the web to anybody with the link.
    </p>

    <div class="button purple" @click="$emit('click')">
      Start setup
    </div>
  </div>
</template>
