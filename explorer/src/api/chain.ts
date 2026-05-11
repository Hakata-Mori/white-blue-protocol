import { fetchJSON } from './client';
import type { NetworkStats, BlocksResponse, Block } from '../types';

export function getStats(): Promise<NetworkStats> {
  return fetchJSON<NetworkStats>('/api/v1/stats');
}

export function getBlocks(limit: number, offset: number): Promise<BlocksResponse> {
  return fetchJSON<BlocksResponse>(`/api/v1/blocks?limit=${limit}&offset=${offset}`);
}

export function getBlock(height: number): Promise<Block> {
  return fetchJSON<Block>(`/api/v1/chain/block/${height}`);
}

export function getBlockByHash(hash: string): Promise<Block> {
  return fetchJSON<Block>(`/api/v1/block/hash/${hash}`);
}

export function getChainStatus(): Promise<{ height: number; totalMinted: number; chainId: string }> {
  return fetchJSON<{ height: number; totalMinted: number; chainId: string }>('/api/v1/chain/status');
}
