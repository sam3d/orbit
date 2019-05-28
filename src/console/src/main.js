import Vue from "vue";
import App from "./App";
import router from "./router";
import store from "./store";
import api from "./api";

import LoadingList from "./components/LoadingList";
import Empty from "./components/Empty";
import Button from "./components/Button";

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

// Helper to navigate maintaining query parameters.
Vue.use(
  Vue =>
    (Vue.prototype.$push = path =>
      router.push({ path, query: router.currentRoute.query }))
);

// Global components.
Vue.component("LoadingList", LoadingList);
Vue.component("Empty", Empty);
Vue.component("Button", Button);

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
