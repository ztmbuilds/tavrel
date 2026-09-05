import { VueQueryPlugin } from "@tanstack/vue-query";

import { createApp } from "vue";

import App from "./App.vue";
import { queryClient } from "./query-client";
import router from "./router";
import "@/style.css";

const app =
  // eslint-disable-next-line @typescript-eslint/no-unsafe-argument -- SFC type info comes from vue-tsc; typescript-eslint resolves .vue imports as error-typed.
  createApp(App);
app.use(router);
app.use(VueQueryPlugin, { queryClient });
app.mount("#app");
