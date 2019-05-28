<template>
  <div class="empty">
    <h3>You don't currently have any {{ subject }}</h3>
    <p class="description">{{ description }}</p>
    <div class="button">Hello</div>
  </div>
</template>

<script>
export default {
  props: {
    subject: { type: String },
    description: { type: String },
    action: { type: String }
  }
};
</script>

<style lang="scss" scoped>
.empty {
  padding: 50px;
  border-radius: 4px;
  border: solid 1px rgba(0, 0, 0, 0.1);
  text-align: center;
  display: flex;
  align-items: center;
  flex-direction: column;

  h3 {
    font-size: 20px;
    font-weight: bold;
  }

  p.description {
    font-size: 18px;
    margin-top: 20px;
    line-height: 1.6rem;
    max-width: 600px;
    color: rgba(0, 0, 0, 0.7);
  }
}
</style>
