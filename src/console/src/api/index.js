import axios from "axios";

import router from "../router";
import store from "../store";

const instance = axios.create({
  baseURL: "/api",
  validateStatus: status => true, // Always succeed
  redirect: true // Custom redirect handler option
});

/**
 * Custom redirect handler. If certain requests fail with authentication errors,
 * we want to be able to clear out the local user store and then redirect to the
 * login page with an error.
 */
instance.interceptors.response.use(
  res => {
    if (!res.config.redirect) return res;

    // The request was unauthorized, perform the redirect.
    if (res.status === 401 || res.status === 403) {
      store.commit("clearUser");
      router.push("/login");
      return Promise.reject("User is not authorized");
    }

    return res;
  },

  err => {
    console.log(err.message);
    return {};
  }
);

export default instance;
