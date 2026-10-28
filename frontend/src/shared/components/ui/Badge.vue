<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{
    variant?: "ok" | "error" | "outline";
    dot?: boolean;
    class?: HTMLAttributes["class"];
  }>(),
  { variant: "outline" },
);

const variants = {
  ok: "bg-ok/15 text-ok",
  error: "bg-error/15 text-error-text",
  outline: "border border-default text-secondary",
} as const;
</script>

<template>
  <span
    :class="
      cn(
        'inline-flex h-6 items-center gap-1.5 rounded-full px-2.5 text-xs font-medium whitespace-nowrap',
        variants[props.variant],
        props.class,
      )
    "
  >
    <span
      v-if="props.dot"
      aria-hidden="true"
      class="size-2 rounded-full bg-current"
    />
    <slot />
  </span>
</template>
