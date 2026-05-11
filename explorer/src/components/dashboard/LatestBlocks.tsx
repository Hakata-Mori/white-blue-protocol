import { Link } from 'react-router-dom';
import type { BlockSummary } from '../../types';
import { formatHash, timeAgo } from '../../utils/format';
import AddressLink from '../ui/AddressLink';

interface LatestBlocksProps {
  blocks: BlockSummary[];
}

export default function LatestBlocks({ blocks }: LatestBlocksProps) {
  const display = blocks.slice(0, 10);

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-xl p-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-white">Latest Blocks</h2>
        <Link to="/blocks" className="text-sm text-blue-400 hover:text-blue-300">
          View all
        </Link>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-gray-400 border-b border-gray-700">
              <th className="text-left py-2 pr-3">Height</th>
              <th className="text-left py-2 pr-3">Hash</th>
              <th className="text-left py-2 pr-3">Validator</th>
              <th className="text-left py-2 pr-3">Txs</th>
              <th className="text-left py-2">Time</th>
            </tr>
          </thead>
          <tbody>
            {display.map((block) => (
              <tr key={block.height} className="border-b border-gray-700/50 hover:bg-gray-750">
                <td className="py-2 pr-3">
                  <Link to={`/block/${block.height}`} className="text-blue-400 hover:text-blue-300 font-mono">
                    {block.height}
                  </Link>
                </td>
                <td className="py-2 pr-3 font-mono text-gray-300">
                  {formatHash(block.hash)}
                </td>
                <td className="py-2 pr-3">
                  <AddressLink address={block.validator} />
                </td>
                <td className="py-2 pr-3 font-mono text-gray-300">
                  {block.txCount}
                </td>
                <td className="py-2 text-gray-400">
                  {timeAgo(block.timestamp)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
