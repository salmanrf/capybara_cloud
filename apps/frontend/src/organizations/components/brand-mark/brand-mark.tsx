/**
 * @params none
 * @return the accent "M" tile with the product name
 * The Masbro Cloud mark shown at the top-left of every dashboard page.
 */
export function BrandMark() {
  return (
    <div className="flex items-center gap-2.5">
      <div className="grid size-6.5 place-items-center rounded-lg bg-accent text-sm font-bold text-on-accent">M</div>
      <span className="text-[15px] font-semibold tracking-[-0.01em]">Masbro Cloud</span>
    </div>
  );
}
