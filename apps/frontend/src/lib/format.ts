const MINUTE = 60;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

/**
 * @params count: how many there are; noun: the singular noun
 * @return the count with the noun, singular for exactly one ("1 app", "3 apps")
 * Formats a count for UI copy. Only handles nouns pluralised with a trailing "s".
 */
export function pluralize(count: number, noun: string): string {
  if (count === 1) {
    return `1 ${noun}`;
  }
  return `${count} ${noun}s`;
}

/**
 * @params date: the moment to describe; now: the reference moment (defaults to the current time)
 * @return a short relative time: "just now", "12m ago", "1h ago" or "3d ago"
 * Describes how long ago something happened, in the largest whole unit up to days.
 * Dates in the future are treated as "just now".
 */
export function formatRelativeTime(date: Date, now: Date = new Date()): string {
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
  if (seconds < MINUTE) {
    return "just now";
  }
  if (seconds < HOUR) {
    return `${Math.floor(seconds / MINUTE)}m ago`;
  }
  if (seconds < DAY) {
    return `${Math.floor(seconds / HOUR)}h ago`;
  }
  return `${Math.floor(seconds / DAY)}d ago`;
}
