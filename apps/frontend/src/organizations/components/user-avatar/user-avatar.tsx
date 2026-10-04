import type { User } from "@/api";
import { initials } from "@/lib/initials";

/**
 * @params user: the signed-in user
 * @return a round avatar with the user's initials
 * Shows which account is signed in.
 */
export function UserAvatar({ user }: { user: User }) {
  return (
    <div className="grid size-[30px] cursor-pointer place-items-center rounded-full border border-line-strong bg-avatar text-[11px] font-semibold text-fg-secondary">
      {initials(user.fullName)}
    </div>
  );
}
