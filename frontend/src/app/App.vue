<script setup lang="ts">
import { onMounted } from "vue";
import { RouterView, useRoute } from "vue-router";

import { useRoutePreload } from "@/shared/composables/useRoutePreload";

// navigator.connection is Chromium-only (absent in Safari/Firefox)
interface NetworkInformation {
  effectiveType?: string;
}

const IDLE_FALLBACK_DELAY_MS = 200;

const route = useRoute();
const { preloadRoute } = useRoutePreload();

onMounted(() => {
  const connection = (
    navigator as Navigator & { connection?: NetworkInformation }
  ).connection;
  const isSlowConnection = Boolean(connection?.effectiveType?.includes("2g"));
  if (isSlowConnection) return;

  if (route.name !== "dashboard") return;
  // requestIdleCallback is unsupported in Safari
  const idle: (callback: () => void) => void =
    typeof window.requestIdleCallback === "function"
      ? (callback) => window.requestIdleCallback(callback)
      : (callback) => setTimeout(callback, IDLE_FALLBACK_DELAY_MS);

  idle(() => {
    void preloadRoute("spans");
  });
});
</script>

<template>
  <RouterView />
</template>
