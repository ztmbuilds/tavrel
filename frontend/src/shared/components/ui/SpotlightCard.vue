<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { cn } from "@/shared/utils/cn";

const props = defineProps<{
  eyebrow?: string;
  title: string;
  class?: HTMLAttributes["class"];
}>();
</script>

<template>
  <div :class="cn('relative pt-1.5', props.class)">
    <div
      aria-hidden="true"
      class="bg-spot-from/40 absolute inset-x-3.5 top-0 h-4.5 rounded-t-xl"
    />
    <section
      class="from-spot-from to-spot-to relative flex h-full flex-col gap-1 rounded-xl bg-linear-[165deg] p-5 text-white"
    >
      <div v-if="props.eyebrow" class="text-xs text-white/85">
        {{ props.eyebrow }}
      </div>
      <h2 class="m-0 text-[17px] font-semibold tracking-tight">
        {{ props.title }}
      </h2>
      <p class="m-0 text-[13px] text-white/90"><slot /></p>
      <div v-if="$slots.action" class="mt-auto pt-3">
        <slot name="action" />
      </div>
    </section>
  </div>
</template>
