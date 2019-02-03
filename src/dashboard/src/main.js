import Vue from "vue";
import App from "./App";
import router from "./router";
import store from "./store";

import "reset-css";
import "@/styles/main.scss";

const vue = new Vue({
  store,
  router,
  render: h => h(App)
});

(async () => {
  await window.waitForLoaderTimeout();
  vue.$mount("#dashboard");
  window.removeLoader();
})();
