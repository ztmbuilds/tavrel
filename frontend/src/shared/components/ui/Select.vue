<script setup lang="ts">
import {
  SelectContent,
  SelectItem,
  SelectItemText,
  SelectPortal,
  SelectRoot,
  SelectTrigger,
  SelectValue,
  SelectViewport,
} from "reka-ui";

import chevronDownIcon from "@/assets/icons/chevron-down.svg?component";
import { cn } from "@/shared/utils/cn";

export interface SelectOption {
  value: string;
  label: string;
}

const props = withDefaults(
  defineProps<{
    label: string;
    options: readonly SelectOption[];
    placeholder?: string;
    triggerClass?: string;
  }>(),
  {
    placeholder: "Select an option",
  },
);

const modelValue = defineModel<string>({ required: true });

const CONTENT_SIDE_OFFSET = 6;
</script>

<template>
  <SelectRoot v-model="modelValue">
    <SelectTrigger
      :aria-label="props.label"
      :class="
        cn(
          'border-default bg-surface text-secondary hover:border-strong flex min-h-9 cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 font-sans text-[0.8125rem] whitespace-nowrap outline-none transition-colors duration-120 ease-out data-[state=open]:border-signal',
          props.triggerClass,
        )
      "
    >
      <span aria-hidden="true">{{ props.label }}</span>
      <SelectValue
        :placeholder="props.placeholder"
        class="flex-1 truncate text-left text-primary"
      />
      <component :is="chevronDownIcon" class="text-tertiary size-2.5 shrink-0" />
    </SelectTrigger>
    <SelectPortal>
      <SelectContent
        position="popper"
        :side-offset="CONTENT_SIDE_OFFSET"
        class="z-50 min-w-[var(--reka-select-trigger-width)] overflow-hidden rounded-lg border border-default bg-raised p-1.5 shadow-[0_0.25rem_0.75rem_rgba(0,0,0,0.4)]"
      >
        <SelectViewport class="max-h-60 overflow-y-auto">
          <SelectItem
            v-for="option in props.options"
            :key="option.value"
            :value="option.value"
            class="flex cursor-pointer items-center rounded-md px-2 py-1.5 font-sans text-[0.8125rem] text-secondary outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-40 data-[highlighted]:bg-hover data-[highlighted]:text-primary data-[state=checked]:bg-active data-[state=checked]:text-ink data-[highlighted]:data-[state=checked]:bg-active data-[highlighted]:data-[state=checked]:text-ink"
          >
            <SelectItemText>{{ option.label }}</SelectItemText>
          </SelectItem>
        </SelectViewport>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>
