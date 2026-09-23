<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";

import checkIcon from "@/assets/icons/check.svg";
import chevronDownIcon from "@/assets/icons/chevron-down.svg?component";
import searchIcon from "@/assets/icons/search.svg?component";
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

const modelValue = defineModel<string[]>({ required: true });

const open = ref(false);
const filter = ref("");
const rootRef = ref<HTMLDivElement | null>(null);
const triggerRef = ref<HTMLButtonElement | null>(null);
const searchRef = ref<HTMLInputElement | null>(null);
const listRef = ref<HTMLDivElement | null>(null);

const pinnedOption = computed(() =>
  props.options.find((option) => option.value === "all"),
);

const selectedLabels = computed(() => {
  if (modelValue.value.length === 0)
    return [pinnedOption.value?.label ?? "All"];
  return modelValue.value.map(
    (value) =>
      props.options.find((option) => option.value === value)?.label ?? value,
  );
});

const firstLabel = computed(
  () => selectedLabels.value[0] ?? pinnedOption.value?.label ?? "All",
);

const extraCount = computed(() =>
  selectedLabels.value.length > 1 ? selectedLabels.value.length - 1 : 0,
);

const serviceRows = computed(() => {
  const query = filter.value.trim().toLowerCase();
  return props.options.filter(
    (option) =>
      option.value !== "all" && option.label.toLowerCase().includes(query),
  );
});

function isSelected(value: string): boolean {
  return modelValue.value.includes(value);
}

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

function toggleValue(value: string): void {
  if (value === "all") {
    modelValue.value = [];
    return;
  }
  if (isSelected(value)) {
    modelValue.value = modelValue.value.filter((item) => item !== value);
  } else {
    modelValue.value = [...modelValue.value, value];
  }
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
        firstLabel
      }}</span>
      <span
        v-if="extraCount > 0"
        aria-hidden="true"
        class="bg-hover text-secondary shrink-0 rounded-md px-1.5 py-0.5 text-[0.6875rem] font-medium tabular-nums"
        >+{{ extraCount }}</span
      >
      <component :is="chevronDownIcon" class="text-tertiary size-2.5 shrink-0" />
    </button>
    <div
      v-if="open"
      role="listbox"
      aria-multiselectable="true"
      :aria-label="`${props.label} options`"
      class="border-default bg-raised absolute top-[calc(100%+0.375rem)] left-0 z-30 w-55 rounded-lg border p-1.5 shadow-[0_0.25rem_0.75rem_rgba(0,0,0,0.4)]"
      @keydown="handlePanelKeydown"
    >
      <div
        class="border-default bg-surface text-tertiary mb-1 flex items-center gap-1.5 rounded border px-2 py-1.5"
      >
        <component :is="searchIcon" class="size-3.5 shrink-0" />
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
      <div ref="listRef" class="flex max-h-60 flex-col gap-0.5 overflow-y-auto">
        <button
          v-if="pinnedOption !== undefined"
          type="button"
          role="option"
          :aria-selected="modelValue.length === 0"
          :class="
            cn(
              'flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 font-sans text-[0.8125rem] text-secondary outline-none select-none hover:bg-hover hover:text-primary',
              modelValue.length === 0 && 'bg-hover text-primary',
            )
          "
          @click="toggleValue(pinnedOption.value)"
        >
          <span
            :class="
              cn(
                'flex size-3.5 shrink-0 items-center justify-center rounded border',
                modelValue.length === 0
                  ? 'border-signal bg-signal'
                  : 'border-strong',
              )
            "
            aria-hidden="true"
          >
            <img
              v-if="modelValue.length === 0"
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
          :aria-selected="isSelected(option.value)"
          :class="
            cn(
              'flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 font-sans text-[0.8125rem] text-secondary outline-none select-none hover:bg-hover hover:text-primary',
              isSelected(option.value) && 'bg-hover text-primary',
            )
          "
          @click="toggleValue(option.value)"
        >
          <span
            :class="
              cn(
                'flex size-3.5 shrink-0 items-center justify-center rounded border',
                isSelected(option.value)
                  ? 'border-signal bg-signal'
                  : 'border-strong',
              )
            "
            aria-hidden="true"
          >
            <img
              v-if="isSelected(option.value)"
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
