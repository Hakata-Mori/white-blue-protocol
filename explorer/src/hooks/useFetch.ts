import { usePolling } from './usePolling';

export function useFetch<T>(fetcher: () => Promise<T>) {
  return usePolling(fetcher, 0);
}
