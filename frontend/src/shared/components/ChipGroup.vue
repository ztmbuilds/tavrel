<script setup lang="ts">
import { cn } from "@/shared/utils/cn";

interface ChipOption {
  value: string;
  label: string;
}

defineProps<{
  options: readonly ChipOption[];
  groupLabel: string;
}>();

const modelValue = defineModel<string>({ required: true });
</script>

<template>
  <div
    class="border-default bg-surface flex gap-0.5 rounded-md border p-0.5"
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
          'text-secondary hover:text-primary min-h-6 cursor-pointer rounded border-none bg-transparent px-2.5 py-1 text-xs font-medium whitespace-nowrap transition-colors duration-120 ease-out',
          modelValue === option.value && 'bg-active text-ink hover:bg-active hover:text-ink',
        )
      "
      @click="modelValue = option.value"
    >
      {{ option.label }}
    </button>
  </div>
</template>
