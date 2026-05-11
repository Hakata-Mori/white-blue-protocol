import { Link } from 'react-router-dom';
import type { Transaction } from '../../types';
import { formatHash, timeAgo } from '../../utils/format';
import TxTypeLabel from '../ui/TxTypeLabel';
import AddressLink from '../ui/AddressLink';
import Amount from '../ui/Amount';

interface LatestTxsProps {
  transactions: { tx: Transaction; blockHeight: number }[];
}

export default function LatestTxs({ transactions }: LatestTxsProps) {
  const display = transactions.slice(0, 10);

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-white">Latest Transactions</h2>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-gray-400 border-b border-gray-700">
              <th className="text-left py-2 pr-3">Hash</th>
              <th className="text-left py-2 pr-3">Type</th>
              <th className="text-left py-2 pr-3">From</th>
              <th className="text-left py-2 pr-3">To</th>
              <th className="text-left py-2 pr-3">Amount</th>
              <th className="text-left py-2">Time</th>
            </tr>
          </thead>
          <tbody>
            {display.map((item) => (
              <tr key={item.tx.hash} className="border-b border-gray-700/50 hover:bg-gray-750">
                <td className="py-2 pr-3">
                  <Link to={`/tx/${item.tx.hash}`} className="text-blue-400 hover:text-blue-300 font-mono">
                    {formatHash(item.tx.hash)}
                  </Link>
                </td>
                <td className="py-2 pr-3">
                  <TxTypeLabel type={item.tx.type} />
                </td>
                <td className="py-2 pr-3">
                  <AddressLink address={item.tx.from} />
                </td>
                <td className="py-2 pr-3">
                  <AddressLink address={item.tx.to} />
                </td>
                <td className="py-2 pr-3">
                  <Amount value={item.tx.amount} />
                </td>
                <td className="py-2 text-gray-400">
                  {timeAgo(item.tx.timestamp)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
