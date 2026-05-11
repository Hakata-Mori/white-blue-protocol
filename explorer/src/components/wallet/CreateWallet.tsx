import { useState } from 'react';
import { generateKeyPair, encryptKeystore } from '../../lib/wallet';
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

  const handleCreate = async () => {
    setError('');
    if (password.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setLoading(true);
    try {
      const kp = generateKeyPair();
      const ks = await encryptKeystore(kp.privateKey, kp.publicKey, kp.address, password);
      const blob = new Blob([JSON.stringify(ks, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `keystore-${kp.address}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      setCreatedAddress(kp.address);
      onCreated(kp);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
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
      {createdAddress && (
        <div className="mt-4 p-3 bg-green-900/30 border border-green-700 rounded-lg">
          <p className="text-green-400 text-sm">Wallet created successfully!</p>
          <p className="text-gray-300 text-sm font-mono mt-1 break-all">{createdAddress}</p>
        </div>
      )}
    </div>
  );
}
