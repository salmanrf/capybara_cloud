/**
 * @params none
 * @return the search box with its ⌘K hint
 * Placeholder for search. Visual only for now.
 */
export function SearchBox() {
  return (
    <div className="flex w-[230px] cursor-text items-center gap-2 rounded-lg border border-line bg-surface px-3 py-1.5 text-[13px] text-fg-subtle">
      <span>Search…</span>
      <span className="flex-1" />
      <span className="rounded border border-line-kbd px-[5px] py-px font-mono text-[11px]">⌘K</span>
    </div>
  );
}
