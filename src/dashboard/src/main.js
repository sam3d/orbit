import Vue from "vue";
import App from "./App";
import router from "./router";
import store from "./store";

const vue = new Vue({
  store,
  router,
  render: h => h(App)
});

vue.$mount("#dashboard");
