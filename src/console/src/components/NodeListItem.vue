<template>
  <div class="item" @click="$push(`/nodes/${node.id}`)">
    <div class="status green"></div>
    <div class="address">{{ node.address }}</div>

    <NodeRoles :roles="node.node_roles" />
  </div>
</template>

<script>
import NodeRoles from "@/components/NodeRoles";

export default {
  components: { NodeRoles },
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

.address {
  font-weight: bold;
  margin-right: 10px;
}
</style>
