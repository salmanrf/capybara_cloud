/**
 * @params fullName: a person's full name
 * @return up to two uppercase letters: the first letters of the first two words, or the first
 *   two letters of a single-word name; "?" when the name is blank
 * Builds the letters shown in an avatar, so it never renders empty.
 */
export function initials(fullName: string): string {
  const words = fullName.trim().split(/\s+/).filter((word) => word !== "");
  const first = words[0];
  if (first === undefined) {
    return "?";
  }
  const second = words[1];
  if (second === undefined) {
    return first.slice(0, 2).toUpperCase();
  }
  return `${first.charAt(0)}${second.charAt(0)}`.toUpperCase();
}
