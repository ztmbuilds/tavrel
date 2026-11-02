import type { RouteRecordRaw } from "vue-router";

export const routes: readonly RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/app/layouts/AppLayout.vue"),
    children: [
      {
        path: "",
        name: "dashboard",
        component: () => import("@/features/dashboard/views/DashboardView.vue"),
      },
      {
        path: "traces",
        name: "traces",
        component: () => import("@/features/traces/views/TracesView.vue"),
        meta: {
          preload: {
            prefetchQueries: () =>
              import("@/shared/queries/filter-options").then(
                (m) => m.prefetchFilterOptions,
              ),
          },
        },
      },
      {
        path: "traces/:id",
        name: "trace-detail",
        component: () => import("@/features/traces/views/TraceDetailView.vue"),
      },
      {
        path: "spans",
        name: "spans",
        component: () => import("@/features/spans/views/SpansView.vue"),
        meta: {
          preload: {
            prefetchQueries: () =>
              import("@/shared/queries/filter-options").then(
                (m) => m.prefetchFilterOptions,
              ),
          },
        },
      },
      {
        path: "logs",
        name: "logs",
        component: () => import("@/features/logs/views/LogsView.vue"),
        meta: {
          preload: {
            prefetchQueries: () =>
              import("@/shared/queries/filter-options").then(
                (m) => m.prefetchFilterOptions,
              ),
          },
        },
      },
      {
        path: "services",
        name: "services",
        component: () => import("@/features/services/views/ServicesView.vue"),
      },
    ],
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/",
  },
];
