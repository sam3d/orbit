<template>
  <div class="roles">
    <div class="role yellow" v-if="hasRole('LOAD_BALANCER')">
      Load Balancer
    </div>
    <div class="role blue" v-if="hasRole('STORAGE')">
      Storage
    </div>
    <div class="role pink" v-if="hasRole('BUILDER')">
      Builder
    </div>
  </div>
</template>

<script>
export default {
  props: {
    roles: { type: Array, default: [] }
  },

  methods: {
    hasRole(role) {
      return this.roles.includes(role);
    }
  }
};
</script>

<style lang="scss" scoped>
.roles {
  display: flex;
  align-items: center;

  .role {
    margin-left: 5px;
    font-size: 14px;
    padding: 5px 10px;
    border-radius: 4px;
    font-weight: bold;
    white-space: nowrap;

    &.yellow {
      background-color: transparentize(#ff9f43, 0.9);
      color: #ff9f43;
    }

    &.pink {
      background-color: transparentize(#f368e0, 0.9);
      color: #f368e0;
    }

    &.blue {
      background-color: transparentize(#54a0ff, 0.9);
      color: #54a0ff;
    }
  }
}
</style>
