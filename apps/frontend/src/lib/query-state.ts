/**
 * The part of a TanStack Query result the UI reads. Fixture-backed hooks return it,
 * so moving a hook to `useQuery` later does not change its call sites.
 */
export type QueryState<T> = {
  data: T | undefined;
  isLoading: boolean;
  error: Error | null;
};

/**
 * @params data: the value the query resolved to
 * @return a settled query state: the data, not loading, no error
 * Wraps synchronously available (fixture) data in the shape of a successful query result.
 */
export function settledQuery<T>(data: T): QueryState<T> {
  return { data, isLoading: false, error: null };
}
