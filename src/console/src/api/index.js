import axios from "axios";

const instance = axios.create({
  baseURL: process.env.VUE_APP_ORBIT_API_URL,
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
    return res;
  },

  err => {
    console.log(err.message);
    return {};
  }
);

export default instance;
