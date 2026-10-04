import { ApiError } from "./errors";

/** Mirrors utils.BaseResponse in packages/shared-go/utils/http_response.go. */
type Envelope<T> = {
  success: boolean;
  message: string;
  data: T;
  error: { error_code: number; message: string; context: Record<string, unknown> | null } | null;
};

export type Fetch = typeof fetch;

export type Request = <T>(method: string, path: string, body?: unknown) => Promise<T>;

/**
 * @params fetchImpl: the fetch to call; baseUrl: prefix for every path
 * @return a request function resolving to the envelope's `data`; it throws ApiError on network failures, non-2xx statuses and unsuccessful envelopes
 * Internal seam: every endpoint goes through here, so URL prefix, cookies,
 * JSON encoding and envelope unwrapping live in one place.
 */
export function createRequest(fetchImpl: Fetch, baseUrl: string): Request {
  return async <T>(method: string, path: string, body?: unknown): Promise<T> => {
    let res: Response;
    try {
      res = await fetchImpl(`${baseUrl}${path}`, {
        method,
        credentials: "same-origin",
        headers: body === undefined ? undefined : { "Content-Type": "application/json" },
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (cause) {
      throw new ApiError(0, "Unable to reach the server", { cause });
    }

    let envelope: Envelope<T> | undefined;
    try {
      envelope = (await res.json()) as Envelope<T>;
    } catch {
      envelope = undefined;
    }

    if (!res.ok || !envelope?.success) {
      throw new ApiError(
        res.status,
        envelope?.error?.message ?? envelope?.message ?? res.statusText ?? "Request failed",
        envelope?.error?.context ?? {},
      );
    }

    return envelope.data;
  };
}
