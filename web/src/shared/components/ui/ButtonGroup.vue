<script setup lang="ts">
import type { HTMLAttributes } from "vue";

import { cn } from "@/shared/utils/cn";

const props = defineProps<{ class?: HTMLAttributes["class"] }>();
</script>

<template>
  <div
    role="group"
    :class="
      cn(
        'shadow-raised inline-flex rounded-lg',
        '[&>a]:rounded-none [&>a]:shadow-none [&>button]:rounded-none [&>button]:shadow-none',
        '[&>*:first-child]:rounded-l-lg [&>*:last-child]:rounded-r-lg',
        '[&>*+*]:border-l-black/45',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
