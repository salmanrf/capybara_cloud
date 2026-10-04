import { useMutation } from "@tanstack/react-query";
import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import type { FormEvent } from "react";
import { Field, FormError, readFields } from "../components/form";
import { sessionQuery } from "../session";

type Search = { redirect?: string };

export const Route = createFileRoute("/signin")({
  validateSearch: validateSearch,
  component: Signin,
});

/**
 * @params search: the raw search params from the URL
 * @return the typed search, keeping `redirect` only when it is a string
 * Validates the signin page's search params.
 */
function validateSearch(search: Record<string, unknown>): Search {
  if (typeof search.redirect !== "string") {
    return {};
  }
  return { redirect: search.redirect };
}

/**
 * @params none
 * @return the signin page
 * Signs the user in, refreshes the session query and sends them to `redirect` (or the dashboard).
 */
function Signin() {
  const { api, queryClient } = Route.useRouteContext();
  const { redirect } = Route.useSearch();
  const router = useRouter();

  const signin = useMutation({
    mutationFn: api.auth.signin,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: sessionQuery(api).queryKey });
      await router.navigate({ to: redirect ?? "/" });
    },
  });

  const onSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    signin.mutate(readFields(e.currentTarget, ["email", "password"]));
  };

  return (
    <main className="page">
      <h1>Sign in</h1>
      <form className="form" onSubmit={onSubmit}>
        <Field label="Email" name="email" type="email" autoComplete="email" required />
        <Field label="Password" name="password" type="password" autoComplete="current-password" required />
        <FormError error={signin.error} />
        <button type="submit" disabled={signin.isPending}>
          {signin.isPending ? "Signing in…" : "Sign in"}
        </button>
      </form>
      <p>
        No account? <Link to="/signup">Sign up</Link>
      </p>
    </main>
  );
}
