<template>
  <div class="spinner"></div>
</template>

<style lang="scss" scoped>
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 3px solid transparent;
  border-radius: 50%;
  border-top-color: #02101c;
  border-bottom-color: #02101c;
  animation: spin 1s linear infinite;
}
</style>
