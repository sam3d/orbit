import Vue from "vue";
import Vuex from "vuex";
Vue.use(Vuex);

import api from "@/api";

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
    }
  }
});
export default store;
