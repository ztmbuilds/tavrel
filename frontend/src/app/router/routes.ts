import type { RouteRecordRaw } from "vue-router";

// Single route table (deviation from VUE_STYLE_GUIDE.md §1: six routes fit
// comfortably in one file; split per-feature only if this grows past ~150 lines).
export const routes: readonly RouteRecordRaw[] = [
  {
    path: "/",
    name: "dashboard",
    component: () => import("@/features/dashboard/views/DashboardView.vue"),
  },
  {
    path: "/traces",
    name: "traces",
    component: () => import("@/features/traces/views/TracesView.vue"),
  },
  {
    path: "/traces/:id",
    name: "trace-detail",
    component: () => import("@/features/traces/views/TraceDetailView.vue"),
  },
  {
    path: "/spans",
    name: "spans",
    component: () => import("@/features/spans/views/SpansView.vue"),
  },
  {
    path: "/logs",
    name: "logs",
    component: () => import("@/features/logs/views/LogsView.vue"),
  },
  {
    path: "/services",
    name: "services",
    component: () => import("@/features/services/views/ServicesView.vue"),
  },
];
