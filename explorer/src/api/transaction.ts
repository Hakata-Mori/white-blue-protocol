import { fetchJSON } from './client';
import type { TxReceipt } from '../types';

export function getTx(hash: string): Promise<TxReceipt> {
  return fetchJSON<TxReceipt>(`/api/v1/tx/${hash}`);
}
