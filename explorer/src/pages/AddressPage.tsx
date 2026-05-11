import { useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useFetch } from '../hooks/useFetch';
import { getWallet } from '../api/wallet';
import HashDisplay from '../components/ui/HashDisplay';
import Amount from '../components/ui/Amount';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import { formatHash } from '../utils/format';

export default function AddressPage() {
  const { address } = useParams<{ address: string }>();

  const fetcher = useCallback(
    () => getWallet(address!),
    [address]
  );

  const { data: account, loading, error } = useFetch(fetcher);

  if (loading && !account) {
    return <LoadingSpinner />;
  }

  const display = account ?? {
    address: address!,
    publicKey: '',
    whiteBalance: 0,
    blueBalances: {},
    nonce: 0,
    createdAt: 0,
    stakedBalance: 0,
  };

  const blueEntries = Object.entries(display.blueBalances);
  const hasBlue = blueEntries.length > 0;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Account</h1>

      {error && (
        <div className="bg-yellow-900/50 border border-yellow-700 rounded-xl p-4 mb-6">
          <span className="text-yellow-300 text-sm">Account not found on-chain. Showing default values.</span>
        </div>
      )}

      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 mb-6">
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Address</span>
          <HashDisplay hash={display.address} truncate={false} />
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">White Balance</span>
          <Amount value={display.whiteBalance} />
        </div>
        {(display.stakedBalance ?? 0) > 0 && (
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Staked Balance</span>
            <Amount value={display.stakedBalance!} />
          </div>
        )}
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Nonce</span>
          <span className="font-mono">{display.nonce}</span>
        </div>
        {display.publicKey && (
          <div className="flex justify-between py-3">
            <span className="text-gray-400">Public Key</span>
            <span className="font-mono">{formatHash(display.publicKey)}</span>
          </div>
        )}
      </div>

      {hasBlue && (
        <div className="mb-6">
          <h2 className="text-xl font-bold mb-4">Blue Coin Holdings</h2>
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-gray-400 text-sm text-left">
                  <th className="pb-3 pr-4">Token ID</th>
                  <th className="pb-3">Balance</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {blueEntries.map(([tokenId, balance]) => (
                  <tr key={tokenId}>
                    <td className="py-3 pr-4">
                      <Link
                        to={`/bluecoin/${tokenId}`}
                        className="font-mono text-blue-400 hover:text-blue-300"
                      >
                        {tokenId}
                      </Link>
                    </td>
                    <td className="py-3">
                      <Amount value={balance} suffix="BC" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 text-center">
        <span className="text-gray-400">Transaction history coming soon</span>
      </div>
    </div>
  );
}
