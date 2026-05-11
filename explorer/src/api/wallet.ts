import { fetchJSON } from './client';
import type { Account, AddressTxsResponse } from '../types';

export function getWallet(address: string): Promise<Account> {
  return fetchJSON<Account>(`/api/v1/wallet/${address}`);
}

export function getAddressTxs(address: string, limit: number, offset: number): Promise<AddressTxsResponse> {
  return fetchJSON<AddressTxsResponse>(`/api/v1/address/${address}/txs?limit=${limit}&offset=${offset}`);
}
