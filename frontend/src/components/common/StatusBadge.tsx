import { cn } from '@/lib/utils';

type StatusType = 'online' | 'offline' | 'error' | 'idle' | 'info' | 'warning' | 'critical' | 'success';

const statusStyles: Record<StatusType, string> = {
  online: 'bg-green-500/20 text-green-400 border-green-500/30',
  success: 'bg-green-500/20 text-green-400 border-green-500/30',
  offline: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  idle: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  error: 'bg-red-500/20 text-red-400 border-red-500/30',
  critical: 'bg-red-500/20 text-red-400 border-red-500/30',
  warning: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  info: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
};

const dotStyles: Record<StatusType, string> = {
  online: 'bg-green-500',
  success: 'bg-green-500',
  offline: 'bg-gray-500',
  idle: 'bg-gray-500',
  error: 'bg-red-500',
  critical: 'bg-red-500',
  warning: 'bg-yellow-500',
  info: 'bg-blue-500',
};

interface StatusBadgeProps {
  status: string;
  className?: string;
  showDot?: boolean;
}

export function StatusBadge({ status, className, showDot = true }: StatusBadgeProps) {
  const normalized = status.toLowerCase() as StatusType;
  const style = statusStyles[normalized] ?? statusStyles.info;
  const dot = dotStyles[normalized] ?? dotStyles.info;

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium',
        style,
        className,
      )}
    >
      {showDot && <span className={cn('h-1.5 w-1.5 rounded-full', dot)} />}
      {status}
    </span>
  );
}
