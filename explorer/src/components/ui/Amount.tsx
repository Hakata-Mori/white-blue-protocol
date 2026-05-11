import { formatAmount } from '../../utils/format';

interface AmountProps {
  value: number;
  suffix?: string;
}

export default function Amount({ value, suffix = 'WC' }: AmountProps) {
  return (
    <span className="font-mono">
      {formatAmount(value)} {suffix}
    </span>
  );
}
