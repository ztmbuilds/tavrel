<script setup lang="ts">
import { cn } from "@/shared/utils/cn";
import Switch from "@/shared/components/ui/Switch.vue";

const modelValue = defineModel<boolean>({ required: true });
</script>

<template>
  <label
    for="live-refresh"
    :class="
      cn(
        'text-secondary flex h-8 shrink-0 cursor-pointer items-center gap-2 text-[0.78125rem] whitespace-nowrap select-none',
        modelValue && 'text-heading',
      )
    "
  >
    <span
      :class="
        cn(
          'size-1.5 rounded-full',
          modelValue ? 'bg-ok animate-pulse' : 'bg-tertiary',
        )
      "
      aria-hidden="true"
    />
    <span>Live</span>
    <Switch id="live-refresh" v-model="modelValue" aria-label="Live refresh" />
  </label>
</template>
