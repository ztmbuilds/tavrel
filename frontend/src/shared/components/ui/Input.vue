<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{ type?: string; class?: HTMLAttributes["class"] }>(),
  { type: "text" },
);
defineOptions({ inheritAttrs: false });

const model = defineModel<string>({ default: "" });
</script>

<template>
  <label
    :class="
      cn(
        'border-subtle bg-canvas text-secondary flex h-8 items-center gap-2.5 rounded-lg border px-3',
        'hover:border-default transition-[border-color,box-shadow] duration-150',
        'focus-within:border-signal/70 focus-within:ring-signal/15 focus-within:ring-3',
        props.class,
      )
    "
  >
    <span
      v-if="$slots.leading"
      aria-hidden="true"
      class="inline-flex [&_svg]:size-4 [&_svg]:fill-none [&_svg]:stroke-current [&_svg]:stroke-[1.6] [&_svg]:[stroke-linecap:round] [&_svg]:[stroke-linejoin:round]"
    >
      <slot name="leading" />
    </span>
    <input
      v-bind="$attrs"
      v-model="model"
      :type="props.type"
      class="text-primary placeholder:text-tertiary h-full min-w-0 flex-1 bg-transparent outline-none focus-visible:outline-none"
    />
    <slot name="trailing" />
  </label>
</template>
