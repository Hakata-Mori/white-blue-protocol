import { fetchJSON } from './client';
import type { BlueCoinConfig, BlueCoinState } from '../types';

export function getBlueCoins(): Promise<BlueCoinConfig[]> {
  return fetchJSON<BlueCoinConfig[]>('/api/v1/bluecoin');
}

export function getBlueCoin(tokenId: string): Promise<BlueCoinConfig> {
  return fetchJSON<BlueCoinConfig>(`/api/v1/bluecoin/${tokenId}`);
}

export function getBlueCoinState(tokenId: string): Promise<BlueCoinState> {
  return fetchJSON<BlueCoinState>(`/api/v1/bluecoin/${tokenId}/state`);
}
