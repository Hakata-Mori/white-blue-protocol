import { useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useFetch } from '../hooks/useFetch';
import { getBlueCoin, getBlueCoinState } from '../api/bluecoin';
import { getPool } from '../api/pool';
import type { BlueCoinConfig, BlueCoinState, AMMPool } from '../types';
import AddressLink from '../components/ui/AddressLink';
import Amount from '../components/ui/Amount';
import HashDisplay from '../components/ui/HashDisplay';
import Timestamp from '../components/ui/Timestamp';
import LoadingSpinner from '../components/ui/LoadingSpinner';

type BlueCoinDetail = {
  config: BlueCoinConfig;
  state?: BlueCoinState;
  pool?: AMMPool;
};

export default function BlueCoinDetailPage() {
  const { tokenId } = useParams<{ tokenId: string }>();

  const fetcher = useCallback(async (): Promise<BlueCoinDetail> => {
    const [config, state, pool] = await Promise.all([
      getBlueCoin(tokenId!),
      getBlueCoinState(tokenId!).catch(() => undefined),
      getPool(tokenId!).catch(() => undefined),
    ]);
    return { config, state, pool };
  }, [tokenId]);

  const { data, loading } = useFetch(fetcher);

  if (loading && !data) {
    return <LoadingSpinner />;
  }

  if (!data) {
    return (
      <div className="text-center text-gray-400 py-12">Token not found</div>
    );
  }

  const { config, state, pool } = data;

  const burned = state?.burned ?? 0;
  const teamLocked = state?.teamLocked ?? 0;
  const burnPercent =
    config.totalSupply > 0
      ? ((burned / config.totalSupply) * 100).toFixed(2)
      : '0.00';
  const circulatingSupply = config.totalSupply - burned - teamLocked;

  const spotPrice =
    pool && pool.blueReserve > 0
      ? pool.whiteReserve / pool.blueReserve
      : 0;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">
        {config.name} ({config.symbol})
      </h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
          <h2 className="text-lg font-semibold mb-4">Token Info</h2>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Token ID</span>
            <HashDisplay hash={config.tokenId} truncate={false} />
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Name</span>
            <span>{config.name}</span>
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Symbol</span>
            <span>{config.symbol}</span>
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Creator</span>
            <AddressLink address={config.creator} />
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Total Supply</span>
            <Amount value={config.totalSupply} suffix={config.symbol} />
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Pool / Team Ratio</span>
            <span>
              {(config.poolRatio * 100).toFixed(0)}% / {(config.teamRatio * 100).toFixed(0)}%
            </span>
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Init White</span>
            <Amount value={config.initWhite} />
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Monthly Release</span>
            <Amount value={config.releaseMonthly} suffix={config.symbol} />
          </div>
          <div className="flex justify-between py-3 border-b border-gray-700/50">
            <span className="text-gray-400">Deploy Tx Hash</span>
            <HashDisplay hash={config.deployTxHash} link={`/tx/${config.deployTxHash}`} />
          </div>
          <div className="flex justify-between py-3">
            <span className="text-gray-400">Deployed At</span>
            <Timestamp value={config.deployedAt} />
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
            <h2 className="text-lg font-semibold mb-4">Token State</h2>
            {state ? (
              <>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">Burned</span>
                  <span>
                    <Amount value={burned} suffix={config.symbol} />
                    <span className="text-gray-400 text-xs ml-1">({burnPercent}%)</span>
                  </span>
                </div>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">Team Locked</span>
                  <Amount value={state.teamLocked} suffix={config.symbol} />
                </div>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">Team Released</span>
                  <Amount value={state.teamReleased} suffix={config.symbol} />
                </div>
                <div className="flex justify-between py-3">
                  <span className="text-gray-400">Circulating Supply</span>
                  <Amount value={circulatingSupply} suffix={config.symbol} />
                </div>
              </>
            ) : (
              <div className="text-gray-500 text-center py-4">No state data available</div>
            )}
          </div>

          <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
            <h2 className="text-lg font-semibold mb-4">AMM Pool</h2>
            {pool ? (
              <>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">White Reserve</span>
                  <Amount value={pool.whiteReserve} />
                </div>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">Blue Reserve</span>
                  <Amount value={pool.blueReserve} suffix={config.symbol} />
                </div>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">Spot Price</span>
                  <span className="font-mono">
                    {spotPrice > 0 ? `${spotPrice.toFixed(6)} WC` : '-'}
                  </span>
                </div>
                <div className="flex justify-between py-3 border-b border-gray-700/50">
                  <span className="text-gray-400">Total Fee Burned</span>
                  <Amount value={pool.totalFeeBurned} />
                </div>
                <div className="flex justify-between py-3">
                  <span className="text-gray-400">Last Traded</span>
                  <Timestamp value={pool.lastTradedAt} />
                </div>
              </>
            ) : (
              <div className="text-gray-500 text-center py-4">No pool data available</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
