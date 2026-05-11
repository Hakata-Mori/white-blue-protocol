import { useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { usePolling } from '../hooks/usePolling';
import { getBlocks } from '../api/chain';
import { BLOCKS_PER_PAGE } from '../config';
import HashDisplay from '../components/ui/HashDisplay';
import AddressLink from '../components/ui/AddressLink';
import Amount from '../components/ui/Amount';
import Timestamp from '../components/ui/Timestamp';
import Badge from '../components/ui/Badge';
import Pagination from '../components/ui/Pagination';
import LoadingSpinner from '../components/ui/LoadingSpinner';

export default function BlocksPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Math.max(1, parseInt(searchParams.get('page') ?? '1', 10));
  const offset = (page - 1) * BLOCKS_PER_PAGE;

  const fetcher = useCallback(
    () => getBlocks(BLOCKS_PER_PAGE, offset),
    [offset]
  );

  const interval = page === 1 ? 10_000 : 0;
  const { data, loading, error } = usePolling(fetcher, interval);

  const handlePageChange = (newPage: number) => {
    setSearchParams({ page: String(newPage) });
  };

  if (loading && !data) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="text-center text-red-400 py-12">
        Failed to load blocks: {error}
      </div>
    );
  }

  const blocks = data?.blocks ?? [];
  const total = data?.total ?? 0;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Blocks</h1>
        <span className="text-gray-400 text-sm">Total: {total.toLocaleString()}</span>
      </div>

      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="text-gray-400 text-sm text-left">
              <th className="pb-3 pr-4">Height</th>
              <th className="pb-3 pr-4">Hash</th>
              <th className="pb-3 pr-4">Validator</th>
              <th className="pb-3 pr-4">Txs</th>
              <th className="pb-3 pr-4">Reward</th>
              <th className="pb-3">Time</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700/50">
            {blocks.map((block) => (
              <tr key={block.height}>
                <td className="py-3 pr-4">
                  <Link
                    to={`/block/${block.height}`}
                    className="text-blue-400 hover:text-blue-300 font-mono"
                  >
                    {block.height.toLocaleString()}
                  </Link>
                </td>
                <td className="py-3 pr-4">
                  <HashDisplay hash={block.hash} truncate link={`/block/${block.height}`} />
                </td>
                <td className="py-3 pr-4">
                  <AddressLink address={block.validator} />
                </td>
                <td className="py-3 pr-4">
                  <Badge label={String(block.txCount)} />
                </td>
                <td className="py-3 pr-4">
                  <Amount value={block.reward} />
                </td>
                <td className="py-3">
                  <Timestamp value={block.timestamp} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Pagination
        current={page}
        total={total}
        perPage={BLOCKS_PER_PAGE}
        onChange={handlePageChange}
      />
    </div>
  );
}
