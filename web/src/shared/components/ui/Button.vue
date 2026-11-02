<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { Primitive } from "reka-ui";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{
    variant?: "default" | "primary" | "ghost" | "on-gradient";
    size?: "md" | "sm";
    as?: string;
    asChild?: boolean;
    type?: "button" | "submit" | "reset";
    disabled?: boolean;
    class?: HTMLAttributes["class"];
  }>(),
  { variant: "default", size: "md", as: "button", type: "button" },
);

const base =
  "inline-flex cursor-pointer items-center justify-center gap-2 whitespace-nowrap font-medium select-none no-underline " +
  "transition-[background-color,background-image,box-shadow,transform,filter] duration-150 ease-out motion-reduce:transition-none " +
  "disabled:pointer-events-none disabled:opacity-50 disabled:shadow-none aria-disabled:pointer-events-none aria-disabled:opacity-50 " +
  "[&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:fill-none [&_svg]:stroke-current [&_svg]:stroke-[1.6] [&_svg]:[stroke-linecap:round] [&_svg]:[stroke-linejoin:round]";

const variants = {
  default: "raised rounded-lg text-heading",
  primary:
    "rounded-lg border border-white/25 border-b-white/10 bg-linear-to-b from-signal to-signal-hover text-ink shadow-raised " +
    "hover:brightness-95 active:translate-y-px active:brightness-90 active:shadow-raised-pressed",
  ghost:
    "rounded-lg border border-transparent text-secondary hover:bg-hover hover:text-heading",
  "on-gradient":
    "rounded-lg border border-white/30 bg-white/15 text-white hover:bg-white/25",
} as const;

const sizes = {
  md: "h-8 px-3.5",
  sm: "h-7 px-3 text-xs",
} as const;
</script>

<template>
  <Primitive
    :as="props.as"
    :as-child="props.asChild"
    :type="props.as === 'button' ? props.type : undefined"
    :disabled="props.as === 'button' ? props.disabled : undefined"
    :aria-disabled="props.disabled || undefined"
    :class="cn(base, variants[props.variant], sizes[props.size], props.class)"
  >
    <slot />
  </Primitive>
</template>
