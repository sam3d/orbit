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
import SecurityView from "@/views/Security";

import NodesView from "@/views/Nodes";
import NodeView from "@/views/Node";

import NamespaceView from "@/views/Namespace";
import NamespacesView from "@/views/Namespaces";
import NewNamespaceView from "@/views/NewNamespace";

import UsersView from "@/views/Users";
import UserView from "@/views/User";
import NewUserView from "@/views/NewUser";

import RepositoriesView from "@/views/Repositories";
import RepositoryView from "@/views/Repository";
import NewRepositoryView from "@/views/NewRepository";

import DeploymentsView from "@/views/Deployments";
import DeploymentView from "@/views/Deployment";
import NewDeploymentView from "@/views/NewDeployment";

import RoutersView from "@/views/Routers";
import RouterView from "@/views/Router";
import NewRouterView from "@/views/NewRouter";

import CertificatesView from "@/views/Certificates";
import CertificateView from "@/views/Certificate";
import NewCertificateView from "@/views/NewCertificate";

import VolumesView from "@/views/Volumes";
import VolumeView from "@/views/Volume";
import NewVolumeView from "@/views/NewVolume";

import NotFoundView from "@/views/NotFound";

const routes = [
  { path: "/setup", component: SetupView },
  { path: "/login", component: LoginView },
  {
    path: "/",
    component: MainView,
    children: [
      {
        path: "/nodes",
        component: NodesView
      },
      {
        path: "/nodes/:id",
        components: { default: NodesView, slider: NodeView }
      },

      {
        path: "/namespaces",
        component: NamespacesView
      },
      {
        path: "/namespaces/new",
        components: { default: NamespacesView, slider: NewNamespaceView }
      },
      {
        path: "/namespaces/:id",
        components: { default: NamespacesView, slider: NamespaceView }
      },

      {
        path: "/users",
        component: UsersView
      },
      {
        path: "/users/new",
        components: { default: UsersView, slider: NewUserView }
      },
      {
        path: "/users/:id",
        components: { default: UsersView, slider: UserView }
      },

      { path: "/security", component: SecurityView },

      { path: "", component: OverviewView },

      {
        path: "/repositories",
        component: RepositoriesView
      },
      {
        path: "/repositories/new",
        components: { default: RepositoriesView, slider: NewRepositoryView }
      },
      {
        path: "/repositories/:id",
        components: { default: RepositoriesView, slider: RepositoryView }
      },

      {
        path: "/deployments",
        component: DeploymentsView
      },
      {
        path: "/deployments/new",
        components: { default: DeploymentsView, slider: NewDeploymentView }
      },
      {
        path: "/deployments/:id",
        components: { default: DeploymentsView, slider: DeploymentView }
      },

      {
        path: "/routers",
        component: RoutersView
      },
      {
        path: "/routers/new",
        components: { default: RoutersView, slider: NewRouterView }
      },
      {
        path: "/routers/:id",
        components: { default: RoutersView, slider: RouterView }
      },

      {
        path: "/certificates",
        component: CertificatesView
      },
      {
        path: "/certificates/new",
        components: { default: CertificatesView, slider: NewCertificateView }
      },
      {
        path: "/certificates/:id",
        components: { default: CertificatesView, slider: CertificateView }
      },

      {
        path: "/volumes",
        component: VolumesView
      },
      {
        path: "/volumes/new",
        components: { default: VolumesView, slider: NewVolumeView }
      },
      {
        path: "/volumes/:id",
        components: { default: VolumesView, slider: VolumeView }
      },

      { path: "*", component: NotFoundView }
    ]
  }
];

const mode = "history";
const router = new Router({ routes, mode });
export default router;
