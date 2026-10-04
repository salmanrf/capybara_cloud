import { pluralize } from "@/lib/format";
import type { OrganizationEntry } from "../../data/types";

/**
 * @params entry: the current Organization
 * @return the org card: gradient tile, name, "Organization · N members" and the switcher affordance
 * Heads the sidebar. The switcher (⇅) is not functional yet.
 */
export function OrgCard({ entry }: { entry: OrganizationEntry }) {
  return (
    <button
      type="button"
      className="flex w-full cursor-pointer items-center gap-2.5 rounded-[10px] border border-line bg-surface px-2.5 py-[9px] text-left hover:border-line-strong"
    >
      <span className="size-6 shrink-0 rounded-[7px] bg-[linear-gradient(135deg,#5FB8A5,#8B87E8)]" />
      <span className="flex min-w-0 flex-1 flex-col">
        <span className="text-[13px] font-semibold">{entry.organization.name}</span>
        <span className="text-[11px] text-fg-subtle">Organization · {pluralize(entry.view.memberCount, "member")}</span>
      </span>
      <span className="text-[10px] text-fg-subtle">⇅</span>
    </button>
  );
}
