<script setup lang="ts">
import { cn } from "@/shared/utils/cn";

const modelValue = defineModel<boolean>({ required: true });
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    aria-label="Live refresh"
    :class="
      cn(
        'text-secondary hover:text-primary flex min-h-9 shrink-0 cursor-pointer items-center gap-2 rounded-md border border-transparent bg-transparent px-2 font-sans text-[0.78125rem] whitespace-nowrap transition-colors duration-120 ease-out',
        modelValue && 'text-primary',
      )
    "
    @click="modelValue = !modelValue"
  >
    <span
      :class="cn('size-1.5 rounded-full', modelValue ? 'bg-ok' : 'bg-tertiary')"
      aria-hidden="true"
    />
    <span>Live</span>
    <span
      :class="
        cn(
          'flex h-4 w-7 shrink-0 items-center rounded-full border p-px transition-colors duration-150 ease-out',
          modelValue ? 'border-ok bg-ok/20' : 'border-strong bg-surface',
        )
      "
      aria-hidden="true"
    >
      <span
        :class="
          cn(
            'size-3 rounded-full transition-all duration-150 ease-out',
            modelValue ? 'bg-ok translate-x-3' : 'bg-tertiary translate-x-0',
          )
        "
      />
    </span>
  </button>
</template>
