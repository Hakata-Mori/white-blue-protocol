import { useState, useCallback } from 'react';
import type { KeyPair } from '../lib/wallet';

const STORAGE_KEY = 'wblue_wallet_address';
const PUBKEY_KEY = 'wblue_wallet_pubkey';

export interface WalletState {
  address: string | null;
  publicKey: string | null;
  privateKey: string | null;
  isUnlocked: boolean;
}

export function useWallet() {
  const [state, setState] = useState<WalletState>(() => ({
    address: localStorage.getItem(STORAGE_KEY),
    publicKey: localStorage.getItem(PUBKEY_KEY),
    privateKey: null,
    isUnlocked: false,
  }));

  const unlock = useCallback((kp: KeyPair) => {
    localStorage.setItem(STORAGE_KEY, kp.address);
    localStorage.setItem(PUBKEY_KEY, kp.publicKey);
    setState({
      address: kp.address,
      publicKey: kp.publicKey,
      privateKey: kp.privateKey,
      isUnlocked: true,
    });
  }, []);

  const lock = useCallback(() => {
    setState(prev => ({
      ...prev,
      privateKey: null,
      isUnlocked: false,
    }));
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY);
    localStorage.removeItem(PUBKEY_KEY);
    setState({
      address: null,
      publicKey: null,
      privateKey: null,
      isUnlocked: false,
    });
  }, []);

  return { ...state, unlock, lock, logout };
}
