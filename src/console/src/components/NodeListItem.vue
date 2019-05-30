<template>
  <div class="item" @click="$push(`/nodes/${node.id}`)">
    <div class="status green"></div>
    <div class="address">{{ node.address }}</div>

    <div class="role yellow" v-if="node.node_roles.includes('LOAD_BALANCER')">
      Load Balancer
    </div>
    <div class="role blue" v-if="node.node_roles.includes('STORAGE')">
      Storage
    </div>
    <div class="role pink" v-if="node.node_roles.includes('BUILDER')">
      Builder
    </div>
  </div>
</template>

<script>
export default {
  props: ["node"]
};
</script>

<style lang="scss" scoped>
.status {
  width: 8px;
  height: 8px;
  border-radius: 8px;
  background-color: #aaa;
  margin-right: 10px;

  &.green {
    background-color: #1dd1a1;
  }
}

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

.address {
  font-weight: bold;
  margin-right: 10px;
}
</style>
