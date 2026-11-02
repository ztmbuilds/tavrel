<script setup lang="ts">
import { computed, useSlots } from "vue";

const props = defineProps<{
  title: string;
  description?: string;
  descriptionLive?: boolean;
}>();

const slots = useSlots();
const hasActions = computed(() => slots.actions !== undefined);
</script>

<template>
  <div
    class="shrink-0 px-10 pt-7 pb-5 max-[40rem]:px-4 max-[40rem]:pt-5 max-[40rem]:pb-4"
  >
    <div
      class="flex flex-wrap items-end justify-between gap-x-4 gap-y-3 max-[40rem]:flex-col max-[40rem]:items-stretch"
    >
      <div class="min-w-0 flex-1">
        <h1
          class="text-heading text-[1.375rem] leading-[1.3] font-semibold tracking-[-0.015em] text-balance"
        >
          {{ props.title }}
        </h1>
        <p
          v-if="props.description"
          class="text-tertiary mt-1 text-xs leading-relaxed tabular-nums"
          :aria-live="props.descriptionLive ? 'polite' : undefined"
        >
          {{ props.description }}
        </p>
      </div>
      <div
        v-if="hasActions"
        class="flex flex-wrap items-center gap-2 max-[40rem]:w-full"
      >
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>
