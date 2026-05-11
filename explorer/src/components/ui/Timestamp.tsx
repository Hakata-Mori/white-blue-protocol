import { timeAgo, formatTimestamp } from '../../utils/format';

interface TimestampProps {
  value: number;
}

export default function Timestamp({ value }: TimestampProps) {
  return (
    <span title={formatTimestamp(value)} className="text-gray-400">
      {timeAgo(value)}
    </span>
  );
}
