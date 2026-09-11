<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";

import checkIcon from "@/assets/icons/check.svg";
import chevronDownIcon from "@/assets/icons/chevron-down.svg";
import searchIcon from "@/assets/icons/search.svg";
import { cn } from "@/shared/utils/cn";

export interface ServiceSelectOption {
  value: string;
  label: string;
}

const props = withDefaults(
  defineProps<{
    label?: string;
    options: readonly ServiceSelectOption[];
    triggerClass?: string;
    searchPlaceholder?: string;
  }>(),
  {
    label: "Service",
    searchPlaceholder: "Search services…",
  },
);

const modelValue = defineModel<string>({ required: true });

const open = ref(false);
const filter = ref("");
const rootRef = ref<HTMLDivElement | null>(null);
const triggerRef = ref<HTMLButtonElement | null>(null);
const searchRef = ref<HTMLInputElement | null>(null);
const listRef = ref<HTMLDivElement | null>(null);

const selectedLabel = computed(
  () =>
    props.options.find((option) => option.value === modelValue.value)?.label ??
    modelValue.value,
);

const pinnedOption = computed(() =>
  props.options.find((option) => option.value === "all"),
);

const serviceRows = computed(() => {
  const query = filter.value.trim().toLowerCase();
  return props.options.filter(
    (option) =>
      option.value !== "all" && option.label.toLowerCase().includes(query),
  );
});

function close(returnFocus: boolean): void {
  open.value = false;
  if (returnFocus) triggerRef.value?.focus();
}

function toggle(): void {
  if (open.value) {
    close(false);
    return;
  }
  filter.value = "";
  open.value = true;
  void nextTick(() => {
    searchRef.value?.focus();
  });
}

function select(value: string): void {
  modelValue.value = value;
  close(true);
}

function handleOutside(event: PointerEvent): void {
  if (open.value && rootRef.value !== null && !rootRef.value.contains(event.target as Node | null))
    close(false);
}

onMounted(() => {
  document.addEventListener("pointerdown", handleOutside);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", handleOutside);
});

function focusOption(step: number): void {
  if (listRef.value === null) return;
  const buttons = Array.from(
    listRef.value.querySelectorAll<HTMLButtonElement>('[role="option"]'),
  );
  if (buttons.length === 0) return;
  const active = document.activeElement;
  const index = buttons.findIndex((button) => button === active);
  const next = buttons[(index + step + buttons.length) % buttons.length];
  next?.focus();
}

function handlePanelKeydown(event: KeyboardEvent): void {
  if (event.key === "ArrowDown") {
    event.preventDefault();
    focusOption(1);
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    focusOption(-1);
  } else if (event.key === "Escape") {
    event.stopPropagation();
    close(true);
  }
}
</script>

<template>
  <div ref="rootRef" class="relative">
    <button
      ref="triggerRef"
      type="button"
      :aria-label="props.label"
      :aria-expanded="open"
      aria-haspopup="listbox"
      :class="
        cn(
          'border-default bg-surface text-secondary hover:border-strong flex min-h-9 cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 font-sans text-[0.8125rem] whitespace-nowrap transition-colors duration-120 ease-out',
          open && 'border-strong',
          props.triggerClass,
        )
      "
      @click="toggle"
      @keydown.escape="close(true)"
    >
      <span aria-hidden="true">{{ props.label }}</span>
      <span class="text-primary min-w-0 flex-1 truncate text-left">{{
        selectedLabel
      }}</span>
      <img :src="chevronDownIcon" alt="" class="size-2.5 shrink-0" />
    </button>
    <div
      v-if="open"
      role="listbox"
      :aria-label="`${props.label} options`"
      class="border-default bg-raised absolute top-[calc(100%+0.375rem)] left-0 z-30 w-55 rounded-lg border p-1.5 shadow-[0_0.25rem_0.75rem_rgba(0,0,0,0.4)]"
      @keydown="handlePanelKeydown"
    >
      <div
        class="border-default bg-surface text-tertiary mb-1 flex items-center gap-1.5 rounded border px-2 py-1.5"
      >
        <img :src="searchIcon" alt="" class="size-3.5 shrink-0" />
        <label class="sr-only" :for="`${props.label}-search`"
          >Search {{ props.label.toLowerCase() }}s</label
        >
        <input
          :id="`${props.label}-search`"
          ref="searchRef"
          v-model="filter"
          type="text"
          :placeholder="props.searchPlaceholder"
          autocomplete="off"
          spellcheck="false"
          class="text-primary placeholder:text-tertiary w-full min-w-0 flex-1 border-none bg-transparent font-sans text-[0.8125rem] outline-none"
          @click.stop
        />
      </div>
      <div ref="listRef" class="max-h-60 overflow-y-auto">
        <button
          v-if="pinnedOption !== undefined"
          type="button"
          role="option"
          :aria-selected="modelValue === pinnedOption.value"
          :class="
            cn(
              'flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 font-sans text-[0.8125rem] text-secondary outline-none select-none hover:bg-hover hover:text-primary',
              modelValue === pinnedOption.value && 'bg-active text-ink hover:bg-active hover:text-ink',
            )
          "
          @click="select(pinnedOption.value)"
        >
          <span
            :class="
              cn(
                'flex size-3.5 shrink-0 items-center justify-center rounded border',
                modelValue === pinnedOption.value
                  ? 'border-ink'
                  : 'border-strong',
              )
            "
            aria-hidden="true"
          >
            <img
              v-if="modelValue === pinnedOption.value"
              :src="checkIcon"
              alt=""
              class="size-2.5"
            />
          </span>
          <span class="truncate">{{ pinnedOption.label }}</span>
        </button>
        <div
          v-if="pinnedOption !== undefined"
          class="border-subtle my-1 border-t"
          aria-hidden="true"
        />
        <button
          v-for="option in serviceRows"
          :key="option.value"
          type="button"
          role="option"
          :aria-selected="modelValue === option.value"
          :class="
            cn(
              'flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 font-sans text-[0.8125rem] text-secondary outline-none select-none hover:bg-hover hover:text-primary',
              modelValue === option.value && 'bg-active text-ink hover:bg-active hover:text-ink',
            )
          "
          @click="select(option.value)"
        >
          <span
            :class="
              cn(
                'flex size-3.5 shrink-0 items-center justify-center rounded border',
                modelValue === option.value ? 'border-ink' : 'border-strong',
              )
            "
            aria-hidden="true"
          >
            <img
              v-if="modelValue === option.value"
              :src="checkIcon"
              alt=""
              class="size-2.5"
            />
          </span>
          <span class="truncate">{{ option.label }}</span>
        </button>
        <div
          v-if="serviceRows.length === 0"
          class="text-tertiary px-2 py-1.5 font-sans text-[0.8125rem]"
        >
          No services match “{{ filter.trim() }}”.
        </div>
      </div>
    </div>
  </div>
</template>
