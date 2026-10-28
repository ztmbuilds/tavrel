<script setup lang="ts">
import {
  TooltipContent,
  TooltipPortal,
  TooltipProvider,
  TooltipRoot,
  TooltipTrigger,
} from "reka-ui";

const props = withDefaults(
  defineProps<{
    content: string;
    side?: "top" | "right" | "bottom" | "left";
    delay?: number;
  }>(),
  { side: "top", delay: 300 },
);
</script>

<template>
  <TooltipProvider :delay-duration="props.delay">
    <TooltipRoot>
      <TooltipTrigger as-child>
        <slot />
      </TooltipTrigger>
      <TooltipPortal>
        <TooltipContent
          :side="props.side"
          :side-offset="8"
          class="animate-pop border-default bg-raised text-primary shadow-pop z-60 max-w-70 rounded-lg border px-2.5 py-1.5 text-xs leading-snug motion-reduce:animate-none"
        >
          {{ props.content }}
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </TooltipProvider>
</template>
