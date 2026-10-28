<script setup lang="ts">
import { computed } from "vue";

import Input from "@/shared/components/ui/Input.vue";
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

const maskedValue = computed({
  get: () => modelValue.value,
  set: (value: string) => {
    const digits = value.replace(/\D/g, "").slice(0, MAX_TIME_DIGITS);
    modelValue.value =
      digits.length <= HOUR_DIGITS
        ? digits
        : `${digits.slice(0, HOUR_DIGITS)}:${digits.slice(HOUR_DIGITS)}`;
  },
});
</script>

<template>
  <Input
    :id="inputId"
    v-model="maskedValue"
    :aria-label="label"
    inputmode="numeric"
    autocomplete="off"
    spellcheck="false"
    placeholder="HH:MM"
    maxlength="5"
    :aria-invalid="showInvalid || undefined"
    :class="cn('h-7 px-2 tabular-nums', showInvalid && 'border-error')"
  />
</template>
