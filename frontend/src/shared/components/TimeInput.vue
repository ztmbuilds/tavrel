<script setup lang="ts">
import { computed } from "vue";

import { cn } from "@/shared/utils/cn";
import { isValidTimeText } from "@/shared/utils/time";

defineProps<{
  inputId: string;
  label: string;
}>();

const modelValue = defineModel<string>({ required: true });

const MAX_TIME_DIGITS = 4;
const HOUR_DIGITS = 2;
const MASKED_TIME_LENGTH = 5;

const showInvalid = computed(
  () =>
    modelValue.value.length === MASKED_TIME_LENGTH &&
    !isValidTimeText(modelValue.value),
);

function handleInput(event: Event): void {
  const target = event.target;
  if (!(target instanceof HTMLInputElement)) return;
  const digits = target.value.replace(/\D/g, "").slice(0, MAX_TIME_DIGITS);
  modelValue.value =
    digits.length <= HOUR_DIGITS
      ? digits
      : `${digits.slice(0, HOUR_DIGITS)}:${digits.slice(HOUR_DIGITS)}`;
}
</script>

<template>
  <input
    :id="inputId"
    :value="modelValue"
    :aria-label="label"
    type="text"
    inputmode="numeric"
    autocomplete="off"
    spellcheck="false"
    placeholder="HH:MM"
    maxlength="5"
    :aria-invalid="showInvalid"
    :class="
      cn(
        'border-default bg-surface text-primary placeholder:text-tertiary focus:border-signal w-full min-w-0 rounded-md border px-1.5 py-1.5 font-sans text-[0.78125rem] tabular-nums transition-colors duration-120 ease-out outline-none',
        showInvalid && 'border-error',
      )
    "
    @input="handleInput"
  />
</template>
