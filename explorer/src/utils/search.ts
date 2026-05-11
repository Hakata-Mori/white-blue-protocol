export type SearchResult =
  | { type: 'block_height'; value: number }
  | { type: 'address'; value: string }
  | { type: 'hash'; value: string }
  | { type: 'unknown' };

export function detectSearch(query: string): SearchResult {
  const q = query.trim();
  if (/^\d+$/.test(q)) return { type: 'block_height', value: parseInt(q) };
  if (/^0x[0-9a-fA-F]{40}$/.test(q)) return { type: 'address', value: q };
  if (/^[0-9a-fA-F]{64}$/.test(q)) return { type: 'hash', value: q };
  return { type: 'unknown' };
}
