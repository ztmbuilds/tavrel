<script setup lang="ts">
import { computed } from "vue";
import type { RouteLocationRaw } from "vue-router";

import arrowRightIcon from "@/assets/icons/arrow-right.svg?component";
import PreloadLink from "@/shared/components/PreloadLink.vue";
import { cn } from "@/shared/utils/cn";

const props = defineProps<{
  to: RouteLocationRaw;
  routeName: string;
  class?: string;
}>();

const linkClass = computed(() =>
  cn(
    "text-signal inline-flex cursor-pointer items-center gap-1 text-xs transition-colors duration-120 ease-out",
    props.class,
  ),
);
</script>

<template>
  <PreloadLink
    :to="props.to"
    :route-name="props.routeName"
    :class="linkClass"
    data-arrow-link
  >
    <span data-arrow-link-label class="pb-0.5"><slot /></span>
    <component
      :is="arrowRightIcon"
      aria-hidden="true"
      data-arrow-link-icon
      class="size-3 shrink-0"
    />
  </PreloadLink>
</template>

<style scoped>
[data-arrow-link-label] {
  background-image: linear-gradient(currentColor, currentColor);
  background-repeat: no-repeat;
  background-position: 0 100%;
  background-size: 0 1px;
}

[data-arrow-link]:hover [data-arrow-link-label],
[data-arrow-link]:focus-visible [data-arrow-link-label] {
  background-size: 100% 1px;
}

@media (prefers-reduced-motion: no-preference) {
  [data-arrow-link-label] {
    transition: background-size 250ms ease;
  }

  [data-arrow-link-icon] {
    transition: transform 200ms ease-out;
  }

  [data-arrow-link]:hover [data-arrow-link-icon],
  [data-arrow-link]:focus-visible [data-arrow-link-icon] {
    transform: translateX(4px);
  }
}
</style>
