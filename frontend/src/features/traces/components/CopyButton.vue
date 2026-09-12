<script setup lang="ts">
import checkOkIcon from "@/assets/icons/check-ok.svg";
import copyIcon from "@/assets/icons/copy.svg";

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
    class="hover:bg-hover flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md border-none bg-transparent p-0 opacity-0 transition-opacity duration-100 ease-out group-focus-within:opacity-100 group-hover:opacity-100 focus-visible:opacity-100"
    @click="$emit('copy')"
  >
    <img
      :src="copied ? checkOkIcon : copyIcon"
      alt=""
      aria-hidden="true"
      class="size-3.5 shrink-0"
    />
    <span v-if="copied" class="sr-only" role="status">Copied</span>
  </button>
</template>
