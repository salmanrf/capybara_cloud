import { queryOptions } from "@tanstack/react-query";
import type { Api } from "./api";

/**
 * @params api: the api instance
 * @return query options for the current user (or null when signed out)
 * Session lookup cached for a minute. Invalidate it after signin to refresh guarded routes.
 */
export const sessionQuery = (api: Api) =>
  queryOptions({
    queryKey: ["session"],
    queryFn: () => api.auth.me(),
    staleTime: 60_000,
  });
