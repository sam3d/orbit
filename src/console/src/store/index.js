import Vue from "vue";
import Vuex from "vuex";
Vue.use(Vuex);

import api from "@/api";
import router from "@/router";

const store = new Vuex.Store({
  state: {
    init: {
      status: null,
      status_string: null,
      stage: null,
      mode: null
    },
    user: {
      id: null,
      name: null,
      username: null,
      email: null
    },
    token: null,
    ip: null,

    namespace: null, // The currently selected namespace
    title: null // The current page title
  },

  mutations: {
    init(state, data) {
      state.init = data;
    },

    ip(state, ip) {
      state.ip = ip;
    },

    user(state, user) {
      state.user = user;
    },

    token(state, token) {
      state.token = token;
    },

    namespace(state, namespace) {
      state.namespace = namespace;
    },

    title(state, title) {
      state.title = title;
    },

    clearUser(state) {
      state.token = "";
      state.user = {
        id: null,
        name: null,
        username: null,
        email: null
      };
    }
  },

  actions: {
    /**
     * The init action is responsible for retrieving the overall cluster state
     * and then (in the majority of cases) populating it with the current user
     * information. During the setup phase, the user route won't get called as
     * there can't be any user information.
     */
    async init({ commit, dispatch }) {
      var res = await api.get("/state", { redirect: false });
      if (res.status !== 200) return alert(res.data);

      const path = window.location.pathname;
      const engineStatus = res.data.status_string;

      // Keep the data that we get back in the store.
      commit("init", res.data);

      // If we need to retrieve the IP address for the setup process, then do.
      if (engineStatus === "setup") {
        var res = await api.get("/ip", { redirect: false });
        if (res.status === 200) {
          commit("ip", res.data.ip);
        }
      }

      /**
       * If not already on the setup page and the engine status is setup, then
       * push it in that direction. This is done because if the user is already
       * on the setup page, the query parameter is lost as a result of the
       * "push".
       */
      if (path !== "/setup" && engineStatus !== "running")
        return router.push("/setup");
      if (path === "/setup" && engineStatus === "running")
        return router.push("/");

      // Check and update the user details.
      await dispatch("updateUser");
    },

    /**
     * Check the user login status. If the user is logged in then we can leave
     * them exactly where they are, otherwise we need to push them to the log in
     * screen. Also, if the token that they have is not valid, then revoke it
     * and redirect them to the login screen.
     */
    async updateUser({ commit }) {
      const token = localStorage.getItem("token");
      if (!token) return router.push("/login");

      // Check if the token is valid, and if it isn't, redirect to login and
      // remove the token.
      var res = await api.get(`/user/${token}`);
      if (res.status !== 200) {
        localStorage.removeItem("token");
        commit("clearUser"); // Clear the token and user information
        return router.push("/login");
      }

      // Update the store with the user information.
      commit("token", token);
      commit("user", res.data);
    }
  }
});

export default store;
