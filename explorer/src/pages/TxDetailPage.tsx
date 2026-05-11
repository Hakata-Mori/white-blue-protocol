import { useCallback, useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { usePolling } from '../hooks/usePolling';
import { useFetch } from '../hooks/useFetch';
import { getTx } from '../api/transaction';
import { getBlock } from '../api/chain';
import HashDisplay from '../components/ui/HashDisplay';
import AddressLink from '../components/ui/AddressLink';
import Amount from '../components/ui/Amount';
import TxTypeLabel from '../components/ui/TxTypeLabel';
import Badge from '../components/ui/Badge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import { formatTimestamp } from '../utils/format';
import type { TxReceipt, Transaction } from '../types';

function StatusBanner({ receipt }: { receipt: TxReceipt }) {
  if (receipt.status === 'pending') {
    return (
      <div className="bg-yellow-900/50 border border-yellow-700 rounded-xl p-4 mb-6 flex items-center gap-3">
        <span className="relative flex h-3 w-3">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-yellow-400 opacity-75" />
          <span className="relative inline-flex rounded-full h-3 w-3 bg-yellow-500" />
        </span>
        <span className="text-yellow-300 font-medium">Transaction is pending confirmation</span>
      </div>
    );
  }
  if (receipt.status === 'success') {
    return (
      <div className="bg-green-900/50 border border-green-700 rounded-xl p-4 mb-6 flex items-center gap-3">
        <span className="h-3 w-3 rounded-full bg-green-500" />
        <span className="text-green-300 font-medium">Transaction succeeded</span>
      </div>
    );
  }
  return (
    <div className="bg-red-900/50 border border-red-700 rounded-xl p-4 mb-6">
      <div className="flex items-center gap-3">
        <span className="h-3 w-3 rounded-full bg-red-500" />
        <span className="text-red-300 font-medium">Transaction failed</span>
      </div>
      {receipt.error && (
        <p className="text-red-400 text-sm mt-2 ml-6">{receipt.error}</p>
      )}
    </div>
  );
}

function statusBadge(status: string): { label: string; className: string } {
  switch (status) {
    case 'success':
      return { label: 'Success', className: 'bg-green-900 text-green-300' };
    case 'pending':
      return { label: 'Pending', className: 'bg-yellow-900 text-yellow-300' };
    case 'failed':
      return { label: 'Failed', className: 'bg-red-900 text-red-300' };
    default:
      return { label: status, className: 'bg-gray-700 text-gray-300' };
  }
}

function TxDetails({ tx }: { tx: Transaction }) {
  return (
    <>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">Type</span>
        <TxTypeLabel type={tx.type} />
      </div>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">From</span>
        <AddressLink address={tx.from} truncate={false} />
      </div>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">To</span>
        <AddressLink address={tx.to} truncate={false} />
      </div>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">Amount</span>
        <Amount value={tx.amount} />
      </div>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">Fee</span>
        <Amount value={tx.fee} />
      </div>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">Nonce</span>
        <span className="font-mono">{tx.nonce}</span>
      </div>
      <div className="flex justify-between py-3 border-b border-gray-700/50">
        <span className="text-gray-400">Timestamp</span>
        <span className="text-gray-300">{formatTimestamp(tx.timestamp)}</span>
      </div>
      {tx.tokenId && (
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Token ID</span>
          <Link
            to={`/bluecoin/${tx.tokenId}`}
            className="font-mono text-blue-400 hover:text-blue-300"
          >
            {tx.tokenId}
          </Link>
        </div>
      )}
      {tx.minAmountOut !== undefined && (
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Min Amount Out</span>
          <Amount value={tx.minAmountOut} />
        </div>
      )}
    </>
  );
}

export default function TxDetailPage() {
  const { hash } = useParams<{ hash: string }>();

  const receiptFetcher = useCallback(
    () => getTx(hash!),
    [hash]
  );

  const { data: initialReceipt, loading: receiptLoading, error: receiptError } = useFetch(receiptFetcher);

  const isPending = initialReceipt?.status === 'pending';

  const pollFetcher = useCallback(
    () => getTx(hash!),
    [hash]
  );

  const { data: polledReceipt } = usePolling(pollFetcher, isPending ? 3000 : 0);

  const receipt = useMemo(() => {
    if (isPending && polledReceipt) return polledReceipt;
    return initialReceipt;
  }, [initialReceipt, polledReceipt, isPending]);

  const blockFetcher = useCallback(() => {
    if (!receipt || receipt.blockHeight <= 0) {
      return Promise.resolve(null);
    }
    return getBlock(receipt.blockHeight);
  }, [receipt]);

  const { data: block } = useFetch(blockFetcher);

  const tx = useMemo(() => {
    if (!block || !receipt) return null;
    return block.transactions.find((t) => t.hash === receipt.txHash) ?? null;
  }, [block, receipt]);

  if (receiptLoading && !initialReceipt) {
    return <LoadingSpinner />;
  }

  if (receiptError || !receipt) {
    return (
      <div className="text-center py-12">
        <h1 className="text-2xl font-bold text-red-400 mb-2">Transaction Not Found</h1>
        <p className="text-gray-400">No transaction found with hash {hash}</p>
        <Link to="/" className="text-blue-400 hover:text-blue-300 mt-4 inline-block">
          Back to home
        </Link>
      </div>
    );
  }

  const { label: statusLabel, className: statusClassName } = statusBadge(receipt.status);

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Transaction Details</h1>

      <StatusBanner receipt={receipt} />

      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 mb-6">
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Hash</span>
          <HashDisplay hash={receipt.txHash} truncate={false} />
        </div>
        <div className="flex justify-between py-3 border-b border-gray-700/50">
          <span className="text-gray-400">Status</span>
          <Badge label={statusLabel} className={statusClassName} />
        </div>
        {tx && <TxDetails tx={tx} />}
      </div>

      {receipt.blockHeight > 0 && (
        <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
          <h2 className="text-lg font-bold mb-4">Receipt</h2>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Block Height</span>
            <Link
              to={`/block/${receipt.blockHeight}`}
              className="text-blue-400 hover:text-blue-300 font-mono"
            >
              {receipt.blockHeight.toLocaleString()}
            </Link>
          </div>
          <div className="flex justify-between py-3">
            <span className="text-gray-400">Block Hash</span>
            <HashDisplay
              hash={receipt.blockHash}
              truncate
              link={`/block/hash/${receipt.blockHash}`}
            />
          </div>
        </div>
      )}
    </div>
  );
}
