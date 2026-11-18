<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { queryClient } from "@/app/query-client";
import alertIcon from "@/assets/icons/alert-circle.svg?component";
import arrowRightIcon from "@/assets/icons/arrow-right-long.svg?component";
import eyeOffIcon from "@/assets/icons/eye-off.svg?component";
import eyeIcon from "@/assets/icons/eye.svg?component";
import lockIcon from "@/assets/icons/lock.svg?component";
import logo from "@/assets/icons/logo.svg";
import spinnerIcon from "@/assets/icons/spinner-current.svg?component";
import AuthAside from "@/features/auth/components/AuthAside.vue";
import { authKeys, useLoginMutation } from "@/features/auth/queries";
import { ApiError } from "@/shared/api/client";
import { cn } from "@/shared/utils/cn";

const router = useRouter();
const route = useRoute();
const loginMutation = useLoginMutation();

const token = ref("");
const showToken = ref(false);
const errorMessage = ref<string | null>(null);
const tokenInput = ref<HTMLInputElement | null>(null);

const isPending = computed(() => loginMutation.isPending.value);
const tokenFieldClass = computed(() =>
  cn(
    "bg-auth-field flex h-12.5 items-center gap-2.5 rounded-xl border pr-1.5 pl-3.5 transition-[border-color,box-shadow] duration-150 focus-within:ring-3",
    errorMessage.value === null
      ? "border-auth-line hover:border-auth-line-hover focus-within:border-auth-focus focus-within:ring-auth-focus/25"
      : "border-error focus-within:ring-error/25",
  ),
);

onMounted(() => {
  tokenInput.value?.focus();
});

function clearError(): void {
  errorMessage.value = null;
}

function toggleTokenVisibility(): void {
  showToken.value = !showToken.value;
}

function focusToken(): void {
  tokenInput.value?.focus();
  tokenInput.value?.select();
}

function loginErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.isUnauthorized())
      return "That token doesn’t look right. Check it and try again.";
    if (error.status === 0) return "Could not reach the server. Try again.";
  }

  return "Something went wrong. Try again.";
}

function redirectTarget(): string {
  const redirect = route.query.redirect;
  if (
    typeof redirect === "string" &&
    redirect.startsWith("/") &&
    !redirect.startsWith("//")
  ) {
    return redirect;
  }

  return "/";
}

async function submitLogin(): Promise<void> {
  if (isPending.value) return;

  const candidate = token.value.trim();
  if (candidate === "") {
    errorMessage.value = "Enter your token to continue.";
    focusToken();
    return;
  }

  errorMessage.value = null;

  try {
    const session = await loginMutation.mutateAsync(candidate);
    queryClient.clear();
    queryClient.setQueryData(authKeys.session(), session);
    await router.replace(redirectTarget());
  } catch (error) {
    errorMessage.value = loginErrorMessage(error);
    focusToken();
  }
}

function handleSubmit(): void {
  void submitLogin();
}
</script>

<template>
  <div class="bg-canvas flex min-h-dvh flex-wrap p-3 lg:flex-nowrap">
    <AuthAside class="hidden lg:block" />

    <main
      class="flex min-w-0 shrink grow basis-[27.5rem] items-center justify-center px-6 py-10 sm:px-8 sm:py-12"
    >
      <div class="max-w-auth w-full">
        <img :src="logo" alt="tavrel" class="mb-10 size-9 lg:hidden" />

        <form
          class="flex flex-col gap-7"
          novalidate
          @submit.prevent="handleSubmit"
        >
          <div class="flex flex-col gap-2.5">
            <h1
              class="text-heading text-3xl leading-[1.15] font-semibold tracking-[-0.02em]"
            >
              Let’s get you in
            </h1>
            <p class="text-tertiary text-[0.9375rem] leading-normal">
              Enter your access token to start exploring.
            </p>
          </div>

          <div class="flex flex-col gap-2.5">
            <label
              for="token"
              class="text-primary text-[0.8125rem] font-medium"
            >
              Access token
            </label>

            <div :class="tokenFieldClass">
              <component
                :is="lockIcon"
                aria-hidden="true"
                class="text-tertiary size-4.5 shrink-0"
              />
              <input
                id="token"
                ref="tokenInput"
                v-model="token"
                :type="showToken ? 'text' : 'password'"
                name="token"
                placeholder="Enter your token"
                autocomplete="current-password"
                autocapitalize="off"
                autocorrect="off"
                spellcheck="false"
                :aria-invalid="errorMessage !== null"
                :aria-describedby="errorMessage ? 'token-error' : undefined"
                class="text-heading placeholder:text-tertiary h-full min-w-0 flex-1 bg-transparent text-[0.9375rem] tracking-[0.01em] outline-none focus-visible:outline-none"
                @input="clearError"
              />
              <button
                type="button"
                class="text-tertiary hover:bg-raised hover:text-heading focus-visible:outline-auth-focus flex size-9.5 shrink-0 items-center justify-center rounded-lg transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2"
                :aria-label="showToken ? 'Hide token' : 'Show token'"
                :aria-pressed="showToken"
                @click="toggleTokenVisibility"
              >
                <component
                  :is="showToken ? eyeOffIcon : eyeIcon"
                  aria-hidden="true"
                  class="size-4.5"
                />
              </button>
            </div>

            <div
              v-if="errorMessage"
              id="token-error"
              role="alert"
              class="text-error-text flex items-center gap-2 text-[0.8125rem]"
            >
              <component
                :is="alertIcon"
                aria-hidden="true"
                class="size-4 shrink-0"
              />
              <span>{{ errorMessage }}</span>
            </div>
          </div>

          <div class="flex flex-col items-center">
            <button
              type="submit"
              :disabled="isPending"
              :aria-busy="isPending"
              class="text-ink bg-heading focus-visible:outline-auth-focus flex h-12.5 w-full items-center justify-center gap-2.5 rounded-xl text-[0.9375rem] font-medium transition-[background-color,transform] duration-150 hover:bg-white focus-visible:outline-2 focus-visible:outline-offset-2 active:scale-[0.99] disabled:cursor-progress disabled:opacity-75"
            >
              <component
                :is="spinnerIcon"
                v-if="isPending"
                aria-hidden="true"
                class="size-4.5 animate-spin"
              />
              <span>{{ isPending ? "Signing you in…" : "Sign in" }}</span>
              <component
                :is="arrowRightIcon"
                v-if="!isPending"
                aria-hidden="true"
                class="size-4.5"
              />
            </button>
          </div>
        </form>
      </div>
    </main>
  </div>
</template>
