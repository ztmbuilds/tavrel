<script setup lang="ts">
import { onUnmounted, ref } from "vue";

import checkOkIcon from "@/assets/icons/check-ok.svg?component";
import copyIcon from "@/assets/icons/copy.svg?component";
import { cn } from "@/shared/utils/cn";

const props = defineProps<{
  text: string;
  label: string;
  title?: string;
  compact?: boolean;
}>();

const COPY_FLASH_MS = 1600;
const copied = ref(false);
let copyTimer: ReturnType<typeof setTimeout> | undefined;

function copyText(): void {
  void navigator.clipboard?.writeText(props.text).catch(() => {});
  copied.value = true;
  if (copyTimer !== undefined) clearTimeout(copyTimer);
  copyTimer = setTimeout(() => {
    copied.value = false;
  }, COPY_FLASH_MS);
}

onUnmounted(() => {
  if (copyTimer !== undefined) clearTimeout(copyTimer);
});
</script>

<template>
  <button
    type="button"
    :aria-label="copied ? 'Copied' : label"
    :title="copied ? 'Copied' : title"
    :class="
      cn(
        'text-tertiary hover:bg-hover flex h-5 shrink-0 cursor-pointer items-center justify-center rounded-md border-none bg-transparent p-0 opacity-0 transition-[width,opacity] duration-100 ease-out group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:opacity-100',
        compact
          ? 'w-0 overflow-hidden group-focus-within:w-5 group-hover:w-5 focus-visible:w-5 [@media(hover:none)]:w-5 [@media(hover:none)]:opacity-100'
          : 'w-5',
      )
    "
    @click.stop="copyText"
    @keydown.stop
  >
    <component
      :is="copied ? checkOkIcon : copyIcon"
      aria-hidden="true"
      :class="cn('size-3.5 shrink-0', copied && 'text-ok')"
    />
    <span v-if="copied" class="sr-only" role="status">Copied</span>
  </button>
</template>
