<script setup lang="ts">
import { RouterView } from "vue-router";
import { watch } from "vue";

import AppSidebar from "@/shared/components/AppSidebar.vue";
import { useServicesQuery } from "@/shared/queries/filter-options";
import { setServiceColorRegistry } from "@/shared/utils/service-color";

const { data: services } = useServicesQuery();

watch(
  services,
  (names) => {
    if (names !== undefined) setServiceColorRegistry(names);
  },
  { immediate: true },
);
</script>

<template>
  <div class="flex h-screen overflow-hidden">
    <a class="skip-link" href="#main">Skip to content</a>
    <AppSidebar />
    <main
      id="main"
      tabindex="-1"
      class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
    >
      <RouterView />
    </main>
  </div>
</template>
