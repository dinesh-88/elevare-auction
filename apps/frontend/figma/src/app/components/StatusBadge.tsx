interface StatusBadgeProps {
  status: 'LIVE' | 'ENDED' | 'RESERVE_MET' | 'UPCOMING';
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const styles = {
    LIVE: 'bg-red-50 text-red-700 border-red-200',
    ENDED: 'bg-gray-100 text-gray-600 border-gray-200',
    RESERVE_MET: 'bg-green-50 text-green-700 border-green-200',
    UPCOMING: 'bg-blue-50 text-blue-700 border-blue-200',
  };

  const labels = {
    LIVE: 'LIVE',
    ENDED: 'ENDED',
    RESERVE_MET: 'Reserve Met ✓',
    UPCOMING: 'UPCOMING',
  };

  return (
    <span className={`inline-flex items-center px-3 py-1 rounded-md text-xs font-medium border ${styles[status]}`}>
      {labels[status]}
    </span>
  );
}
