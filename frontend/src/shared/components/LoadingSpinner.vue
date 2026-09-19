<script setup lang="ts">
import spinnerIcon from "@/assets/icons/spinner.svg";
import { cn } from "@/shared/utils/cn";

const props = withDefaults(
  defineProps<{
    label: string;
    size?: "xs" | "sm" | "md";
  }>(),
  { size: "md" },
);

const sizeClass =
  props.size === "xs" ? "size-3.5" : props.size === "sm" ? "size-4" : "size-6";
</script>

<template>
  <span role="status" class="inline-flex items-center justify-center">
    <img
      :src="spinnerIcon"
      alt=""
      aria-hidden="true"
      :class="cn('shrink-0 animate-spin', sizeClass)"
    />
    <span class="sr-only">{{ props.label }}</span>
  </span>
</template>
