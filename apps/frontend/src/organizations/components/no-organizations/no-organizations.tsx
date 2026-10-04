/**
 * @params none
 * @return a short message saying the user belongs to no Organization
 * Shown in the main area instead of a page when the user has no Organizations.
 */
export function NoOrganizations() {
  return (
    <div className="flex flex-col items-center gap-2 py-24 text-center">
      <h1 className="m-0 text-lg font-semibold">No organizations yet</h1>
      <p className="m-0 text-[13px] text-fg-muted">You are not a member of any Organization.</p>
    </div>
  );
}
