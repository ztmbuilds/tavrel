import { createRouter, createWebHistory } from "vue-router";

import { routes } from "./routes";

const router = createRouter({
  // Clean URLs; Go serves index.html fallback for unknown paths (Phase 2).
  history: createWebHistory(),
  routes: [...routes],
});

export default router;
