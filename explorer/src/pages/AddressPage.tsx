import { useCallback, useState } from 'react';
import { useParams, useSearchParams, Link } from 'react-router-dom';
import { useFetch } from '../hooks/useFetch';
import { getWallet, getAddressTxs } from '../api/wallet';
import { TXS_PER_PAGE } from '../config';
import HashDisplay from '../components/ui/HashDisplay';
import AddressLink from '../components/ui/AddressLink';
import Amount from '../components/ui/Amount';
import TxTypeLabel from '../components/ui/TxTypeLabel';
import Badge from '../components/ui/Badge';
import Timestamp from '../components/ui/Timestamp';
import Pagination from '../components/ui/Pagination';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import { formatHash } from '../utils/format';

export default function AddressPage() {
  const { address } = useParams<{ address: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Math.max(1, parseInt(searchParams.get('page') ?? '1', 10));
  const offset = (page - 1) * TXS_PER_PAGE;
  const [showAll, setShowAll] = useState(false);

  const walletFetcher = useCallback(
    () => getWallet(address!),
    [address]
  );

  const txsFetcher = useCallback(
    () => getAddressTxs(address!, TXS_PER_PAGE, offset, showAll),
    [address, offset, showAll]
  );

  const { data: account, loading, error } = useFetch(walletFetcher);
  const { data: txsData, loading: txsLoading } = useFetch(txsFetcher);

  const handlePageChange = (newPage: number) => {
    setSearchParams({ page: String(newPage) });
  };

  const handleToggleAll = () => {
    setShowAll(!showAll);
    setSearchParams({ page: '1' });
  };

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
  const txs = txsData?.txs ?? [];
  const totalTxs = txsData?.total ?? 0;

  const statusBadge = (status: string) => {
    switch (status) {
      case 'success':
        return <Badge label={status} className="bg-green-900 text-green-300" />;
      case 'failed':
        return <Badge label={status} className="bg-red-900 text-red-300" />;
      case 'pending':
        return <Badge label={status} className="bg-yellow-900 text-yellow-300" />;
      default:
        return <Badge label={status} />;
    }
  };

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

      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold">Transaction History</h2>
          <div className="flex items-center gap-4">
            <button
              onClick={handleToggleAll}
              className={`px-3 py-1 rounded text-sm ${
                showAll
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-800 text-gray-400 border border-gray-700 hover:text-gray-200'
              }`}
            >
              {showAll ? 'Show All' : 'Hide System Txs'}
            </button>
            <span className="text-gray-400 text-sm">Total: {totalTxs.toLocaleString()}</span>
          </div>
        </div>

        {txsLoading && !txsData ? (
          <LoadingSpinner />
        ) : txs.length === 0 ? (
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 text-center">
            <span className="text-gray-400">No transactions found</span>
          </div>
        ) : (
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-gray-400 text-sm text-left">
                  <th className="pb-3 pr-4">Hash</th>
                  <th className="pb-3 pr-4">Type</th>
                  <th className="pb-3 pr-4">From</th>
                  <th className="pb-3 pr-4">To</th>
                  <th className="pb-3 pr-4">Amount</th>
                  <th className="pb-3 pr-4">Block</th>
                  <th className="pb-3 pr-4">Time</th>
                  <th className="pb-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700/50">
                {txs.map((tx) => (
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
                    <td className="py-3 pr-4">
                      <Link
                        to={`/block/${tx.blockHeight}`}
                        className="text-blue-400 hover:text-blue-300 font-mono"
                      >
                        {tx.blockHeight.toLocaleString()}
                      </Link>
                    </td>
                    <td className="py-3 pr-4">
                      <Timestamp value={tx.timestamp} />
                    </td>
                    <td className="py-3">
                      {statusBadge(tx.status)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <Pagination
          current={page}
          total={totalTxs}
          perPage={TXS_PER_PAGE}
          onChange={handlePageChange}
        />
      </div>
    </div>
  );
}
