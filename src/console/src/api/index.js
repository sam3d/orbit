import axios from "axios";

const api = axios.create({
  baseURL: process.env.VUE_APP_ORBIT_API_URL,
  validateStatus: status => true
});

export default api;
