<script setup lang="ts">
import { RouterLink, type RouteLocationRaw } from "vue-router";

import { useRoutePreload } from "@/shared/composables/useRoutePreload";

const props = defineProps<{ to: RouteLocationRaw; routeName: string }>();

const { preloadRoute } = useRoutePreload();

function handleIntent(): void {
  void preloadRoute(props.routeName);
}
</script>

<template>
  <RouterLink :to="to" @mouseenter="handleIntent" @focus="handleIntent">
    <slot />
  </RouterLink>
</template>
