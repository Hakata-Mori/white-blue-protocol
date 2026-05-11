import { useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useFetch } from '../hooks/useFetch';
import { getBlueCoins, getBlueCoinState } from '../api/bluecoin';
import { getPool } from '../api/pool';
import { MICRO } from '../config';
import { formatAmount } from '../utils/format';
import type { BlueCoinConfig, BlueCoinState, AMMPool } from '../types';
import AddressLink from '../components/ui/AddressLink';
import Amount from '../components/ui/Amount';
import LoadingSpinner from '../components/ui/LoadingSpinner';

type BlueCoinInfo = { config: BlueCoinConfig; pool?: AMMPool; state?: BlueCoinState };

export default function BlueCoinListPage() {
  const fetcher = useCallback(async (): Promise<BlueCoinInfo[]> => {
    const configs = await getBlueCoins();
    const results = await Promise.all(
      configs.map(async (config) => {
        const [pool, state] = await Promise.all([
          getPool(config.tokenId).catch(() => undefined),
          getBlueCoinState(config.tokenId).catch(() => undefined),
        ]);
        return { config, pool, state } as BlueCoinInfo;
      })
    );
    return results;
  }, []);

  const { data, loading } = useFetch(fetcher);

  if (loading && !data) {
    return <LoadingSpinner />;
  }

  const coins = data ?? [];

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Blue Coins</h1>
      {coins.length === 0 ? (
        <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 text-center text-gray-400">
          No Blue Coins deployed yet
        </div>
      ) : (
        <div className="bg-gray-800 border border-gray-700 rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-700 text-gray-400">
                  <th className="text-left px-4 py-3 font-medium">Name</th>
                  <th className="text-left px-4 py-3 font-medium">Creator</th>
                  <th className="text-right px-4 py-3 font-medium">Price</th>
                  <th className="text-right px-4 py-3 font-medium">Market Cap</th>
                  <th className="text-right px-4 py-3 font-medium">Burned</th>
                  <th className="text-right px-4 py-3 font-medium">Pool Size</th>
                </tr>
              </thead>
              <tbody>
                {coins.map((coin) => {
                  const price =
                    coin.pool && coin.pool.blueReserve > 0
                      ? coin.pool.whiteReserve / coin.pool.blueReserve
                      : 0;
                  const burned = coin.state?.burned ?? 0;
                  const circulating = coin.config.totalSupply - burned;
                  const marketCap = (price * circulating) / MICRO;
                  const burnPercent =
                    coin.config.totalSupply > 0
                      ? ((burned / coin.config.totalSupply) * 100).toFixed(1)
                      : '0.0';

                  return (
                    <tr
                      key={coin.config.tokenId}
                      className="border-b border-gray-700/50 hover:bg-gray-700/30"
                    >
                      <td className="px-4 py-3">
                        <Link
                          to={`/bluecoin/${coin.config.tokenId}`}
                          className="text-blue-400 hover:text-blue-300 font-medium"
                        >
                          {coin.config.name} ({coin.config.symbol})
                        </Link>
                      </td>
                      <td className="px-4 py-3">
                        <AddressLink address={coin.config.creator} />
                      </td>
                      <td className="px-4 py-3 text-right font-mono">
                        {price > 0 ? `${price.toFixed(6)} WC` : '-'}
                      </td>
                      <td className="px-4 py-3 text-right font-mono">
                        {marketCap > 0
                          ? `${marketCap.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} WC`
                          : '-'}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <span className="font-mono">
                          <Amount value={burned} suffix={coin.config.symbol} />
                        </span>
                        <span className="text-gray-400 text-xs ml-1">({burnPercent}%)</span>
                      </td>
                      <td className="px-4 py-3 text-right">
                        {coin.pool ? (
                          <Amount value={coin.pool.whiteReserve} />
                        ) : (
                          <span className="text-gray-500">-</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
