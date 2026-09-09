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

import chevronDownIcon from "@/assets/icons/chevron-down.svg";
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
          'flex min-h-7 cursor-pointer items-center gap-1.5 rounded-md border border-default bg-surface px-2.5 py-1 font-sans text-xs whitespace-nowrap text-secondary outline-none transition-colors duration-120 ease-out hover:border-strong data-[state=open]:border-signal',
          props.triggerClass,
        )
      "
    >
      <span aria-hidden="true">{{ props.label }}</span>
      <SelectValue
        :placeholder="props.placeholder"
        class="flex-1 truncate text-left text-primary"
      />
      <img :src="chevronDownIcon" alt="" class="size-2.5 shrink-0" />
    </SelectTrigger>
    <SelectPortal>
      <SelectContent
        position="popper"
        :side-offset="CONTENT_SIDE_OFFSET"
        class="z-50 min-w-[var(--reka-select-trigger-width)] overflow-hidden rounded-md border border-default bg-raised p-1 shadow-[0_0.25rem_0.75rem_rgba(0,0,0,0.4)]"
      >
        <SelectViewport class="max-h-60 overflow-y-auto">
          <SelectItem
            v-for="option in props.options"
            :key="option.value"
            :value="option.value"
            class="flex cursor-pointer items-center rounded px-2 py-1.5 font-sans text-xs text-secondary outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-40 data-[highlighted]:bg-hover data-[highlighted]:text-primary data-[state=checked]:text-signal"
          >
            <SelectItemText>{{ option.label }}</SelectItemText>
          </SelectItem>
        </SelectViewport>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>
