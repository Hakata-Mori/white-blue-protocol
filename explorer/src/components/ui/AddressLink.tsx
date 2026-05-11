import { Link } from 'react-router-dom';
import { formatHash } from '../../utils/format';

interface AddressLinkProps {
  address: string;
  truncate?: boolean;
}

export default function AddressLink({ address, truncate = true }: AddressLinkProps) {
  const display = truncate ? formatHash(address) : address;
  return (
    <Link
      to={`/address/${address}`}
      className="font-mono text-blue-400 hover:text-blue-300"
    >
      {display}
    </Link>
  );
}
