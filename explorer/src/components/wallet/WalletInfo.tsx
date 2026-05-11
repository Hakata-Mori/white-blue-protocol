import { useCallback, useState } from 'react';
import { Link } from 'react-router-dom';
import { useFetch } from '../../hooks/useFetch';
import { getWallet } from '../../api/wallet';
import Amount from '../ui/Amount';
import LoadingSpinner from '../ui/LoadingSpinner';

interface WalletInfoProps {
  address: string;
  publicKey: string;
}

export default function WalletInfo({ address, publicKey }: WalletInfoProps) {
  const [copied, setCopied] = useState(false);

  const fetcher = useCallback(() => getWallet(address), [address]);
  const { data: account, loading, error } = useFetch(fetcher);

  const handleCopy = () => {
    navigator.clipboard.writeText(address);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (loading && !account) {
    return <LoadingSpinner />;
  }

  const display = account ?? {
    address,
    publicKey,
    whiteBalance: 0,
    blueBalances: {},
    nonce: 0,
    createdAt: 0,
    stakedBalance: 0,
  };

  const blueEntries = Object.entries(display.blueBalances);

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
      <h2 className="text-lg font-bold mb-4">Account Info</h2>
      {error && (
        <div className="bg-yellow-900/50 border border-yellow-700 rounded-lg p-3 mb-4">
          <span className="text-yellow-300 text-sm">Account not found on-chain.</span>
        </div>
      )}
      <div className="space-y-3">
        <div className="flex justify-between items-center py-2 border-b border-gray-700/50">
          <span className="text-gray-400">Address</span>
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-gray-100 break-all">{address}</span>
            <button
              onClick={handleCopy}
              className="text-xs text-blue-400 hover:text-blue-300 whitespace-nowrap"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        </div>
        <div className="flex justify-between py-2 border-b border-gray-700/50">
          <span className="text-gray-400">White Balance</span>
          <Amount value={display.whiteBalance} />
        </div>
        {(display.stakedBalance ?? 0) > 0 && (
          <div className="flex justify-between py-2 border-b border-gray-700/50">
            <span className="text-gray-400">Staked Balance</span>
            <Amount value={display.stakedBalance!} />
          </div>
        )}
        <div className="flex justify-between py-2 border-b border-gray-700/50">
          <span className="text-gray-400">Nonce</span>
          <span className="font-mono">{display.nonce}</span>
        </div>
        <div className="flex justify-between py-2">
          <span className="text-gray-400">Public Key</span>
          <span className="font-mono text-sm text-gray-300 break-all">{publicKey}</span>
        </div>
      </div>
      {blueEntries.length > 0 && (
        <div className="mt-6">
          <h3 className="text-md font-bold mb-3">Blue Coin Holdings</h3>
          <div className="space-y-2">
            {blueEntries.map(([tokenId, balance]) => (
              <div key={tokenId} className="flex justify-between py-2 border-b border-gray-700/50">
                <Link
                  to={`/bluecoin/${tokenId}`}
                  className="font-mono text-blue-400 hover:text-blue-300 text-sm"
                >
                  {tokenId}
                </Link>
                <Amount value={balance} suffix="BC" />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
