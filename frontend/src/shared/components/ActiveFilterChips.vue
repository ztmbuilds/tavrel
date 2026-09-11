<script setup lang="ts">
import xIcon from "@/assets/icons/x.svg";

export interface ActiveFilterChip {
  key: string;
  label: string;
  value: string;
}

defineProps<{
  chips: readonly ActiveFilterChip[];
}>();

defineEmits<{
  remove: [key: string];
}>();
</script>

<template>
  <div v-if="chips.length > 0" class="flex flex-wrap gap-1.5">
    <span
      v-for="chip in chips"
      :key="chip.key"
      class="border-default bg-surface text-secondary flex items-center gap-1.5 rounded-full border py-1 pr-1.5 pl-2.5 font-sans text-xs whitespace-nowrap"
    >
      {{ chip.label }}:
      <span class="text-primary font-medium">{{ chip.value }}</span>
      <button
        type="button"
        :aria-label="`Remove ${chip.label} filter`"
        class="bg-hover flex size-4 cursor-pointer items-center justify-center rounded-full border-none p-0 transition-colors duration-120 ease-out"
        @click="$emit('remove', chip.key)"
      >
        <img :src="xIcon" alt="" class="size-2.5" />
      </button>
    </span>
  </div>
</template>
