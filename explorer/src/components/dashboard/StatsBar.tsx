import type { NetworkStats } from '../../types';
import { formatAmount } from '../../utils/format';

interface StatsBarProps {
  stats: NetworkStats | null;
  loading: boolean;
}

export default function StatsBar({ stats, loading }: StatsBarProps) {
  if (loading || !stats) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        {[0, 1, 2, 3].map((i) => (
          <div key={i} className="bg-gray-800 border border-gray-700 rounded-xl p-4 animate-pulse">
            <div className="h-4 bg-gray-700 rounded w-24 mb-2" />
            <div className="h-6 bg-gray-700 rounded w-16" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
        <div className="text-sm text-gray-400 mb-1">Block Height</div>
        <div className="flex items-center gap-2">
          <span className="relative flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500" />
          </span>
          <span className="text-xl font-bold font-mono">{stats.height.toLocaleString()}</span>
        </div>
      </div>
      <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
        <div className="text-sm text-gray-400 mb-1">Total Minted</div>
        <div className="text-xl font-bold font-mono">{formatAmount(stats.totalMinted)} WC</div>
      </div>
      <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
        <div className="text-sm text-gray-400 mb-1">Validators</div>
        <div className="text-xl font-bold font-mono">{stats.activeValidators}</div>
      </div>
      <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
        <div className="text-sm text-gray-400 mb-1">Avg Block Time</div>
        <div className="text-xl font-bold font-mono">{stats.avgBlockTime.toFixed(1)}s</div>
      </div>
    </div>
  );
}
