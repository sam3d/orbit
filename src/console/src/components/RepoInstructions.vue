<template>
  <div class="instructions">
    <label>How to upload your code</label>
    <code class="light">{{ path }}</code>

    <code>
      <p class="comment">
        # Please run the following commands to enable git deployment.
      </p>
      git init<br />
      echo "# {{ name || "{name}" }}" > README.md<br />
      git commit -m "Initial commit"<br />
      git remote add deploy {{ path }}<br />
      git push -u deploy master
    </code>
  </div>
</template>

<script>
export default {
  props: {
    name: { type: String }
  },

  computed: {
    path() {
      const { protocol, host } = window.location;
      const main = `${protocol}//${host}/api/repo`;
      const name = this.paddedName;
      const namespace = this.$store.state.namespaceName;

      if (namespace) return `${main}/${namespace}/${name}`;
      else return `${main}/${name}`;
    },

    paddedName() {
      const name = this.name || "{name}";
      return name;
    }
  }
};
</script>

<style lang="scss" scoped>
.instructions {
  display: flex;
  flex-direction: column;
  align-items: center;

  margin: 30px 0;

  code {
    width: 100%;
    max-width: 480px;
  }

  code:not(.light) {
    font-size: 14px;
    margin-top: 10px;
    max-width: 530px;
  }
}
</style>
