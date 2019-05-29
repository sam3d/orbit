<template>
  <div class="browser">
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
.browser {
  text-align: left;
  border: solid 1px #ddd;
  border-radius: 4px;
  overflow: hidden;

  width: 100%;

  select {
    -webkit-appearance: none;
    -moz-appearance: none;
    -ms-appearance: none;
    -o-appearance: none;
    appearance: none;

    width: 100%;
    background: none;
    cursor: pointer;

    background-image: url("~@/assets/icon/dropdown.svg");
    background-size: 10px;
    background-position: center right 10px;
    background-repeat: no-repeat;

    font-family: "Montserrat", sans-serif;
    font-size: 14px;
    padding: 10px 20px;
    border: none;
    font-weight: bold;
    border-bottom: solid 1px #ddd;
    border-radius: 0;

    &:focus {
      outline: none;
    }
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
