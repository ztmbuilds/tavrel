<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

import arrowRightIcon from "@/assets/icons/arrow-right.svg?component";
import Button from "@/shared/components/ui/Button.vue";
import PreloadLink from "@/shared/components/PreloadLink.vue";

const props = defineProps<{
  to: RouteLocationRaw;
  routeName: string;
  class?: string;
}>();
</script>

<template>
  <Button as="a" as-child variant="ghost" size="sm" :class="props.class">
    <PreloadLink
      :to="props.to"
      :route-name="props.routeName"
      class="group/link inline-flex items-center gap-1"
    >
      <span><slot /></span>
      <component
        :is="arrowRightIcon"
        aria-hidden="true"
        class="size-3.5 shrink-0 transition-transform duration-150 group-hover/link:translate-x-0.5 motion-reduce:transition-none"
      />
    </PreloadLink>
  </Button>
</template>
