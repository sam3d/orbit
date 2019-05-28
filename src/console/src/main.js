import Vue from "vue";
import App from "./App";
import router from "./router";
import store from "./store";
import api from "./api";
import LoadingList from "./components/LoadingList";

import "reset-css";
import "@/styles/main.scss";

// Bind the API to the vue instance.
Vue.use(Vue => (Vue.prototype.$api = api));

// Helper to retrieve the current namespace ID.
Vue.use(Vue => {
  Vue.prototype.$namespace = () => {
    const id = store.state.namespace;
    if (!id || id === "default") return "";
    else return id;
  };
});

// Set the loading list view to be a global component.
Vue.component("LoadingList", LoadingList);

const vue = new Vue({
  store,
  router,
  render: h => h(App)
});

(async () => {
  await store.dispatch("init");
  await window.waitForLoaderTimeout();
  vue.$mount("#console");
  window.removeLoader();
})();
