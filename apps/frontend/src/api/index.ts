import { createAuthApi } from "./auth";
import { createRequest, type Fetch } from "./request";

export { ApiError } from "./errors";
export type { User, SigninInput, SignupInput } from "./auth";

export type ApiOptions = {
  /** Defaults to the global fetch; tests pass a fake. */
  fetch?: Fetch;
  /** Defaults to `/api`, which the dev server proxies to the Go backend. */
  baseUrl?: string;
};

/**
 * @params options: optional fetch implementation and base URL
 * @return the api object, grouped by domain (auth, ...)
 * The frontend's single seam to the backend.
 */
export function createApi(options: ApiOptions = {}) {
  const fetch_impl = options.fetch ?? globalThis.fetch.bind(globalThis);
  const request = createRequest(fetch_impl, options.baseUrl ?? "/api");
  return {
    auth: createAuthApi(request),
  };
}

export type Api = ReturnType<typeof createApi>;
