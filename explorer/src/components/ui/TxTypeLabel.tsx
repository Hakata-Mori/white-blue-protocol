import { getTxTypeLabel } from '../../utils/txTypes';
import Badge from './Badge';

interface TxTypeLabelProps {
  type: number;
}

export default function TxTypeLabel({ type }: TxTypeLabelProps) {
  const { label, className } = getTxTypeLabel(type);
  return <Badge label={label} className={className} />;
}
