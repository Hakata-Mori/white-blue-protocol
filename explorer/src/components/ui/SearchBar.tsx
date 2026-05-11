import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { detectSearch } from '../../utils/search';

export default function SearchBar() {
  const [query, setQuery] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const navigate = useNavigate();

  const handleSearch = () => {
    if (!query.trim()) return;
    const result = detectSearch(query);
    switch (result.type) {
      case 'block_height':
        navigate(`/block/${result.value}`);
        setQuery('');
        setErrorMsg('');
        break;
      case 'address':
        navigate(`/address/${result.value}`);
        setQuery('');
        setErrorMsg('');
        break;
      case 'hash':
        navigate(`/tx/${result.value}`);
        setQuery('');
        setErrorMsg('');
        break;
      default:
        setErrorMsg('Invalid search query');
        setTimeout(() => setErrorMsg(''), 2000);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSearch();
  };

  return (
    <div className="relative">
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Search block height / tx hash / address..."
        className="w-full md:w-80 px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500 text-sm"
      />
      {errorMsg && (
        <div className="absolute top-full mt-1 left-0 text-xs text-red-400 bg-gray-800 border border-red-800 rounded px-2 py-1">
          {errorMsg}
        </div>
      )}
    </div>
  );
}
