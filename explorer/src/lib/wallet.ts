import { p256 } from '@noble/curves/nist.js';
import { sha256 } from '@noble/hashes/sha2.js';
import { gcm } from '@noble/ciphers/aes.js';
import { scrypt } from 'scrypt-js';
import { randomBytes } from '@noble/ciphers/utils.js';
import { generateMnemonic, mnemonicToSeedSync, validateMnemonic } from '@scure/bip39';
import { wordlist } from '@scure/bip39/wordlists/english.js';

function safeHexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
  }
  return bytes;
}

function safeBytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('');
}

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
    privateKey: safeBytesToHex(privBytes),
    publicKey: safeBytesToHex(pubPoint),
    address,
  };
}

export function generateKeyPairFromMnemonic(): { mnemonic: string; keyPair: KeyPair } {
  const mnemonic = generateMnemonic(wordlist, 128);
  const keyPair = recoverKeyPairFromMnemonic(mnemonic);
  return { mnemonic, keyPair };
}

export function recoverKeyPairFromMnemonic(mnemonic: string): KeyPair {
  if (!validateMnemonic(mnemonic, wordlist)) {
    throw new Error('Invalid mnemonic');
  }
  const seed = mnemonicToSeedSync(mnemonic, '');
  const privBytes = seed.slice(0, 32);
  const N = BigInt('0xFFFFFFFF00000000FFFFFFFFFFFFFFFFBCE6FAADA7179E84F3B9CAC2FC632551');
  let privInt = BigInt('0x' + safeBytesToHex(privBytes));
  privInt = privInt % N;
  if (privInt === 0n) throw new Error('Derived key is zero');
  const privHex = privInt.toString(16).padStart(64, '0');
  const privKeyBytes = safeHexToBytes(privHex);
  const pubPoint = p256.getPublicKey(privKeyBytes, true);
  const address = pubKeyToAddress(pubPoint);
  return {
    privateKey: privHex,
    publicKey: safeBytesToHex(pubPoint),
    address,
  };
}

export function pubKeyToAddress(pubKeyBytes: Uint8Array): string {
  const hash = sha256(pubKeyBytes);
  return '0x' + safeBytesToHex(hash.slice(12));
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
  const txHash = safeBytesToHex(hashBytes);

  const sigBytes = p256.sign(jsonBytes, safeHexToBytes(privateKeyHex), { lowS: true });
  const sigHex = safeBytesToHex(sigBytes);

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
  const salt = randomBytes(32);
  const dk = await scrypt(
    new TextEncoder().encode(password),
    salt,
    32768, 8, 1, 32
  );

  const nonce = randomBytes(12);
  const privBytes = safeHexToBytes(privateKeyHex);
  const cipher = gcm(new Uint8Array(dk), nonce);
  const ciphertext = cipher.encrypt(privBytes);

  return {
    version: 1,
    address,
    publicKey: publicKeyHex,
    crypto: {
      cipher: 'aes-256-gcm',
      ciphertext: safeBytesToHex(ciphertext),
      nonce: safeBytesToHex(nonce),
      kdf: 'scrypt',
      kdfparams: {
        n: 32768,
        r: 8,
        p: 1,
        dklen: 32,
        salt: safeBytesToHex(salt),
      },
    },
  };
}

export async function decryptKeystore(
  ks: KeystoreFile,
  password: string
): Promise<KeyPair> {
  const salt = safeHexToBytes(ks.crypto.kdfparams.salt);
  const dk = await scrypt(
    new TextEncoder().encode(password),
    salt,
    ks.crypto.kdfparams.n,
    ks.crypto.kdfparams.r,
    ks.crypto.kdfparams.p,
    ks.crypto.kdfparams.dklen
  );

  const nonce = safeHexToBytes(ks.crypto.nonce);
  const ciphertext = safeHexToBytes(ks.crypto.ciphertext);
  const decipher = gcm(new Uint8Array(dk), nonce);
  const plaintext = decipher.decrypt(ciphertext);

  const privateKeyHex = safeBytesToHex(plaintext);
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
