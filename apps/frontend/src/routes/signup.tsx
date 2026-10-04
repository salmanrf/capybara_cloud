import { useMutation } from "@tanstack/react-query";
import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import type { FormEvent } from "react";
import { Field, FormError, readFields } from "../components/form";

export const Route = createFileRoute("/signup")({
  component: Signup,
});

/**
 * @params none
 * @return the signup page
 * Creates an account and sends the user to the signin page on success.
 */
function Signup() {
  const { api } = Route.useRouteContext();
  const router = useRouter();

  const signup = useMutation({
    mutationFn: api.auth.signup,
    onSuccess: () => router.navigate({ to: "/signin" }),
  });

  const onSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    signup.mutate(readFields(e.currentTarget, ["email", "password", "username", "fullName"]));
  };

  return (
    <main className="page">
      <h1>Sign up</h1>
      <form className="form" onSubmit={onSubmit}>
        <Field label="Full name" name="fullName" autoComplete="name" required />
        <Field label="Username" name="username" autoComplete="username" minLength={4} maxLength={100} required />
        <Field label="Email" name="email" type="email" autoComplete="email" required />
        <Field label="Password" name="password" type="password" autoComplete="new-password" minLength={8} required />
        <FormError error={signup.error} />
        <button type="submit" disabled={signup.isPending}>
          {signup.isPending ? "Creating account…" : "Create account"}
        </button>
      </form>
      <p>
        Already have an account? <Link to="/signin">Sign in</Link>
      </p>
    </main>
  );
}
