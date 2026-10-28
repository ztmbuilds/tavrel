<script setup lang="ts">
import Chip from "@/shared/components/ui/Chip.vue";

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
    <Chip
      v-for="chip in chips"
      :key="chip.key"
      :label="chip.label"
      :value="chip.value"
      @remove="$emit('remove', chip.key)"
    />
  </div>
</template>
