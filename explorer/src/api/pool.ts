import { fetchJSON } from './client';
import type { AMMPool } from '../types';

export function getPool(tokenId: string): Promise<AMMPool> {
  return fetchJSON<AMMPool>(`/api/v1/pool/${tokenId}`);
}
