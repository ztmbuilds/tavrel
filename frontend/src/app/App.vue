<script setup lang="ts">
import { onMounted } from "vue";
import { RouterView, useRoute } from "vue-router";

import { useRoutePreload } from "@/shared/composables/useRoutePreload";

interface NetworkInformation {
  saveData?: boolean;
  effectiveType?: string;
}

const IDLE_FALLBACK_DELAY_MS = 200;

const nextRouteMap: Record<string, string[]> = {
  dashboard: ["spans"],
};

const route = useRoute();
const { preloadRoute } = useRoutePreload();

onMounted(() => {
  const connection = (navigator as Navigator & { connection?: NetworkInformation })
    .connection;
  const isSlowConnection = Boolean(
    connection?.saveData || connection?.effectiveType?.includes("2g"),
  );
  if (isSlowConnection) return;

  const targets = nextRouteMap[String(route.name)] ?? [];
  const idle: (callback: () => void) => void =
    typeof window.requestIdleCallback === "function"
      ? (callback) => window.requestIdleCallback(callback)
      : (callback) => setTimeout(callback, IDLE_FALLBACK_DELAY_MS);
  idle(() => {
    targets.forEach((routeName) => void preloadRoute(routeName));
  });
});
</script>

<template>
  <RouterView />
</template>
