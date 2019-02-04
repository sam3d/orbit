import Vue from "vue";
import Router from "vue-router";

Vue.use(Router);

import MainView from "@/views/Main";

const routes = [{ path: "/", component: MainView }];

const mode = "history";
const router = new Router({ routes, mode });
export default router;
