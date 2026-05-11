import { useState } from 'react';
import { Link } from 'react-router-dom';
import { calcFee, signTransactionWithKey, submitTransaction } from '../../lib/wallet';
import type { TransactionInput } from '../../lib/wallet';
import { formatAmount } from '../../utils/format';

interface DeployFormProps {
  from: string;
  publicKey: string;
  privateKey: string;
  currentNonce: number;
}

export default function DeployForm({ from, publicKey, privateKey, currentNonce }: DeployFormProps) {
  const [name, setName] = useState('');
  const [symbol, setSymbol] = useState('');
  const [poolRatio, setPoolRatio] = useState('70');
  const [initWhite, setInitWhite] = useState('');
  const [releaseMonthly, setReleaseMonthly] = useState('');
  const [multiSigAddr, setMultiSigAddr] = useState('');
  const [sourceUrls, setSourceUrls] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [successHash, setSuccessHash] = useState('');

  const poolRatioNum = parseInt(poolRatio) || 0;
  const teamRatio = 100 - poolRatioNum;
  const initWhiteMicro = initWhite ? Math.floor(parseFloat(initWhite) * 1_000_000) : 0;
  const releaseMicro = releaseMonthly ? Math.floor(parseFloat(releaseMonthly) * 1_000_000) : 0;
  const fee = initWhiteMicro > 0 ? calcFee(initWhiteMicro) : 0;

  const handleDeploy = async () => {
    setError('');
    setSuccessHash('');
    if (!name || !symbol || !poolRatio || !initWhite || !releaseMonthly) {
      setError('Please fill all fields');
      return;
    }
    if (poolRatioNum < 1 || poolRatioNum > 99) {
      setError('Pool ratio must be between 1 and 99');
      return;
    }
    if (initWhiteMicro <= 0) {
      setError('Init White must be greater than 0');
      return;
    }
    if (releaseMicro <= 0) {
      setError('Monthly Release must be greater than 0');
      return;
    }

    setLoading(true);
    try {
      const urls = sourceUrls.trim() ? sourceUrls.split(',').map(s => s.trim()).filter(Boolean) : [];
      const deployParams = {
        name,
        symbol,
        poolRatio: poolRatioNum,
        teamRatio,
        initWhite: initWhiteMicro,
        releaseMonthly: releaseMicro,
        multiSigAddr: multiSigAddr.trim(),
        sourceUrls: urls,
      };
      const payloadBase64 = btoa(JSON.stringify(deployParams));

      const tx: TransactionInput = {
        type: 3,
        from,
        to: '',
        amount: 0,
        tokenId: '',
        fee: 0,
        nonce: currentNonce + 1,
        payload: payloadBase64,
        publicKey,
        timestamp: Math.floor(Date.now() / 1000),
      };
      const signed = signTransactionWithKey(tx, privateKey);
      const result = await submitTransaction(signed);
      setSuccessHash(result.hash);
      setName('');
      setSymbol('');
      setPoolRatio('70');
      setInitWhite('');
      setReleaseMonthly('');
      setMultiSigAddr('');
      setSourceUrls('');
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
      <h2 className="text-lg font-bold mb-4">Deploy Blue Coin</h2>
      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="My Blue Coin"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Symbol</label>
          <input
            type="text"
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="MBC"
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Pool Ratio (%)</label>
            <input
              type="number"
              value={poolRatio}
              onChange={(e) => setPoolRatio(e.target.value)}
              className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
              min="1"
              max="99"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-1">Team Ratio (%)</label>
            <div className="w-full bg-gray-700/50 border border-gray-600 rounded-lg px-4 py-2 text-gray-400">
              {teamRatio}
            </div>
          </div>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Init White (WC)</label>
          <input
            type="number"
            value={initWhite}
            onChange={(e) => setInitWhite(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="0.00"
            step="0.000001"
            min="0"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Monthly Release (tokens)</label>
          <input
            type="number"
            value={releaseMonthly}
            onChange={(e) => setReleaseMonthly(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="0.00"
            step="0.000001"
            min="0"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">MultiSig Address (optional)</label>
          <input
            type="text"
            value={multiSigAddr}
            onChange={(e) => setMultiSigAddr(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500 font-mono"
            placeholder="0x... (team fund managed by multisig)"
          />
          <p className="text-xs text-gray-500 mt-1">Team tokens will be released to this multisig wallet monthly</p>
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Source URLs (optional, comma separated)</label>
          <input
            type="text"
            value={sourceUrls}
            onChange={(e) => setSourceUrls(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="https://example.com, https://twitter.com/..."
          />
          <p className="text-xs text-gray-500 mt-1">Links to the project website, social media, etc.</p>
        </div>
        {fee > 0 && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Estimated Fee</span>
            <span className="font-mono text-gray-300">{formatAmount(fee)} WC</span>
          </div>
        )}
        {error && <p className="text-red-400 text-sm">{error}</p>}
        {successHash && (
          <div className="p-3 bg-green-900/30 border border-green-700 rounded-lg">
            <p className="text-green-400 text-sm">Blue coin deployed!</p>
            <Link
              to={`/tx/${successHash}`}
              className="text-blue-400 hover:text-blue-300 text-sm font-mono break-all"
            >
              {successHash}
            </Link>
          </div>
        )}
        <button
          onClick={handleDeploy}
          disabled={loading || !name || !symbol || !initWhite || !releaseMonthly}
          className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Deploying...' : 'Deploy'}
        </button>
      </div>
    </div>
  );
}
