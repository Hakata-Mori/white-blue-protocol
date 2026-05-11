import { useState } from 'react';
import { Link } from 'react-router-dom';
import { requestFaucet } from '../api/faucet';

export default function FaucetPage() {
  const [address, setAddress] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ status?: string; hash?: string; amount?: number; error?: string; retryAfter?: number; retryAfterH?: string } | null>(null);

  const handleClaim = async () => {
    if (!address.trim()) return;
    setLoading(true);
    setResult(null);
    try {
      const res = await requestFaucet(address.trim());
      setResult(res);
    } catch {
      setResult({ error: 'Network error. Please try again.' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-8">
      <div className="bg-gradient-to-r from-blue-900 to-indigo-900 rounded-xl p-8">
        <h1 className="text-2xl font-bold text-white mb-2">Testnet Early Adopter Program</h1>
        <p className="text-gray-300 mb-4">
          Claim free test WC and explore the White &amp; Blue Protocol. All testnet participants will receive mainnet rewards when we launch.
        </p>
        <ul className="space-y-2 text-gray-200">
          <li className="flex items-start gap-2">
            <span className="text-green-400 font-bold">&#10003;</span>
            <span>Claim 100 test WC every 24 hours</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-green-400 font-bold">&#10003;</span>
            <span>Deploy your own Blue Coin token</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-green-400 font-bold">&#10003;</span>
            <span>Trade on the built-in AMM</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-green-400 font-bold">&#10003;</span>
            <span>Early adopters get priority mainnet rewards</span>
          </li>
        </ul>
      </div>

      <div className="max-w-md mx-auto bg-gray-800 border border-gray-700 rounded-xl p-6">
        <h2 className="text-xl font-bold text-white mb-4">Claim Test WC</h2>
        <input
          type="text"
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          placeholder="0x..."
          className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 mb-4"
        />
        <button
          onClick={handleClaim}
          disabled={loading || !address.trim()}
          className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-semibold py-3 rounded-lg transition-colors"
        >
          {loading ? 'Claiming...' : 'Claim 100 WC'}
        </button>

        {result && result.error && (
          <p className="mt-4 text-red-400 text-sm">{result.error}</p>
        )}

        {result && result.status === 'ok' && result.hash && (
          <div className="mt-4 bg-green-900/40 border border-green-700 rounded-lg p-4">
            <p className="text-green-400 font-semibold mb-1">100 WC sent!</p>
            <p className="text-sm text-gray-300">
              Tx:{' '}
              <Link to={`/tx/${result.hash}`} className="text-blue-400 hover:underline font-mono break-all">
                {result.hash}
              </Link>
            </p>
          </div>
        )}

        {result && result.retryAfter !== undefined && result.retryAfter > 0 && (
          <div className="mt-4 bg-yellow-900/40 border border-yellow-700 rounded-lg p-4">
            <p className="text-yellow-400 text-sm">
              Next claim available in: {result.retryAfterH || `${Math.ceil(result.retryAfter / 3600)}h`}
            </p>
          </div>
        )}
      </div>

      <div className="max-w-2xl mx-auto bg-gray-800 border border-gray-700 rounded-xl p-6">
        <h2 className="text-xl font-bold text-white mb-4">How to Participate</h2>
        <ol className="space-y-4">
          <li className="flex items-start gap-3">
            <span className="flex-shrink-0 w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white font-bold text-sm">1</span>
            <div>
              <p className="text-white font-medium">Create a wallet</p>
              <Link to="/wallet" className="text-blue-400 hover:underline text-sm">Go to Wallet</Link>
            </div>
          </li>
          <li className="flex items-start gap-3">
            <span className="flex-shrink-0 w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white font-bold text-sm">2</span>
            <div>
              <p className="text-white font-medium">Claim test WC on this page</p>
            </div>
          </li>
          <li className="flex items-start gap-3">
            <span className="flex-shrink-0 w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white font-bold text-sm">3</span>
            <div>
              <p className="text-white font-medium">Deploy your own Blue Coin</p>
              <Link to="/wallet" className="text-blue-400 hover:underline text-sm">Go to Wallet (Deploy tab)</Link>
            </div>
          </li>
          <li className="flex items-start gap-3">
            <span className="flex-shrink-0 w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white font-bold text-sm">4</span>
            <div>
              <p className="text-white font-medium">Trade on the AMM</p>
              <Link to="/wallet" className="text-blue-400 hover:underline text-sm">Go to Wallet (Swap tab)</Link>
            </div>
          </li>
        </ol>
      </div>

      <div className="max-w-2xl mx-auto bg-gray-800 border border-gray-700 rounded-xl p-6">
        <h2 className="text-xl font-bold text-white mb-4">Mainnet Rewards</h2>
        <ul className="space-y-2 text-gray-300">
          <li className="flex items-start gap-2">
            <span className="text-blue-400 mt-1">&#8226;</span>
            <span>Airdrop for active testnet users</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-blue-400 mt-1">&#8226;</span>
            <span>Bonus rewards for validators</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-blue-400 mt-1">&#8226;</span>
            <span>Bonus rewards for Blue Coin creators</span>
          </li>
          <li className="flex items-start gap-2">
            <span className="text-blue-400 mt-1">&#8226;</span>
            <span>Details to be announced</span>
          </li>
        </ul>
      </div>
    </div>
  );
}
