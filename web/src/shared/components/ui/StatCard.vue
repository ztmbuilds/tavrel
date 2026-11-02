<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{
    label: string;
    value: string;
    unit?: string;
    tone?: "default" | "error";
    class?: HTMLAttributes["class"];
  }>(),
  { tone: "default" },
);
</script>

<template>
  <div
    :class="
      cn('border-subtle bg-panel rounded-xl border px-5 py-4', props.class)
    "
  >
    <div class="text-secondary text-[0.8125rem] leading-relaxed">
      {{ props.label }}
    </div>
    <div
      :class="
        cn(
          'mt-2 text-[2.125rem] leading-[1.1] font-semibold tracking-[-0.02em] tabular-nums',
          props.tone === 'error' ? 'text-error' : 'text-heading',
        )
      "
    >
      {{ props.value
      }}<span
        v-if="props.unit"
        class="text-tertiary ml-1 text-sm font-medium"
        >{{ props.unit }}</span
      >
    </div>
  </div>
</template>
