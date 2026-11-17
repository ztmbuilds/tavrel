import { createRouter, createWebHistory } from "vue-router";

import { authSessionQueryOptions } from "@/features/auth/queries";

import { queryClient } from "../query-client";
import { routes } from "./routes";

const router = createRouter({
  // Clean URLs; Go serves index.html fallback for unknown paths (Phase 2).
  history: createWebHistory(),
  routes: [...routes],
});

router.beforeEach(async (to) => {
  try {
    const session = await queryClient.query(authSessionQueryOptions());

    if (!session.enabled) {
      return to.name === "login" ? { name: "dashboard" } : true;
    }

    if (!session.authenticated) {
      if (to.name === "login") return true;
      return { name: "login", query: { redirect: to.fullPath } };
    }

    return to.name === "login" ? { name: "dashboard" } : true;
  } catch {
    // Let navigation proceed so page-level error states surface when the
    // session endpoint is unreachable.
    return true;
  }
});

export default router;
