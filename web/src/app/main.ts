import { VueQueryPlugin } from "@tanstack/vue-query";

import { createApp } from "vue";

import { authKeys } from "@/features/auth/queries";
import { setUnauthorizedHandler } from "@/shared/api/client";

import App from "./App.vue";
import { queryClient } from "./query-client";
import router from "./router";
import "@/style.css";

const app =
  // eslint-disable-next-line @typescript-eslint/no-unsafe-argument -- SFC type info comes from vue-tsc; typescript-eslint resolves .vue imports as error-typed.
  createApp(App);
app.use(router);
app.use(VueQueryPlugin, { queryClient });

setUnauthorizedHandler(() => {
  queryClient.setQueryData(authKeys.session(), {
    enabled: true,
    authenticated: false,
  });

  const current = router.currentRoute.value;
  if (current.name !== "login") {
    void router.replace({
      name: "login",
      query: { redirect: current.fullPath },
    });
  }
});

app.mount("#app");
