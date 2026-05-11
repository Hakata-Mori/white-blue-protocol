import { useState, useRef } from 'react';
import { decryptKeystore, recoverKeyPairFromMnemonic, encryptKeystore } from '../../lib/wallet';
import type { KeyPair, KeystoreFile } from '../../lib/wallet';
import LoadingSpinner from '../ui/LoadingSpinner';

interface ImportWalletProps {
  onImported: (kp: KeyPair) => void;
}

type ImportTab = 'keystore' | 'mnemonic';

export default function ImportWallet({ onImported }: ImportWalletProps) {
  const [activeTab, setActiveTab] = useState<ImportTab>('keystore');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [fileName, setFileName] = useState('');
  const [keystoreData, setKeystoreData] = useState<KeystoreFile | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const [mnemonicInput, setMnemonicInput] = useState('');
  const [mnemonicPassword, setMnemonicPassword] = useState('');
  const [mnemonicConfirmPassword, setMnemonicConfirmPassword] = useState('');

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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Wrong password or corrupted keystore');
    } finally {
      setLoading(false);
    }
  };

  const handleMnemonicRecover = async () => {
    setError('');
    if (mnemonicPassword.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }
    if (mnemonicPassword !== mnemonicConfirmPassword) {
      setError('Passwords do not match');
      return;
    }
    setLoading(true);
    try {
      const kp = recoverKeyPairFromMnemonic(mnemonicInput.trim());
      const ks = await encryptKeystore(kp.privateKey, kp.publicKey, kp.address, mnemonicPassword);
      const blob = new Blob([JSON.stringify(ks, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `keystore-${kp.address}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      onImported(kp);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to recover from mnemonic');
    } finally {
      setLoading(false);
    }
  };

  const handleTabChange = (tab: ImportTab) => {
    setActiveTab(tab);
    setError('');
  };

  if (loading) {
    return (
      <div>
        <LoadingSpinner />
        <p className="text-center text-gray-400 text-sm mt-2">
          {activeTab === 'keystore' ? 'Decrypting keystore...' : 'Recovering wallet...'}
        </p>
      </div>
    );
  }

  const tabs: { key: ImportTab; label: string }[] = [
    { key: 'keystore', label: 'Keystore File' },
    { key: 'mnemonic', label: 'Mnemonic Phrase' },
  ];

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => handleTabChange(tab.key)}
            className={`py-2 px-4 rounded-lg font-medium transition-colors ${
              activeTab === tab.key
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-400 hover:text-gray-200'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'keystore' && (
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
      )}

      {activeTab === 'mnemonic' && (
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Mnemonic Phrase</label>
            <textarea
              value={mnemonicInput}
              onChange={(e) => setMnemonicInput(e.target.value)}
              className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500 resize-none"
              rows={3}
              placeholder="Enter your 12-word mnemonic phrase separated by spaces"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">Password</label>
            <input
              type="password"
              value={mnemonicPassword}
              onChange={(e) => setMnemonicPassword(e.target.value)}
              className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
              placeholder="Enter password"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">Confirm Password</label>
            <input
              type="password"
              value={mnemonicConfirmPassword}
              onChange={(e) => setMnemonicConfirmPassword(e.target.value)}
              className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
              placeholder="Confirm password"
            />
          </div>
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <button
            onClick={handleMnemonicRecover}
            disabled={!mnemonicInput.trim() || !mnemonicPassword || !mnemonicConfirmPassword}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Recover
          </button>
        </div>
      )}
    </div>
  );
}
