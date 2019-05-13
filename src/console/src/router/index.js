import Vue from "vue";
import Router from "vue-router";
import Meta from "vue-meta";

Vue.use(Router);
Vue.use(Meta, { keyName: "meta" });

import MainView from "@/views/Main";
import SetupView from "@/views/Setup";
import LoginView from "@/views/Login";

const routes = [
  { path: "/setup", component: SetupView },
  { path: "/login", component: LoginView },
  { path: "/", component: MainView }
];

const mode = "history";
const router = new Router({ routes, mode });
export default router;
