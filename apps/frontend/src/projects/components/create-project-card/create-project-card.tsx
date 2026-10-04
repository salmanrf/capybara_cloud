/**
 * @params none
 * @return the dashed "Create a project" card
 * Entry point for adding a Project at the end of the grid. Not functional yet.
 */
export function CreateProjectCard() {
  return (
    <button
      type="button"
      className="grid min-h-40 cursor-pointer place-items-center rounded-[14px] border border-dashed border-line-dashed p-5 text-[13px] text-fg-subtle hover:border-faint hover:text-fg-muted"
    >
      <span className="flex flex-col items-center gap-2">
        <span className="text-xl font-light">+</span>
        <span>Create a project</span>
      </span>
    </button>
  );
}
