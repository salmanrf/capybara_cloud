import { ApiError } from "./errors";
import type { Request } from "./request";

export type User = {
  userId: string;
  email: string;
  fullName: string;
};

export type SigninInput = { email: string; password: string };

export type SignupInput = {
  email: string;
  password: string;
  username: string;
  fullName: string;
};

type AuthMeResponse = { user_id: string; email: string; full_name: string };

/**
 * @params err: the error thrown by a request
 * @return true when the error means there is no valid session (401 or 404)
 * Tells "not signed in" apart from real failures of the session lookup.
 */
function isNoSession(err: unknown): boolean {
  if (!(err instanceof ApiError)) {
    return false;
  }
  return err.status === 401 || err.status === 404;
}

/**
 * @params request: the shared request function
 * @return the auth endpoints: me, signin and signup
 * Builds the auth part of the api, mapping the backend's snake_case to camelCase.
 */
export function createAuthApi(request: Request) {
  return {
    /**
     * @params none
     * @return the signed-in user, or null when there is no valid session; throws ApiError on other failures
     * Looks up the current session.
     */
    async me(): Promise<User | null> {
      try {
        const res = await request<AuthMeResponse>("GET", "/auth/me");
        return { userId: res.user_id, email: res.email, fullName: res.full_name };
      } catch (err) {
        if (isNoSession(err)) {
          return null;
        }
        throw err;
      }
    },

    /**
     * @params input: the email and password
     * @return nothing; throws ApiError when the credentials are rejected
     * Starts a session (the server sets the `sid` cookie).
     */
    async signin(input: SigninInput): Promise<void> {
      await request("POST", "/auth/signin", input);
    },

    /**
     * @params input: the new account's email, password, username and full name
     * @return nothing; throws ApiError when the backend rejects the account
     * Creates an account. It does not sign the user in.
     */
    async signup(input: SignupInput): Promise<void> {
      await request("POST", "/auth/signup", {
        email: input.email,
        password: input.password,
        username: input.username,
        full_name: input.fullName,
      });
    },
  };
}
