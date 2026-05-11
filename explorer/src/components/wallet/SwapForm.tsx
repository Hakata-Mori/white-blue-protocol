import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { signTransactionWithKey, submitTransaction } from '../../lib/wallet';
import type { TransactionInput } from '../../lib/wallet';
import { useFetch } from '../../hooks/useFetch';
import { getBlueCoins } from '../../api/bluecoin';
import { getPool } from '../../api/pool';
import { formatAmount } from '../../utils/format';
import LoadingSpinner from '../ui/LoadingSpinner';
import type { AMMPool } from '../../types';

interface SwapFormProps {
  from: string;
  publicKey: string;
  privateKey: string;
  currentNonce: number;
}

export default function SwapForm({ from, publicKey, privateKey, currentNonce }: SwapFormProps) {
  const fetchCoins = useCallback(() => getBlueCoins(), []);
  const { data: coins, loading: coinsLoading } = useFetch(fetchCoins);

  const [selectedToken, setSelectedToken] = useState('');
  const [direction, setDirection] = useState<'buy' | 'sell'>('buy');
  const [amount, setAmount] = useState('');
  const [slippage, setSlippage] = useState('1');
  const [pool, setPool] = useState<AMMPool | null>(null);
  const [poolLoading, setPoolLoading] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [successHash, setSuccessHash] = useState('');

  useEffect(() => {
    if (coins && coins.length > 0 && !selectedToken) {
      setSelectedToken(coins[0].tokenId);
    }
  }, [coins, selectedToken]);

  useEffect(() => {
    if (!selectedToken) return;
    setPoolLoading(true);
    getPool(selectedToken)
      .then(setPool)
      .catch(() => setPool(null))
      .finally(() => setPoolLoading(false));
  }, [selectedToken]);

  const amountMicro = amount ? Math.floor(parseFloat(amount) * 1_000_000) : 0;

  let estimatedOutput = 0;
  if (pool && amountMicro > 0) {
    if (direction === 'buy') {
      estimatedOutput = Math.floor(amountMicro * pool.blueReserve / (pool.whiteReserve + amountMicro));
    } else {
      estimatedOutput = Math.floor(amountMicro * pool.whiteReserve / (pool.blueReserve + amountMicro));
    }
  }

  const slippagePct = parseFloat(slippage) || 1;
  const minAmountOut = Math.floor(estimatedOutput * (1 - slippagePct / 100));

  const handleSwap = async () => {
    setError('');
    setSuccessHash('');
    if (!selectedToken || !amount) {
      setError('Please fill all fields');
      return;
    }
    if (amountMicro <= 0) {
      setError('Amount must be greater than 0');
      return;
    }

    setLoading(true);
    try {
      const tx: TransactionInput = {
        type: direction === 'buy' ? 4 : 5,
        from,
        to: '',
        amount: amountMicro,
        tokenId: selectedToken,
        fee: 0,
        nonce: currentNonce + 1,
        payload: null,
        publicKey,
        timestamp: Math.floor(Date.now() / 1000),
        minAmountOut,
      };
      const signed = signTransactionWithKey(tx, privateKey);
      const result = await submitTransaction(signed);
      setSuccessHash(result.hash);
      setAmount('');
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  if (coinsLoading) {
    return <LoadingSpinner />;
  }

  if (!coins || coins.length === 0) {
    return (
      <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
        <h2 className="text-lg font-bold mb-4">Swap</h2>
        <p className="text-gray-400">No blue coins available for swap.</p>
      </div>
    );
  }

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
      <h2 className="text-lg font-bold mb-4">Swap</h2>
      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Token</label>
          <select
            value={selectedToken}
            onChange={(e) => setSelectedToken(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
          >
            {coins.map((c) => (
              <option key={c.tokenId} value={c.tokenId}>{c.symbol} ({c.name})</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Direction</label>
          <div className="flex gap-2">
            <button
              onClick={() => setDirection('buy')}
              className={`flex-1 py-2 px-4 rounded-lg font-medium transition-colors ${
                direction === 'buy'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-400 hover:text-gray-200'
              }`}
            >
              Buy Blue
            </button>
            <button
              onClick={() => setDirection('sell')}
              className={`flex-1 py-2 px-4 rounded-lg font-medium transition-colors ${
                direction === 'sell'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-700 text-gray-400 hover:text-gray-200'
              }`}
            >
              Sell Blue
            </button>
          </div>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">
            Amount ({direction === 'buy' ? 'WC' : 'BC'})
          </label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="0.00"
            step="0.000001"
            min="0"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Slippage (%)</label>
          <input
            type="number"
            value={slippage}
            onChange={(e) => setSlippage(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="1"
            step="0.1"
            min="0.1"
            max="50"
          />
        </div>
        {poolLoading && <p className="text-gray-400 text-sm">Loading pool data...</p>}
        {pool && estimatedOutput > 0 && (
          <div className="bg-gray-700/50 rounded-lg p-3 space-y-1">
            <div className="flex justify-between text-sm">
              <span className="text-gray-400">Estimated Output</span>
              <span className="font-mono text-gray-300">
                {formatAmount(estimatedOutput)} {direction === 'buy' ? 'BC' : 'WC'}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-400">Min Amount Out</span>
              <span className="font-mono text-gray-300">
                {formatAmount(minAmountOut)} {direction === 'buy' ? 'BC' : 'WC'}
              </span>
            </div>
          </div>
        )}
        {error && <p className="text-red-400 text-sm">{error}</p>}
        {successHash && (
          <div className="p-3 bg-green-900/30 border border-green-700 rounded-lg">
            <p className="text-green-400 text-sm">Swap submitted!</p>
            <Link
              to={`/tx/${successHash}`}
              className="text-blue-400 hover:text-blue-300 text-sm font-mono break-all"
            >
              {successHash}
            </Link>
          </div>
        )}
        <button
          onClick={handleSwap}
          disabled={loading || !selectedToken || !amount}
          className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Swapping...' : 'Swap'}
        </button>
      </div>
    </div>
  );
}
