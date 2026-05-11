interface PaginationProps {
  current: number;
  total: number;
  perPage: number;
  onChange: (page: number) => void;
}

export default function Pagination({ current, total, perPage, onChange }: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / perPage));

  return (
    <div className="flex items-center justify-center gap-4 mt-4">
      <button
        onClick={() => onChange(current - 1)}
        disabled={current <= 1}
        className="px-3 py-1 rounded bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Prev
      </button>
      <span className="text-sm text-gray-400">
        Page {current} of {totalPages}
      </span>
      <button
        onClick={() => onChange(current + 1)}
        disabled={current >= totalPages}
        className="px-3 py-1 rounded bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Next
      </button>
    </div>
  );
}
