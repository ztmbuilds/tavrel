<script setup lang="ts">
import {
  TooltipContent,
  TooltipPortal,
  TooltipProvider,
  TooltipRoot,
  TooltipTrigger,
} from "reka-ui";

import { cn } from "@/shared/utils/cn";

type TooltipSide = "top" | "right" | "bottom" | "left";

const props = withDefaults(
  defineProps<{
    label: string;
    side?: TooltipSide;
    sideOffset?: number;
    delayDuration?: number;
    contentClass?: string;
  }>(),
  {
    side: "right",
    sideOffset: 8,
    delayDuration: 300,
  },
);
</script>

<template>
  <TooltipProvider :delay-duration="props.delayDuration">
    <TooltipRoot>
      <TooltipTrigger as-child>
        <slot />
      </TooltipTrigger>
      <TooltipPortal>
        <TooltipContent
          :side="props.side"
          :side-offset="props.sideOffset"
          :class="
            cn(
              'border-default bg-raised text-primary z-50 rounded-md border px-2 py-1 font-sans text-xs shadow-[0_0.25rem_0.75rem_rgba(0,0,0,0.4)]',
              props.contentClass,
            )
          "
        >
          {{ props.label }}
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </TooltipProvider>
</template>
