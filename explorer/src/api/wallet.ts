import { fetchJSON } from './client';
import type { Account } from '../types';

export function getWallet(address: string): Promise<Account> {
  return fetchJSON<Account>(`/api/v1/wallet/${address}`);
}
