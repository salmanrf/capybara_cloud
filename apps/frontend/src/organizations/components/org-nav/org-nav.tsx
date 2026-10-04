import { Link } from "@tanstack/react-router";

const NAV_ITEM = "flex w-full cursor-pointer items-center rounded-[7px] px-2.5 py-[7px] text-left text-[13px]";
const NAV_ITEM_INERT = `${NAV_ITEM} text-fg-muted hover:bg-surface-hover hover:text-fg`;

/**
 * @params memberCount: how many members the Organization has
 * @return the org nav: Overview (highlighted when active), Members with its count, Git connections and Organization settings
 * Navigation between org-level pages. Only Overview has a page so far; the others are inert.
 */
export function OrgNav({ memberCount }: { memberCount: number }) {
  return (
    <nav className="mt-1.5 flex flex-col gap-px">
      <Link
        to="/"
        activeOptions={{ exact: true }}
        className={NAV_ITEM}
        activeProps={{ className: "bg-surface-raised font-medium text-fg" }}
        inactiveProps={{ className: "text-fg-muted hover:bg-surface-hover hover:text-fg" }}
      >
        Overview
      </Link>
      <button type="button" className={NAV_ITEM_INERT}>
        Members
        <span className="flex-1" />
        <span className="font-mono text-[10.5px] text-fg-subtle">{memberCount}</span>
      </button>
      <button type="button" className={NAV_ITEM_INERT}>
        Git connections
      </button>
      <button type="button" className={NAV_ITEM_INERT}>
        Organization settings
      </button>
    </nav>
  );
}
