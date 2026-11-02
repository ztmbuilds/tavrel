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
  <div class="flex flex-col items-center gap-3 px-5 py-14 text-center">
    <p
      class="text-tertiary m-0 max-w-[52ch] text-[0.8125rem] leading-relaxed text-pretty"
    >
      {{ message ?? title }}
    </p>
    <Button
      v-if="actionLabel"
      variant="ghost"
      size="sm"
      @click="$emit('clear')"
    >
      {{ actionLabel }}
    </Button>
  </div>
</template>
