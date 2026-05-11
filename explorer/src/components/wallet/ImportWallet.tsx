import { useState, useRef } from 'react';
import { decryptKeystore } from '../../lib/wallet';
import type { KeyPair, KeystoreFile } from '../../lib/wallet';
import LoadingSpinner from '../ui/LoadingSpinner';

interface ImportWalletProps {
  onImported: (kp: KeyPair) => void;
}

export default function ImportWallet({ onImported }: ImportWalletProps) {
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [fileName, setFileName] = useState('');
  const [keystoreData, setKeystoreData] = useState<KeystoreFile | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setFileName(file.name);
    setError('');
    const reader = new FileReader();
    reader.onload = (ev) => {
      try {
        const ks = JSON.parse(ev.target!.result as string) as KeystoreFile;
        setKeystoreData(ks);
      } catch {
        setError('Invalid keystore file');
        setKeystoreData(null);
      }
    };
    reader.readAsText(file);
  };

  const handleUnlock = async () => {
    if (!keystoreData) return;
    setError('');
    setLoading(true);
    try {
      const kp = await decryptKeystore(keystoreData, password);
      onImported(kp);
    } catch {
      setError('Wrong password or corrupted keystore');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div>
        <LoadingSpinner />
        <p className="text-center text-gray-400 text-sm mt-2">Decrypting keystore...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm text-gray-400 mb-1">Keystore File</label>
        <div
          onClick={() => fileRef.current?.click()}
          className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 cursor-pointer hover:border-gray-500 transition-colors"
        >
          {fileName || 'Click to select keystore JSON file...'}
        </div>
        <input
          ref={fileRef}
          type="file"
          accept=".json"
          onChange={handleFileChange}
          className="hidden"
        />
      </div>
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
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <button
        onClick={handleUnlock}
        disabled={!keystoreData || !password}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Unlock
      </button>
    </div>
  );
}
