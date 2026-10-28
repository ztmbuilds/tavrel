<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { Primitive } from "reka-ui";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{
    as?: string;
    asChild?: boolean;
    class?: HTMLAttributes["class"];
  }>(),
  { as: "div" },
);
</script>

<template>
  <Primitive
    :as="props.as"
    :as-child="props.asChild"
    :class="
      cn(
        'border-subtle flex items-center gap-3.5 border-t px-5 py-3 text-inherit no-underline hover:bg-white/[0.03]',
        props.class,
      )
    "
  >
    <slot />
  </Primitive>
</template>
