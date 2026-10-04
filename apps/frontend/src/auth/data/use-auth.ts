import type { User } from "@/api";

/** TEMPORARY: the user returned while auth is bypassed. Only the fields `/auth/me` returns. */
const MOCK_USER: User = {
  userId: "7d1f3c2a-5b9e-4c1d-8a6f-2e4b9c0d1a37",
  email: "rafi@masbro.dev",
  fullName: "Rafi Akbar",
};

/**
 * @params none
 * @return `{ user }`, the signed-in user
 * The only way components learn who is signed in.
 * TEMPORARY: auth is bypassed, so this returns a mocked user. When auth is wired,
 * read `sessionQuery` here (and restore the `_authed` guard); call sites do not change.
 */
export function useAuth(): { user: User } {
  return { user: MOCK_USER };
}
