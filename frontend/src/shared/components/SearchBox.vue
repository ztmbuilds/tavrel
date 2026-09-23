<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";

import searchIcon from "@/assets/icons/search.svg?component";
import xIcon from "@/assets/icons/x.svg?component";

defineProps<{
  inputId: string;
  label: string;
  placeholder?: string;
}>();

const modelValue = defineModel<string>({ required: true });

const inputRef = ref<HTMLInputElement | null>(null);

function isTypingTarget(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement ||
    (target instanceof HTMLElement && target.isContentEditable)
  );
}

function handleShortcut(event: KeyboardEvent): void {
  if (event.key !== "/" || event.ctrlKey || event.metaKey || event.altKey)
    return;
  if (isTypingTarget(event.target)) return;
  event.preventDefault();
  inputRef.value?.focus();
}

onMounted(() => {
  document.addEventListener("keydown", handleShortcut);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleShortcut);
});

function clearSearch(): void {
  modelValue.value = "";
  inputRef.value?.focus();
}
</script>

<template>
  <div
    class="border-default bg-surface text-tertiary focus-within:border-signal flex min-h-9 w-full items-center gap-2 rounded-lg border px-3 py-2 text-[0.8125rem] transition-colors duration-120 ease-out"
  >
    <component :is="searchIcon" class="size-3.5 shrink-0" />
    <label class="sr-only" :for="inputId">{{ label }}</label>
    <input
      :id="inputId"
      ref="inputRef"
      v-model="modelValue"
      type="text"
      class="text-primary placeholder:text-tertiary w-full min-w-0 flex-1 border-none bg-transparent text-[0.8125rem] outline-none"
      :placeholder="placeholder"
      autocomplete="off"
      spellcheck="false"
    />
    <kbd
      v-if="modelValue === ''"
      class="border-strong text-tertiary hidden shrink-0 rounded border px-1 font-sans text-[0.6875rem] sm:block"
      aria-hidden="true"
      >/</kbd
    >
    <button
      v-else
      type="button"
      class="hover:bg-hover flex size-5 shrink-0 cursor-pointer items-center justify-center rounded border-none bg-transparent p-0 transition-colors duration-120 ease-out"
      aria-label="Clear search"
      @click="clearSearch"
    >
      <component :is="xIcon" class="size-3" />
    </button>
  </div>
</template>
