import { Outlet, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed")({
  component: AuthedLayout,
});

/**
 * @params none
 * @return the matched child route
 * Layout route for every page that requires a session.
 * TEMPORARY: auth is bypassed, so there is no guard and signed-out users are not redirected.
 * When auth is wired, restore the `beforeLoad` that reads `sessionQuery` and redirects to
 * /signin (with a `redirect` back), and read `sessionQuery` inside `useAuth`.
 */
function AuthedLayout() {
  return <Outlet />;
}
