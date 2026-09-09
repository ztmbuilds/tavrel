<script setup lang="ts">
import searchIcon from "@/assets/icons/search.svg";

defineProps<{
  inputId: string;
  label: string;
  placeholder?: string;
}>();

const modelValue = defineModel<string>({ required: true });
</script>

<template>
  <div
    class="border-default bg-surface text-tertiary focus-within:border-signal flex min-h-7 min-w-64 items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs transition-colors duration-120 ease-out"
  >
    <img :src="searchIcon" alt="" class="size-3.5 shrink-0" />
    <label class="sr-only" :for="inputId">{{ label }}</label>
    <input
      :id="inputId"
      v-model="modelValue"
      type="text"
      class="text-primary placeholder:text-tertiary w-full border-none bg-transparent text-xs outline-none"
      :placeholder="placeholder"
      autocomplete="off"
      spellcheck="false"
    />
  </div>
</template>
