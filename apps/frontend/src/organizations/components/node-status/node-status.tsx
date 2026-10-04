import type { NodeUsage } from "../../data/types";

/**
 * @params node: the host node's name and memory usage, or null when it is unknown
 * @return the sidebar footer with a health dot and "node · X of Y GB used", or null when the node is unknown
 * A quick read on the host's capacity.
 */
export function NodeStatus({ node }: { node: NodeUsage | null }) {
  if (node === null) {
    return null;
  }
  return (
    <div className="flex items-center gap-2.5 border-t border-line-muted p-2.5 text-xs text-fg-subtle">
      <span className="size-1.5 rounded-full bg-status-running" />
      <span className="whitespace-nowrap">
        {node.name} · {node.usedGb} of {node.totalGb} GB used
      </span>
    </div>
  );
}
