import Vue from "vue";
import Vuex from "vuex";
Vue.use(Vuex);

import api from "@/api";
import router from "@/router";

const store = new Vuex.Store({
  state: {
    ip: null
  },

  mutations: {
    ip(state, ip) {
      state.ip = ip;
    }
  },

  actions: {
    /**
     * The init action is responsible for retrieving the overall cluster state
     * and then (in the majority of cases) populating it with the current user
     * information. During the setup phase, the user route won't get called as
     * there can't be any user information.
     */
    async init({ commit }) {
      const res = await api.get("/state", { redirect: false });
      if (res.status !== 200) return;

      // Keep the public IP that we get back.
      commit("ip", res.data.public_ip);

      const path = window.location.pathname;
      const engineStatus = res.data.status_string;

      /**
       * If not already on the setup page and the engine status is setup, then
       * push it in that direction. This is done because if the user is already
       * on the setup page, the query parameter is lost as a result of the
       * "push".
       */
      if (path !== "/setup" && engineStatus === "setup")
        return router.push("/setup");
    }
  }
});
export default store;
