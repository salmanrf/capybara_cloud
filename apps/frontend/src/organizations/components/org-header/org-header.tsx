import { Link } from "@tanstack/react-router";
import type { User } from "@/api";
import type { Organization } from "../../data/types";
import { BrandMark } from "../brand-mark/brand-mark";
import { SearchBox } from "../search-box/search-box";
import { UserAvatar } from "../user-avatar/user-avatar";

type OrgHeaderProps = {
  user: User;
  organization: Organization | null;
};

/**
 * @params user: the signed-in user; organization: the current Organization, or null when there is none
 * @return the fixed dashboard header: mark, breadcrumb, search, Docs/Feedback links and avatar
 * The top bar of every org-level page. The Organization crumb links back to the Overview.
 */
export function OrgHeader({ user, organization }: OrgHeaderProps) {
  let breadcrumb = null;
  if (organization !== null) {
    breadcrumb = (
      <>
        <span className="text-[15px] text-faint">/</span>
        <Link to="/" className="whitespace-nowrap text-[13px] text-fg">
          {organization.name}
        </Link>
      </>
    );
  }

  return (
    <header className="flex h-14.5 shrink-0 items-center gap-3.5 border-b border-line-muted px-7">
      <BrandMark />
      {breadcrumb}
      <div className="flex-1" />
      <SearchBox />
      <nav className="flex items-center gap-4.5 text-[13px]">
        <button type="button" className="cursor-pointer text-fg-muted hover:text-fg">
          Docs
        </button>
        <button type="button" className="cursor-pointer text-fg-muted hover:text-fg">
          Feedback
        </button>
      </nav>
      <UserAvatar user={user} />
    </header>
  );
}
