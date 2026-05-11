import { useState } from 'react';
import { Link } from 'react-router-dom';
import { formatHash } from '../../utils/format';

interface HashDisplayProps {
  hash: string;
  truncate?: boolean;
  link?: string;
}

export default function HashDisplay({ hash, truncate = true, link }: HashDisplayProps) {
  const [copied, setCopied] = useState(false);

  const display = truncate ? formatHash(hash) : hash;

  const handleCopy = () => {
    navigator.clipboard.writeText(hash);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <span className="inline-flex items-center gap-1 font-mono">
      {link ? (
        <Link to={link} className="text-blue-400 hover:text-blue-300">
          {display}
        </Link>
      ) : (
        <span>{display}</span>
      )}
      <button
        onClick={handleCopy}
        className="text-gray-500 hover:text-gray-300 cursor-pointer"
        title="Copy to clipboard"
      >
        {copied ? (
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-green-400" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
        ) : (
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
            <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
          </svg>
        )}
      </button>
    </span>
  );
}
