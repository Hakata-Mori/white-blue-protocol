import { Link } from 'react-router-dom';
import SearchBar from '../ui/SearchBar';

export default function Header() {
  return (
    <header className="sticky top-0 z-50 bg-gray-900 border-b border-gray-800">
      <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between gap-4 flex-wrap">
        <Link to="/" className="text-xl font-bold text-white whitespace-nowrap">
          White-Blue Explorer
        </Link>
        <nav className="flex items-center gap-6 text-sm">
          <Link to="/" className="text-gray-300 hover:text-white">Home</Link>
          <Link to="/blocks" className="text-gray-300 hover:text-white">Blocks</Link>
          <Link to="/bluecoins" className="text-gray-300 hover:text-white">Blue Coins</Link>
          <Link to="/validators" className="text-gray-300 hover:text-white">Validators</Link>
          <Link to="/faucet" className="text-gray-300 hover:text-white">Faucet</Link>
          <Link to="/wallet" className="text-gray-300 hover:text-white">Wallet</Link>
        </nav>
        <SearchBar />
      </div>
    </header>
  );
}
