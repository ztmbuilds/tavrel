<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";

import searchIcon from "@/assets/icons/search.svg?component";
import xIcon from "@/assets/icons/x.svg?component";
import Button from "@/shared/components/ui/Button.vue";
import Input from "@/shared/components/ui/Input.vue";

const props = defineProps<{
  inputId: string;
  label: string;
  placeholder?: string;
}>();

const modelValue = defineModel<string>({ required: true });

function isTypingTarget(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement ||
    (target instanceof HTMLElement && target.isContentEditable)
  );
}

function focusInput(): void {
  document.getElementById(props.inputId)?.focus();
}

function handleShortcut(event: KeyboardEvent): void {
  if (event.key !== "/" || event.ctrlKey || event.metaKey || event.altKey)
    return;
  if (isTypingTarget(event.target)) return;
  event.preventDefault();
  focusInput();
}

onMounted(() => {
  document.addEventListener("keydown", handleShortcut);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleShortcut);
});

function clearSearch(): void {
  modelValue.value = "";
  focusInput();
}
</script>

<template>
  <Input
    :id="inputId"
    v-model="modelValue"
    :placeholder="placeholder"
    :aria-label="label"
    autocomplete="off"
    spellcheck="false"
    class="w-full"
  >
    <template #leading>
      <component :is="searchIcon" />
    </template>
    <template #trailing>
      <kbd
        v-if="modelValue === ''"
        class="border-strong text-tertiary hidden shrink-0 rounded border px-1 font-sans text-[0.6875rem] sm:block"
        aria-hidden="true"
        >/</kbd
      >
      <Button
        v-else
        variant="ghost"
        size="sm"
        class="size-5 shrink-0 p-0 [&_svg]:size-3"
        aria-label="Clear search"
        @click="clearSearch"
      >
        <component :is="xIcon" aria-hidden="true" />
      </Button>
    </template>
  </Input>
</template>
