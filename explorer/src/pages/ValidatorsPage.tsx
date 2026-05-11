import { useCallback } from 'react';
import { Link } from 'react-router-dom';
import { usePolling } from '../hooks/usePolling';
import { getValidators } from '../api/validators';
import { POLL_SLOW } from '../config';
import type { ValidatorSet } from '../types';
import AddressLink from '../components/ui/AddressLink';
import Badge from '../components/ui/Badge';
import LoadingSpinner from '../components/ui/LoadingSpinner';

const STATUS_COLORS: Record<string, string> = {
  active: 'bg-emerald-900 text-emerald-300',
  candidate: 'bg-blue-900 text-blue-300',
  suspended: 'bg-yellow-900 text-yellow-300',
  removed: 'bg-gray-700 text-gray-400',
  slashed: 'bg-red-900 text-red-300',
};

function getUptimeDisplay(
  status: string,
  lastHeartbeatHeight: number,
  currentHeight: number
) {
  if (status === 'active') {
    if (currentHeight - lastHeartbeatHeight <= 100) {
      return { label: 'Online', className: 'text-emerald-400' };
    }
    return { label: 'Online', className: 'text-emerald-400' };
  }
  if (status === 'suspended') {
    return { label: 'Suspended', className: 'text-yellow-400' };
  }
  return { label: 'Offline', className: 'text-red-400' };
}

export default function ValidatorsPage() {
  const fetcher = useCallback(() => getValidators(), []);
  const { data, loading } = usePolling<ValidatorSet>(fetcher, POLL_SLOW);

  if (loading && !data) {
    return <LoadingSpinner />;
  }

  const validators = data?.validators ?? [];
  const currentHeight = data?.updatedAt ?? 0;
  const activeCount = validators.filter((v) => v.status === 'active').length;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">
        Validators ({activeCount} active)
      </h1>

      <div className="bg-gray-800 border border-gray-700 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700 text-gray-400">
                <th className="text-left px-4 py-3 font-medium">Address</th>
                <th className="text-left px-4 py-3 font-medium">Status</th>
                <th className="text-right px-4 py-3 font-medium">Join Height</th>
                <th className="text-right px-4 py-3 font-medium">Last Heartbeat</th>
                <th className="text-left px-4 py-3 font-medium">Uptime</th>
              </tr>
            </thead>
            <tbody>
              {validators.map((v) => {
                const uptime = getUptimeDisplay(
                  v.status,
                  v.lastHeartbeatHeight,
                  currentHeight
                );
                return (
                  <tr
                    key={v.address}
                    className="border-b border-gray-700/50 hover:bg-gray-700/30"
                  >
                    <td className="px-4 py-3">
                      <AddressLink address={v.address} />
                    </td>
                    <td className="px-4 py-3">
                      <Badge
                        label={v.status}
                        className={STATUS_COLORS[v.status] ?? 'bg-gray-700 text-gray-300'}
                      />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link
                        to={`/block/${v.joinHeight}`}
                        className="text-blue-400 hover:text-blue-300 font-mono"
                      >
                        {v.joinHeight.toLocaleString()}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-right font-mono">
                      {v.lastHeartbeatHeight.toLocaleString()}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`font-medium ${uptime.className}`}>
                        {uptime.label}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
