import { fetchJSON } from './client';
import type { ValidatorSet } from '../types';

export function getValidators(): Promise<ValidatorSet> {
  return fetchJSON<ValidatorSet>('/api/v1/validators');
}
