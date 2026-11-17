import {
  useMutation,
  useQuery,
  type UseMutationReturnType,
  type UseQueryReturnType,
} from "@tanstack/vue-query";

import { getAuthSession, login, logout } from "./api";
import type { AuthSession } from "./types";

export const authKeys = {
  all: ["auth"] as const,
  session: () => [...authKeys.all, "session"] as const,
};

export function authSessionQueryOptions(): {
  queryKey: ReturnType<typeof authKeys.session>;
  queryFn: typeof getAuthSession;
  retry: boolean;
  staleTime: "static";
} {
  return {
    queryKey: authKeys.session(),
    queryFn: getAuthSession,
    retry: false,
    staleTime: "static",
  };
}

export function useAuthSessionQuery(): UseQueryReturnType<AuthSession, Error> {
  return useQuery(authSessionQueryOptions());
}

export function useLoginMutation(): UseMutationReturnType<
  AuthSession,
  Error,
  string,
  unknown
> {
  return useMutation({ mutationFn: login });
}

export function useLogoutMutation(): UseMutationReturnType<
  void,
  Error,
  void,
  unknown
> {
  return useMutation({ mutationFn: logout });
}
