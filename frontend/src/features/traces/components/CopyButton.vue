<script setup lang="ts">
import checkOkIcon from "@/assets/icons/check-ok.svg?component";
import copyIcon from "@/assets/icons/copy.svg?component";

import { cn } from "@/shared/utils/cn";

defineProps<{
  label: string;
  title?: string;
  copied: boolean;
}>();

defineEmits<{
  copy: [];
}>();
</script>

<template>
  <button
    type="button"
    :aria-label="copied ? 'Copied' : label"
    :title="copied ? 'Copied' : title"
    class="text-tertiary hover:bg-hover flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md border-none bg-transparent p-0 opacity-0 transition-opacity duration-100 ease-out group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:opacity-100"
    @click="$emit('copy')"
  >
    <component
      :is="copied ? checkOkIcon : copyIcon"
      aria-hidden="true"
      :class="cn('size-3.5 shrink-0', copied && 'text-ok')"
    />
    <span v-if="copied" class="sr-only" role="status">Copied</span>
  </button>
</template>
