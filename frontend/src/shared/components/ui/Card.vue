<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { computed, useSlots } from "vue";
import { Primitive } from "reka-ui";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{
    title?: string;
    as?: string;
    flush?: boolean;
    class?: HTMLAttributes["class"];
  }>(),
  { as: "section" },
);

const slots = useSlots();
const hasHeader = computed(() => !!props.title || !!slots.action);
</script>

<template>
  <Primitive
    :as="props.as"
    :class="
      cn(
        'border-subtle bg-panel overflow-hidden rounded-xl border',
        props.class,
      )
    "
  >
    <header
      v-if="hasHeader"
      class="flex items-center justify-between gap-3 px-5 pt-4 pb-3.5"
    >
      <h2 class="text-heading m-0 text-sm font-semibold">{{ props.title }}</h2>
      <slot name="action" />
    </header>
    <div :class="cn('px-5 pb-5', !hasHeader && 'pt-5', props.flush && 'p-0')">
      <slot />
    </div>
  </Primitive>
</template>
