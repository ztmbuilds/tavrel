<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";

import checkIcon from "@/assets/icons/check.svg";
import chevronDownIcon from "@/assets/icons/chevron-down.svg?component";
import searchIcon from "@/assets/icons/search.svg?component";
import Button from "@/shared/components/ui/Button.vue";
import Input from "@/shared/components/ui/Input.vue";
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
const draft = ref<string[]>([]);
const rootRef = ref<HTMLDivElement | null>(null);
const triggerRef = ref<HTMLButtonElement | null>(null);
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

const selectedCount = computed(() => modelValue.value.length);

const serviceRows = computed(() => {
  const query = filter.value.trim().toLowerCase();
  return props.options.filter(
    (option) =>
      option.value !== "all" && option.label.toLowerCase().includes(query),
  );
});

function isDraftSelected(value: string): boolean {
  return draft.value.includes(value);
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
  draft.value = [...modelValue.value];
  open.value = true;
  void nextTick(() => {
    document.getElementById(`${props.label}-search`)?.focus();
  });
}

function toggleValue(value: string): void {
  if (isDraftSelected(value)) {
    draft.value = draft.value.filter((item) => item !== value);
  } else {
    draft.value = [...draft.value, value];
  }
}

function clearDraft(): void {
  draft.value = [];
}

function applyDraft(): void {
  const changed =
    draft.value.length !== modelValue.value.length ||
    draft.value.some((value, index) => value !== modelValue.value[index]);
  if (changed) modelValue.value = [...draft.value];
  close(true);
}

function handleOutside(event: PointerEvent): void {
  if (
    open.value &&
    rootRef.value !== null &&
    !rootRef.value.contains(event.target as Node | null)
  )
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
      :aria-expanded="open"
      aria-haspopup="listbox"
      :aria-controls="`${props.label}-listbox`"
      :class="
        cn(
          'raised text-heading inline-flex h-8 min-w-0 cursor-pointer items-center justify-between gap-2 rounded-lg pr-2.5 pl-3',
          'transition-[background-image,box-shadow] duration-150',
          open && 'raised-pressed',
          props.triggerClass,
        )
      "
      @click="toggle"
      @keydown.escape="close(true)"
    >
      <span class="text-secondary shrink-0">{{ props.label }}</span>
      <span class="min-w-0 flex-1 truncate text-left font-medium">{{
        firstLabel
      }}</span>
      <span
        v-if="selectedCount > 1"
        class="bg-raised text-heading ring-default flex size-5 shrink-0 items-center justify-center rounded-full text-[0.6875rem] font-medium tabular-nums ring-1 ring-inset"
        >{{ selectedCount }}</span
      >
      <component :is="chevronDownIcon" class="text-secondary size-4 shrink-0" />
    </button>
    <div
      v-if="open"
      class="border-default bg-raised animate-pop shadow-pop absolute top-[calc(100%+0.375rem)] left-0 z-30 w-80 max-w-[calc(100vw-2rem)] rounded-xl border p-2 motion-reduce:animate-none"
      @keydown="handlePanelKeydown"
    >
      <Input
        :id="`${props.label}-search`"
        v-model="filter"
        :aria-label="`Search ${props.label.toLowerCase()}s`"
        :placeholder="props.searchPlaceholder"
        autocomplete="off"
        spellcheck="false"
        class="bg-canvas mb-2 h-8 text-xs"
        @click.stop
      >
        <template #leading>
          <component :is="searchIcon" />
        </template>
      </Input>
      <div
        :id="`${props.label}-listbox`"
        ref="listRef"
        role="listbox"
        aria-multiselectable="true"
        :aria-label="`${props.label} options`"
        class="scroll-thin flex max-h-60 flex-col gap-0.5 overflow-y-auto"
      >
        <button
          v-for="option in serviceRows"
          :key="option.value"
          type="button"
          role="option"
          :aria-selected="isDraftSelected(option.value)"
          :class="
            cn(
              'text-primary focus-visible:outline-signal flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-1.5 text-left font-sans text-[0.8125rem] outline-none select-none hover:bg-white/[0.03] focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-solid',
            )
          "
          @click="toggleValue(option.value)"
        >
          <span
            :class="
              cn(
                'flex size-4 shrink-0 items-center justify-center rounded-[4px] border',
                isDraftSelected(option.value)
                  ? 'border-signal bg-signal'
                  : 'border-strong bg-transparent',
              )
            "
            aria-hidden="true"
          >
            <img
              v-if="isDraftSelected(option.value)"
              :src="checkIcon"
              alt=""
              class="size-3"
            />
          </span>
          <span class="truncate">{{ option.label }}</span>
        </button>
      </div>
      <p
        v-if="serviceRows.length === 0"
        class="text-tertiary m-0 px-2.5 py-2 text-xs"
      >
        No {{ props.label.toLowerCase() }}s match “{{ filter.trim() }}”.
      </p>
      <div
        class="border-subtle mt-2 flex items-center justify-between gap-2 border-t pt-2"
      >
        <Button
          variant="default"
          size="sm"
          :disabled="draft.length === 0"
          @click="clearDraft"
        >
          Clear
        </Button>
        <Button variant="primary" size="sm" @click="applyDraft"> Apply </Button>
      </div>
    </div>
  </div>
</template>
