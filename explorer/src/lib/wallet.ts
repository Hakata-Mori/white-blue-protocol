import { p256 } from '@noble/curves/nist.js';
import { sha256 } from '@noble/hashes/sha2.js';
import { bytesToHex, hexToBytes } from '@noble/curves/utils.js';
import { scrypt } from 'scrypt-js';

export interface KeyPair {
  privateKey: string;
  publicKey: string;
  address: string;
}

export interface KeystoreFile {
  version: number;
  address: string;
  publicKey: string;
  crypto: {
    cipher: string;
    ciphertext: string;
    nonce: string;
    kdf: string;
    kdfparams: {
      n: number;
      r: number;
      p: number;
      dklen: number;
      salt: string;
    };
  };
}

export interface TransactionInput {
  type: number;
  from: string;
  to: string;
  amount: number;
  tokenId: string;
  fee: number;
  nonce: number;
  payload: string | null;
  publicKey: string;
  timestamp: number;
  minAmountOut?: number;
}

export interface SignedTransaction extends TransactionInput {
  hash: string;
  signature: string;
}

export function generateKeyPair(): KeyPair {
  const privBytes = p256.utils.randomSecretKey();
  const pubPoint = p256.getPublicKey(privBytes, true);
  const address = pubKeyToAddress(pubPoint);
  return {
    privateKey: bytesToHex(privBytes),
    publicKey: bytesToHex(pubPoint),
    address,
  };
}

export function pubKeyToAddress(pubKeyBytes: Uint8Array): string {
  const hash = sha256(pubKeyBytes);
  return '0x' + bytesToHex(hash.slice(12));
}

function canonicalJSON(tx: TransactionInput & { hash: string; signature: string }): string {
  const obj: Record<string, unknown> = {
    hash: tx.hash,
    type: tx.type,
    from: tx.from,
    to: tx.to,
    amount: tx.amount,
    tokenId: tx.tokenId,
    fee: tx.fee,
    nonce: tx.nonce,
    payload: tx.payload,
    publicKey: tx.publicKey,
    signature: tx.signature,
    timestamp: tx.timestamp,
  };
  if (tx.minAmountOut && tx.minAmountOut > 0) {
    obj.minAmountOut = tx.minAmountOut;
  }
  return JSON.stringify(obj);
}

export function signTransactionWithKey(tx: TransactionInput, privateKeyHex: string): SignedTransaction {
  const forSign = {
    ...tx,
    hash: '',
    signature: '',
  };

  const jsonStr = canonicalJSON(forSign);
  const jsonBytes = new TextEncoder().encode(jsonStr);

  const hashBytes = sha256(jsonBytes);
  const txHash = bytesToHex(hashBytes);

  const sigBytes = p256.sign(jsonBytes, hexToBytes(privateKeyHex), { lowS: true });
  const sigHex = bytesToHex(sigBytes);

  return {
    ...tx,
    hash: txHash,
    signature: sigHex,
  };
}

export async function encryptKeystore(
  privateKeyHex: string,
  publicKeyHex: string,
  address: string,
  password: string
): Promise<KeystoreFile> {
  const salt = crypto.getRandomValues(new Uint8Array(32));
  const dk = await scrypt(
    new TextEncoder().encode(password),
    salt,
    32768, 8, 1, 32
  );

  const key = await crypto.subtle.importKey('raw', dk, 'AES-GCM', false, ['encrypt']);
  const nonce = crypto.getRandomValues(new Uint8Array(12));
  const privBytes = hexToBytes(privateKeyHex);
  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: nonce },
    key,
    privBytes
  );

  return {
    version: 1,
    address,
    publicKey: publicKeyHex,
    crypto: {
      cipher: 'aes-256-gcm',
      ciphertext: bytesToHex(new Uint8Array(ciphertext)),
      nonce: bytesToHex(nonce),
      kdf: 'scrypt',
      kdfparams: {
        n: 32768,
        r: 8,
        p: 1,
        dklen: 32,
        salt: bytesToHex(salt),
      },
    },
  };
}

export async function decryptKeystore(
  ks: KeystoreFile,
  password: string
): Promise<KeyPair> {
  const salt = hexToBytes(ks.crypto.kdfparams.salt);
  const dk = await scrypt(
    new TextEncoder().encode(password),
    salt,
    ks.crypto.kdfparams.n,
    ks.crypto.kdfparams.r,
    ks.crypto.kdfparams.p,
    ks.crypto.kdfparams.dklen
  );

  const key = await crypto.subtle.importKey('raw', dk, 'AES-GCM', false, ['decrypt']);
  const nonce = hexToBytes(ks.crypto.nonce);
  const ciphertext = hexToBytes(ks.crypto.ciphertext);

  const plaintext = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: nonce },
    key,
    ciphertext
  );

  const privateKeyHex = bytesToHex(new Uint8Array(plaintext));
  return {
    privateKey: privateKeyHex,
    publicKey: ks.publicKey,
    address: ks.address,
  };
}

export function calcFee(amount: number): number {
  const fee = Math.floor(amount / 1000);
  return fee < 1000 ? 1000 : fee;
}

export async function submitTransaction(tx: SignedTransaction): Promise<{ status: string; hash: string }> {
  const res = await fetch('/api/v1/tx/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(tx),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text);
  }
  return res.json();
}
