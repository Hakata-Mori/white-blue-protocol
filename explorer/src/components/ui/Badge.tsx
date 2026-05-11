interface BadgeProps {
  label: string;
  className?: string;
}

export default function Badge({ label, className = 'bg-gray-700 text-gray-300' }: BadgeProps) {
  return (
    <span className={`inline-block px-2 py-0.5 text-xs font-medium rounded-full ${className}`}>
      {label}
    </span>
  );
}
