<script setup lang="ts">
import { onMounted } from "vue";
import { RouterView, useRoute } from "vue-router";

import { useRoutePreload } from "@/shared/composables/useRoutePreload";

interface NetworkInformation {
  saveData?: boolean;
  effectiveType?: string;
}

const IDLE_FALLBACK_DELAY_MS = 200;

const route = useRoute();
const { preloadRoute } = useRoutePreload();

onMounted(() => {
  const connection = (navigator as Navigator & { connection?: NetworkInformation })
    .connection;
  const isSlowConnection = Boolean(
    connection?.saveData || connection?.effectiveType?.includes("2g"),
  );
  if (isSlowConnection) return;

  if (route.name !== "dashboard") return;
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
