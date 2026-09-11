<script setup lang="ts">
import { cn } from "@/shared/utils/cn";

interface ChipOption {
  value: string;
  label: string;
  tone?: "error" | "ok" | "warn" | "info";
}

defineProps<{
  options: readonly ChipOption[];
  groupLabel: string;
}>();

const modelValue = defineModel<string>({ required: true });
</script>

<template>
  <div
    class="border-default bg-surface flex gap-0.5 rounded-lg border p-0.5"
    role="group"
    :aria-label="groupLabel"
  >
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      :aria-pressed="modelValue === option.value"
      :class="
        cn(
          'text-secondary hover:text-primary flex min-h-8 cursor-pointer items-center gap-1.5 rounded-md border-none bg-transparent px-3 py-1 text-[0.8125rem] font-medium whitespace-nowrap transition-colors duration-120 ease-out',
          modelValue === option.value && 'bg-active text-ink hover:bg-active hover:text-ink',
        )
      "
      @click="modelValue = option.value"
    >
      <span
        v-if="option.tone !== undefined"
        :class="
          cn(
            'size-1.5 shrink-0 rounded-full',
            option.tone === 'error' && 'bg-error',
            option.tone === 'ok' && 'bg-ok',
            option.tone === 'warn' && 'bg-warn',
            option.tone === 'info' && 'bg-info',
          )
        "
        aria-hidden="true"
      />
      {{ option.label }}
    </button>
  </div>
</template>
