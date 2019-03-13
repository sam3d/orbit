import Vue from "vue";
import Vuex from "vuex";
Vue.use(Vuex);

import api from "@/api";
import router from "@/router";

const store = new Vuex.Store({
  actions: {
    /**
     * The init action is responsible for retrieving the overall cluster state
     * and then (in the majority of cases) populating it with the current user
     * information. During the setup phase, the user route won't get called as
     * there can't be any user information.
     */
    async init() {
      const res = await api.get("/state", { redirect: false });
      if (res.status !== 200) return;

      switch (res.data.status_string) {
        case "setup":
          router.push("/setup");
          return;
        case "ready":
        case "running":
          console.log("Show the standard login process");
          break;
      }
    }
  }
});
export default store;
