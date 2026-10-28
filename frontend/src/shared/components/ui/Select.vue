<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import {
  SelectContent,
  SelectItem,
  SelectItemIndicator,
  SelectItemText,
  SelectPortal,
  SelectRoot,
  SelectTrigger,
  SelectValue,
  SelectViewport,
} from "reka-ui";

import chevronDownIcon from "@/assets/icons/chevron-down.svg?component";
import checkIcon from "@/assets/icons/check-ok.svg?component";
import { cn } from "@/shared/utils/cn";

export interface SelectOption {
  value: string;
  label: string;
}

const props = defineProps<{
  options: readonly SelectOption[];
  label?: string;
  placeholder?: string;
  ariaLabel?: string;
  disabled?: boolean;
  class?: HTMLAttributes["class"];
}>();

const model = defineModel<string>({ required: true });
</script>

<template>
  <SelectRoot v-model="model" :disabled="props.disabled">
    <SelectTrigger
      :aria-label="props.ariaLabel ?? props.label"
      :class="
        cn(
          'raised text-heading inline-flex h-8 min-w-36 cursor-pointer items-center justify-between gap-2.5 rounded-lg pr-2.5 pl-3',
          'data-[state=open]:raised-pressed data-[placeholder]:text-secondary',
          'transition-[background-image,box-shadow] duration-150 disabled:pointer-events-none disabled:opacity-50',
          props.class,
        )
      "
    >
      <span
        v-if="props.label"
        aria-hidden="true"
        class="text-secondary shrink-0"
        >{{ props.label }}</span
      >
      <SelectValue
        :placeholder="props.placeholder ?? ''"
        class="min-w-0 flex-1 truncate text-left"
      />
      <component
        :is="chevronDownIcon"
        aria-hidden="true"
        class="text-secondary size-4 shrink-0"
      />
    </SelectTrigger>
    <SelectPortal>
      <SelectContent
        position="popper"
        :side-offset="6"
        align="start"
        class="animate-pop border-default bg-raised text-primary shadow-pop z-50 min-w-(--reka-select-trigger-width) rounded-xl border motion-reduce:animate-none"
      >
        <SelectViewport class="p-1.5">
          <SelectItem
            v-for="o in props.options"
            :key="o.value"
            :value="o.value"
            class="data-[highlighted]:bg-hover relative flex min-h-8 cursor-pointer items-center justify-between gap-4 rounded-lg px-3 outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50"
          >
            <SelectItemText>{{ o.label }}</SelectItemText>
            <SelectItemIndicator class="inline-flex">
              <component
                :is="checkIcon"
                aria-hidden="true"
                class="text-heading size-4"
              />
            </SelectItemIndicator>
          </SelectItem>
        </SelectViewport>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>
