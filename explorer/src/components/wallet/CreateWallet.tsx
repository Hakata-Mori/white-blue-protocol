import { useState } from 'react';
import { generateKeyPairFromMnemonic, encryptKeystore } from '../../lib/wallet';
import type { KeyPair } from '../../lib/wallet';
import LoadingSpinner from '../ui/LoadingSpinner';

interface CreateWalletProps {
  onCreated: (kp: KeyPair) => void;
}

export default function CreateWallet({ onCreated }: CreateWalletProps) {
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [createdAddress, setCreatedAddress] = useState('');
  const [mnemonic, setMnemonic] = useState('');
  const [keyPair, setKeyPair] = useState<KeyPair | null>(null);
  const [mnemonicSaved, setMnemonicSaved] = useState(false);

  const handleCreate = () => {
    setError('');
    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    try {
      const result = generateKeyPairFromMnemonic();
      setMnemonic(result.mnemonic);
      setKeyPair(result.keyPair);
    } catch (e) {
      setError(String(e));
    }
  };

  const handleDownloadKeystore = async () => {
    if (!keyPair) return;
    setLoading(true);
    try {
      const ks = await encryptKeystore(keyPair.privateKey, keyPair.publicKey, keyPair.address, password);
      const blob = new Blob([JSON.stringify(ks, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `keystore-${keyPair.address}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      setCreatedAddress(keyPair.address);
      onCreated(keyPair);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  if (mnemonic && keyPair) {
    const words = mnemonic.split(' ');
    return (
      <div className="space-y-4">
        <div className="p-4 bg-amber-900/30 border border-amber-600 rounded-lg">
          <p className="text-amber-400 font-bold text-sm mb-2">Write down your mnemonic phrase</p>
          <p className="text-amber-300 text-xs mb-3">
            Store these 12 words in a safe place. They are the only way to recover your wallet. Never share them with anyone.
          </p>
          <div className="grid grid-cols-3 gap-2">
            {words.map((word, i) => (
              <div key={i} className="bg-gray-800 border border-gray-600 rounded px-2 py-1 text-center">
                <span className="text-gray-500 text-xs mr-1">{i + 1}.</span>
                <span className="text-gray-100 text-sm font-mono">{word}</span>
              </div>
            ))}
          </div>
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={mnemonicSaved}
            onChange={(e) => setMnemonicSaved(e.target.checked)}
            className="w-4 h-4 rounded border-gray-600 bg-gray-700 text-blue-600 focus:ring-blue-500"
          />
          <span className="text-gray-300 text-sm">I have saved my mnemonic phrase</span>
        </label>
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <button
          onClick={handleDownloadKeystore}
          disabled={!mnemonicSaved}
          className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Download Keystore & Complete
        </button>
        {createdAddress && (
          <div className="p-3 bg-green-900/30 border border-green-700 rounded-lg">
            <p className="text-green-400 text-sm">Wallet created successfully!</p>
            <p className="text-gray-300 text-sm font-mono mt-1 break-all">{createdAddress}</p>
          </div>
        )}
      </div>
    );
  }

  return (
    <div>
      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="Enter password"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Confirm Password</label>
          <input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="Confirm password"
          />
        </div>
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <button
          onClick={handleCreate}
          disabled={!password || !confirmPassword}
          className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Create Wallet
        </button>
      </div>
    </div>
  );
}
