const TX_TYPES: Record<number, { label: string; className: string }> = {
  1: { label: 'Transfer', className: 'bg-blue-900 text-blue-300' },
  2: { label: 'Blue Transfer', className: 'bg-indigo-900 text-indigo-300' },
  3: { label: 'Deploy Token', className: 'bg-green-900 text-green-300' },
  4: { label: 'Buy Blue', className: 'bg-purple-900 text-purple-300' },
  5: { label: 'Sell Blue', className: 'bg-orange-900 text-orange-300' },
  7: { label: 'Block Reward', className: 'bg-yellow-900 text-yellow-300' },
  8: { label: 'Heartbeat', className: 'bg-gray-800 text-gray-400' },
  9: { label: 'Validator Join', className: 'bg-teal-900 text-teal-300' },
  10: { label: 'Validator Exit', className: 'bg-red-900 text-red-300' },
  11: { label: 'Validator Evict', className: 'bg-red-900 text-red-300' },
  13: { label: 'Slash Evidence', className: 'bg-rose-900 text-rose-300' },
  14: { label: 'Burn Blue', className: 'bg-amber-900 text-amber-300' },
  20: { label: 'Multisig Register', className: 'bg-cyan-900 text-cyan-300' },
  21: { label: 'Multisig Propose', className: 'bg-cyan-900 text-cyan-300' },
  22: { label: 'Multisig Approve', className: 'bg-cyan-900 text-cyan-300' },
};

export function getTxTypeLabel(type: number): { label: string; className: string } {
  return TX_TYPES[type] ?? { label: `Type ${type}`, className: 'bg-gray-700 text-gray-300' };
}
