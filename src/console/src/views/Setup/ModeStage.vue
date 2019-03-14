<template>
  <div>
    <h2>What would you like to do?</h2>
    <div class="options">
      <!-- Option for creating a new cluster -->
      <div
        class="option"
        :class="{ selected: mode === 'bootstrap' }"
        @click="mode = 'bootstrap'"
      >
        <img />
        <h3>Create a new cluster</h3>
        <p>
          Don't already have a cluster? Not a problem, click here to create a
          brand new single-node cluster.
        </p>
      </div>

      <!-- Option for joining an existing cluster -->
      <div
        class="option blue"
        :class="{ selected: mode === 'join' }"
        @click="mode = 'join'"
      >
        <img />
        <h3>Join an existing cluster</h3>
        <p>
          If you already have a cluster in operation, great! Click here and
          continue to get this node joined up to it.
        </p>
      </div>
    </div>

    <div
      class="button green"
      :class="{ disabled: mode === '', blue: mode === 'join' }"
      @click="$emit('click', mode)"
    >
      Continue
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      mode: ""
    };
  }
};
</script>

<style lang="scss" scoped>
.options {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  flex-wrap: wrap;

  margin: 40px 0;

  .option {
    cursor: pointer;

    max-width: 330px;
    padding: 30px;
    background-color: #fff;
    border-radius: 4px;
    margin: 20px;
    border: solid 1px #fff;

    transition: border-color 0.2s;

    img {
      width: 60px;
      height: 60px;
    }

    h3 {
      font-size: 20px;
      font-weight: 500;
      margin-top: 20px;

      transition: color 0.2s;
    }

    p {
      margin-top: 20px;
      line-height: 1.5rem;
    }

    &.selected {
      border: solid 1px #1dd1a1;
      h3 {
        color: #1dd1a1;
      }
    }

    &.blue {
      &.selected {
        border: solid 1px #0abde3;
        h3 {
          color: #0abde3;
        }
      }
    }
  }
}
</style>
