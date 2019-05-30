<template>
  <div class="browser empty" v-if="branches.length == 0 || nodes.length == 0">
    This repository contains no files or branches! Please upload some by using
    the instructions below.
  </div>

  <div class="browser" v-else>
    <select v-model="currentBranch">
      <option v-for="branch in branches">{{ branch }}</option>
    </select>

    <TreeView v-model="nodes" />
  </div>
</template>

<script>
import TreeView from "sl-vue-tree";

export default {
  props: ["files"],
  components: { TreeView },

  data() {
    return {
      currentBranch: "",
      nodes: []
    };
  },

  mounted() {
    this.useDefaultBranch();
  },

  methods: {
    useDefaultBranch() {
      if (!this.branches.length) return;
      if (this.branches.includes("master")) this.currentBranch = "master";
      else this.currentBranch = this.branches[0];
    }
  },

  computed: {
    branches() {
      return Object.keys(this.files);
    }
  },

  watch: {
    currentBranch(branch) {
      this.nodes = tree(this.files[branch]);
    },

    branches() {
      this.useDefaultBranch();
    }
  }
};

function tree(paths) {
  const tree = {}; // Prepare the tree.

  // Split the tree into nested dictionaries.
  for (let path of paths) {
    let currentNode = tree;
    for (let segment of path.split("/")) {
      if (currentNode[segment] === undefined) {
        currentNode[segment] = {};
      }
      currentNode = currentNode[segment];
    }
  }

  return toTreeData(tree);
}

function toTreeData(tree) {
  return Object.keys(tree).map(title => {
    let o = {
      title,
      isLeaf: true,
      isDraggable: false,
      isExpanded: false
    };

    if (Object.keys(tree[title]).length > 0) {
      o.isLeaf = false;
      o.children = toTreeData(tree[title]);
    }
    return o;
  });
}
</script>

<style lang="scss">
.browser.empty {
  max-width: 500px;
  line-height: 1.6rem;
}

.browser:not(.empty) {
  text-align: left;
  border: solid 1px #ddd;
  border-radius: 4px;
  overflow: hidden;

  width: 100%;

  select {
    width: 100%;
    border: none;
    border-bottom: solid 1px #ddd;
    border-radius: 0;
  }

  .sl-vue-tree-nodes-list {
    display: grid;
    gap: 1px;
    background-color: #eee;
  }

  .sl-vue-tree-node {
    background-color: #fff;
    padding: 10px;
    cursor: default;

    &:hover {
      // background-color: #fafafa;
    }
  }
}
</style>
