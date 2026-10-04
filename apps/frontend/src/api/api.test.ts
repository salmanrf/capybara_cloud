import { describe, expect, it } from "vitest";
import { ApiError, createApi } from ".";

type Call = { url: string; init: RequestInit | undefined };

function fakeFetch(status: number, body: unknown) {
  const calls: Call[] = [];
  const fetch = async (url: RequestInfo | URL, init?: RequestInit) => {
    calls.push({ url: String(url), init });
    return new Response(JSON.stringify(body), { status });
  };
  return { fetch: fetch as typeof globalThis.fetch, calls };
}

const ok = (data: unknown, message = "ok") => ({ success: true, message, data, error: null });
const fail = (code: number, message: string) => ({
  success: false,
  message,
  data: null,
  error: { error_code: code, message, context: {} },
});

describe("api.auth.me", () => {
  it("returns the signed-in user", async () => {
    const { fetch, calls } = fakeFetch(200, ok({ user_id: "u1", email: "a@b.co", full_name: "Ada" }));
    const api = createApi({ fetch });

    expect(await api.auth.me()).toEqual({ userId: "u1", email: "a@b.co", fullName: "Ada" });
    expect(calls[0]?.url).toBe("/api/auth/me");
    expect(calls[0]?.init?.credentials).toBe("same-origin");
  });

  it("returns null when there is no session", async () => {
    const { fetch } = fakeFetch(401, fail(401, "Unauthorized"));
    expect(await createApi({ fetch }).auth.me()).toBeNull();
  });

  it("throws on server errors", async () => {
    const { fetch } = fakeFetch(500, fail(500, "Internal server error"));
    await expect(createApi({ fetch }).auth.me()).rejects.toBeInstanceOf(ApiError);
  });
});

describe("api.auth.signin", () => {
  it("posts credentials as JSON", async () => {
    const { fetch, calls } = fakeFetch(200, ok(null));
    await createApi({ fetch }).auth.signin({ email: "a@b.co", password: "Secret1!" });

    expect(calls[0]?.init?.method).toBe("POST");
    expect(JSON.parse(String(calls[0]?.init?.body))).toEqual({ email: "a@b.co", password: "Secret1!" });
  });

  it("surfaces the backend's message", async () => {
    const { fetch } = fakeFetch(400, fail(400, "Incorrect username/email"));
    await expect(createApi({ fetch }).auth.signin({ email: "a@b.co", password: "x" })).rejects.toMatchObject({
      status: 400,
      message: "Incorrect username/email",
    });
  });
});

describe("api.auth.signup", () => {
  it("maps fields to the backend's snake_case", async () => {
    const { fetch, calls } = fakeFetch(200, ok(null));
    await createApi({ fetch }).auth.signup({
      email: "a@b.co",
      password: "Secret1!",
      username: "adal",
      fullName: "Ada Lovelace",
    });

    expect(JSON.parse(String(calls[0]?.init?.body))).toEqual({
      email: "a@b.co",
      password: "Secret1!",
      username: "adal",
      full_name: "Ada Lovelace",
    });
  });
});

describe("network failure", () => {
  it("throws ApiError with status 0", async () => {
    const fetch = (async () => {
      throw new TypeError("Failed to fetch");
    }) as typeof globalThis.fetch;
    await expect(createApi({ fetch }).auth.signin({ email: "a", password: "b" })).rejects.toMatchObject({ status: 0 });
  });
});
