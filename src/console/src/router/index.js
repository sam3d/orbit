import Vue from "vue";
import Router from "vue-router";
import Meta from "vue-meta";

Vue.use(Router);
Vue.use(Meta, { keyName: "meta" });

import SetupView from "@/views/Setup";
import LoginView from "@/views/Login";

// All of the primary views.
import MainView from "@/views/Main";
import OverviewView from "@/views/Overview";
import NotFoundView from "@/views/NotFound";

const routes = [
  { path: "/setup", component: SetupView },
  { path: "/login", component: LoginView },
  {
    path: "/",
    component: MainView,
    children: [
      { path: "", component: OverviewView },
      { path: "*", component: NotFoundView }
    ]
  }
];

const mode = "history";
const router = new Router({ routes, mode });
export default router;
