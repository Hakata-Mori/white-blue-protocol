import { useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useFetch } from '../hooks/useFetch';
import { getBlock, getBlockByHash } from '../api/chain';
import HashDisplay from '../components/ui/HashDisplay';
import AddressLink from '../components/ui/AddressLink';
import Amount from '../components/ui/Amount';
import TxTypeLabel from '../components/ui/TxTypeLabel';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import { formatTimestamp } from '../utils/format';

export default function BlockDetailPage() {
  const { height, hash } = useParams<{ height?: string; hash?: string }>();

  const fetcher = useCallback(() => {
    if (hash) return getBlockByHash(hash);
    return getBlock(parseInt(height ?? '0', 10));
  }, [height, hash]);

  const { data: block, loading, error } = useFetch(fetcher);

  if (loading && !block) {
    return <LoadingSpinner />;
  }

  if (error || !block) {
    return (
      <div className="text-center py-12">
        <h1 className="text-2xl font-bold text-red-400 mb-2">Block Not Found</h1>
        <p className="text-gray-400">
          {hash
            ? `No block found with hash ${hash}`
            : `No block found at height ${height}`}
        </p>
        <Link to="/blocks" className="text-blue-400 hover:text-blue-300 mt-4 inline-block">
          Back to blocks
        </Link>
      </div>
    );
  }

  const { header, transactions } = block;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Block #{header.height.toLocaleString()}</h1>
        <div className="flex gap-2">
          <Link
            to={`/block/${header.height - 1}`}
            className={`px-3 py-1 rounded bg-gray-800 text-gray-300 hover:bg-gray-700 ${
              header.height <= 0 ? 'opacity-40 pointer-events-none' : ''
            }`}
          >
            Prev Block
          </Link>
          <Link
            to={`/block/${header.height + 1}`}
            className="px-3 py-1 rounded bg-gray-800 text-gray-300 hover:bg-gray-700"
          >
            Next Block
          </Link>
        </div>
      </div>

      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 mb-6">
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Height</span>
          <span className="font-mono">{header.height.toLocaleString()}</span>
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Hash</span>
          <HashDisplay hash={header.hash} truncate={false} />
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Previous Hash</span>
          <HashDisplay
            hash={header.prevHash}
            truncate
            link={header.height > 0 ? `/block/${header.height - 1}` : undefined}
          />
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Merkle Root</span>
          <span className="font-mono break-all text-right ml-4">{header.merkleRoot}</span>
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Validator</span>
          <AddressLink address={header.validator} truncate={false} />
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Reward</span>
          <Amount value={header.reward} />
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Timestamp</span>
          <span className="text-gray-300">{formatTimestamp(header.timestamp)}</span>
        </div>
        {header.signature && (
          <div className="flex justify-between py-3">
            <span className="text-gray-400">Signature</span>
            <span className="font-mono break-all text-right ml-4 text-sm">{header.signature}</span>
          </div>
        )}
      </div>

      <h2 className="text-xl font-bold mb-4">Transactions ({transactions.length})</h2>
      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 overflow-x-auto">
        {transactions.length === 0 ? (
          <p className="text-gray-400 text-center py-4">No transactions in this block</p>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="text-gray-400 text-sm text-left">
                <th className="pb-3 pr-4">Hash</th>
                <th className="pb-3 pr-4">Type</th>
                <th className="pb-3 pr-4">From</th>
                <th className="pb-3 pr-4">To</th>
                <th className="pb-3 pr-4">Amount</th>
                <th className="pb-3">Nonce</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-700/50">
              {transactions.map((tx) => (
                <tr key={tx.hash}>
                  <td className="py-3 pr-4">
                    <HashDisplay hash={tx.hash} truncate link={`/tx/${tx.hash}`} />
                  </td>
                  <td className="py-3 pr-4">
                    <TxTypeLabel type={tx.type} />
                  </td>
                  <td className="py-3 pr-4">
                    <AddressLink address={tx.from} />
                  </td>
                  <td className="py-3 pr-4">
                    <AddressLink address={tx.to} />
                  </td>
                  <td className="py-3 pr-4">
                    <Amount value={tx.amount} />
                  </td>
                  <td className="py-3 font-mono">{tx.nonce}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
