import Vue from "vue";
import Vuex from "vuex";
Vue.use(Vuex);

import api from "@/api";

const store = new Vuex.Store({
  actions: {
    /**
     * The "init" user action is responsible for setting up the entire state of
     * the store. That means that if the "user" route from the Orbit API server
     * tells us that there are no user and the system requires a setup, then
     * there can't be any user information and we simply redirect to that.
     */
    async init() {
      const res = await api.get("/user");
      console.log(res);
    }
  }
});
export default store;
