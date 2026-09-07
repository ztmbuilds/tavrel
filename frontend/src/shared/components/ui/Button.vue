<script setup lang="ts">
import type { PrimitiveProps } from "reka-ui";

import { Primitive } from "reka-ui";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/shared/utils/cn";

const props = withDefaults(defineProps<ButtonProps>(), {
  as: "button",
});

const buttonVariants = cva(
  "inline-flex min-h-7 cursor-pointer items-center justify-center gap-2 rounded-md px-3 text-xs font-medium whitespace-nowrap transition-colors duration-120 ease-out outline-none disabled:pointer-events-none disabled:opacity-40",
  {
    variants: {
      variant: {
        default: "bg-signal font-semibold text-ink hover:bg-signal-hover",
        secondary:
          "border border-default bg-raised text-primary hover:border-strong hover:bg-hover",
        ghost: "text-tertiary hover:bg-surface hover:text-secondary",
      },
      size: {
        sm: "min-h-7 px-3",
        md: "min-h-[1.875rem] px-3.5",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "md",
    },
  },
);

type ButtonVariants = VariantProps<typeof buttonVariants>;

interface ButtonProps extends PrimitiveProps {
  variant?: ButtonVariants["variant"];
  size?: ButtonVariants["size"];
  class?: string;
}
</script>

<template>
  <Primitive
    :as="props.as"
    :class="
      cn(
        buttonVariants({ variant: props.variant, size: props.size }),
        props.class,
      )
    "
  >
    <slot />
  </Primitive>
</template>
