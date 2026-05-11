import { useState, useCallback } from 'react';
import { useWallet } from '../hooks/useWallet';
import { useFetch } from '../hooks/useFetch';
import { getWallet } from '../api/wallet';
import type { KeyPair } from '../lib/wallet';
import CreateWallet from '../components/wallet/CreateWallet';
import ImportWallet from '../components/wallet/ImportWallet';
import WalletInfo from '../components/wallet/WalletInfo';
import TransferForm from '../components/wallet/TransferForm';
import BlueTransferForm from '../components/wallet/BlueTransferForm';
import SwapForm from '../components/wallet/SwapForm';
import DeployForm from '../components/wallet/DeployForm';

type Tab = 'transfer' | 'blue' | 'swap' | 'deploy';

export default function WalletPage() {
  const wallet = useWallet();
  const [activeTab, setActiveTab] = useState<Tab>('transfer');
  const [refreshKey, setRefreshKey] = useState(0);

  const fetcher = useCallback(
    () => wallet.address ? getWallet(wallet.address) : Promise.reject('no address'),
    [wallet.address, refreshKey]
  );

  const { data: account } = useFetch(fetcher);

  const handleCreated = (kp: KeyPair) => {
    wallet.unlock(kp);
  };

  const handleImported = (kp: KeyPair) => {
    wallet.unlock(kp);
  };

  const handleTabChange = (tab: Tab) => {
    setActiveTab(tab);
    setRefreshKey((k) => k + 1);
  };

  if (!wallet.address) {
    return (
      <div>
        <h1 className="text-2xl font-bold mb-6">Wallet</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
            <h2 className="text-lg font-bold mb-4">Create New Wallet</h2>
            <CreateWallet onCreated={handleCreated} />
          </div>
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
            <h2 className="text-lg font-bold mb-4">Import Wallet</h2>
            <ImportWallet onImported={handleImported} />
          </div>
        </div>
      </div>
    );
  }

  if (!wallet.isUnlocked) {
    return (
      <div>
        <h1 className="text-2xl font-bold mb-6">Wallet</h1>
        <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 mb-6">
          <div className="flex justify-between items-center">
            <div>
              <p className="text-gray-400 text-sm">Address</p>
              <p className="font-mono text-gray-100 break-all">{wallet.address}</p>
            </div>
            <button
              onClick={wallet.logout}
              className="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
            >
              Logout
            </button>
          </div>
        </div>
        <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
          <h2 className="text-lg font-bold mb-4">Import to Unlock</h2>
          <ImportWallet onImported={handleImported} />
        </div>
      </div>
    );
  }

  const currentNonce = account?.nonce ?? 0;
  const blueBalances = account?.blueBalances ?? {};

  const tabs: { key: Tab; label: string }[] = [
    { key: 'transfer', label: 'Transfer' },
    { key: 'blue', label: 'Blue Transfer' },
    { key: 'swap', label: 'Swap' },
    { key: 'deploy', label: 'Deploy' },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Wallet</h1>
        <div className="flex gap-2">
          <button
            onClick={() => { wallet.lock(); }}
            className="bg-gray-700 hover:bg-gray-600 text-gray-300 font-medium py-2 px-4 rounded-lg transition-colors"
          >
            Lock
          </button>
          <button
            onClick={() => { wallet.logout(); }}
            className="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
          >
            Logout
          </button>
        </div>
      </div>

      <div className="mb-6">
        <WalletInfo address={wallet.address} publicKey={wallet.publicKey!} />
      </div>

      <div className="flex gap-2 mb-6">
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

      {activeTab === 'transfer' && (
        <TransferForm
          from={wallet.address}
          publicKey={wallet.publicKey!}
          privateKey={wallet.privateKey!}
          currentNonce={currentNonce}
        />
      )}
      {activeTab === 'blue' && (
        <BlueTransferForm
          from={wallet.address}
          publicKey={wallet.publicKey!}
          privateKey={wallet.privateKey!}
          currentNonce={currentNonce}
          blueBalances={blueBalances}
        />
      )}
      {activeTab === 'swap' && (
        <SwapForm
          from={wallet.address}
          publicKey={wallet.publicKey!}
          privateKey={wallet.privateKey!}
          currentNonce={currentNonce}
        />
      )}
      {activeTab === 'deploy' && (
        <DeployForm
          from={wallet.address}
          publicKey={wallet.publicKey!}
          privateKey={wallet.privateKey!}
          currentNonce={currentNonce}
        />
      )}
    </div>
  );
}
