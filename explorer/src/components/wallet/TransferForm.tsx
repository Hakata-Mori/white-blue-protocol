import { useState } from 'react';
import { Link } from 'react-router-dom';
import { calcFee, signTransactionWithKey, submitTransaction } from '../../lib/wallet';
import type { TransactionInput } from '../../lib/wallet';
import { formatAmount } from '../../utils/format';

interface TransferFormProps {
  from: string;
  publicKey: string;
  privateKey: string;
  currentNonce: number;
}

export default function TransferForm({ from, publicKey, privateKey, currentNonce }: TransferFormProps) {
  const [to, setTo] = useState('');
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [successHash, setSuccessHash] = useState('');

  const amountMicro = amount ? Math.floor(parseFloat(amount) * 1_000_000) : 0;
  const fee = amountMicro > 0 ? calcFee(amountMicro) : 0;

  const handleSend = async () => {
    setError('');
    setSuccessHash('');
    if (!to || !amount) {
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
        type: 1,
        from,
        to,
        amount: amountMicro,
        tokenId: '',
        fee: 0,
        nonce: currentNonce + 1,
        payload: null,
        publicKey,
        timestamp: Math.floor(Date.now() / 1000),
      };
      const signed = signTransactionWithKey(tx, privateKey);
      const result = await submitTransaction(signed);
      setSuccessHash(result.hash);
      setTo('');
      setAmount('');
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-xl p-6">
      <h2 className="text-lg font-bold mb-4">Transfer White Coin</h2>
      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">To Address</label>
          <input
            type="text"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className="w-full bg-gray-700 border border-gray-600 rounded-lg px-4 py-2 text-gray-100 focus:outline-none focus:border-blue-500"
            placeholder="0x..."
          />
        </div>
        <div>
          <label className="block text-sm text-gray-400 mb-1">Amount (WC)</label>
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
        {fee > 0 && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Estimated Fee</span>
            <span className="font-mono text-gray-300">{formatAmount(fee)} WC</span>
          </div>
        )}
        {error && <p className="text-red-400 text-sm">{error}</p>}
        {successHash && (
          <div className="p-3 bg-green-900/30 border border-green-700 rounded-lg">
            <p className="text-green-400 text-sm">Transaction sent!</p>
            <Link
              to={`/tx/${successHash}`}
              className="text-blue-400 hover:text-blue-300 text-sm font-mono break-all"
            >
              {successHash}
            </Link>
          </div>
        )}
        <button
          onClick={handleSend}
          disabled={loading || !to || !amount}
          className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Sending...' : 'Send'}
        </button>
      </div>
    </div>
  );
}
