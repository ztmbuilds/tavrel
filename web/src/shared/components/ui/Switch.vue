<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { SwitchRoot, SwitchThumb } from "reka-ui";

import { cn } from "@/shared/utils/cn";

const props = defineProps<{
  ariaLabel?: string;
  disabled?: boolean;
  class?: HTMLAttributes["class"];
}>();
const model = defineModel<boolean>({ default: false });
</script>

<template>
  <SwitchRoot
    v-model="model"
    :aria-label="props.ariaLabel"
    :disabled="props.disabled"
    :class="
      cn(
        'border-default bg-raised inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border p-0',
        'shadow-[inset_0_1px_2px_rgb(0_0_0/0.45)] transition-colors duration-150 motion-reduce:transition-none',
        'data-[state=checked]:bg-signal data-[disabled]:pointer-events-none data-[disabled]:opacity-50 data-[state=checked]:border-transparent',
        props.class,
      )
    "
  >
    <SwitchThumb
      class="bg-secondary data-[state=checked]:bg-ink block size-3.5 translate-x-0.5 rounded-full shadow-sm transition-transform duration-150 data-[state=checked]:translate-x-[18px] motion-reduce:transition-none"
    />
  </SwitchRoot>
</template>
