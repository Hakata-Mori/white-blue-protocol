import { useCallback, useMemo } from 'react';
import { usePolling } from '../hooks/usePolling';
import { getStats, getBlocks } from '../api/chain';
import { POLL_NORMAL } from '../config';
import type { Transaction } from '../types';
import StatsBar from '../components/dashboard/StatsBar';
import LatestBlocks from '../components/dashboard/LatestBlocks';
import LatestTxs from '../components/dashboard/LatestTxs';
import LoadingSpinner from '../components/ui/LoadingSpinner';

export default function HomePage() {
  const fetchStats = useCallback(() => getStats(), []);
  const fetchBlocks = useCallback(() => getBlocks(10, 0), []);

  const { data: stats, loading: statsLoading } = usePolling(fetchStats, POLL_NORMAL);
  const { data: blocksData, loading: blocksLoading } = usePolling(fetchBlocks, POLL_NORMAL);

  const latestTxs = useMemo(() => {
    if (!blocksData?.blocks) return [];
    const txs: { tx: Transaction; blockHeight: number }[] = [];
    for (const block of blocksData.blocks) {
      const blockDetail = block as unknown as { height: number; transactions?: Transaction[] };
      if (blockDetail.transactions) {
        for (const tx of blockDetail.transactions) {
          txs.push({ tx, blockHeight: blockDetail.height });
        }
      }
    }
    return txs.slice(0, 10);
  }, [blocksData]);

  if (blocksLoading && !blocksData) {
    return <LoadingSpinner />;
  }

  return (
    <div>
      <StatsBar stats={stats} loading={statsLoading} />
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <LatestBlocks blocks={blocksData?.blocks ?? []} />
        <LatestTxs transactions={latestTxs} />
      </div>
    </div>
  );
}
