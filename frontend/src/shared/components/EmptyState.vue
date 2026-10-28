<script setup lang="ts">
import Button from "@/shared/components/ui/Button.vue";

defineProps<{
  title: string;
  message?: string | undefined;
  actionLabel?: string | undefined;
}>();

defineEmits<{
  clear: [];
}>();
</script>

<template>
  <div
    class="text-tertiary flex flex-col items-center gap-3.5 px-5 py-12 text-center text-[0.8125rem]"
  >
    <div class="flex max-w-[52ch] flex-col items-center gap-1.5">
      <span class="text-heading text-sm font-semibold text-balance">{{
        title
      }}</span>
      <span v-if="message" class="leading-relaxed text-pretty">{{
        message
      }}</span>
    </div>
    <Button v-if="actionLabel" @click="$emit('clear')">
      {{ actionLabel }}
    </Button>
  </div>
</template>
