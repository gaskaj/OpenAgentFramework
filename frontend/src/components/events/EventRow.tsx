import {
  Activity,
  AlertTriangle,
  AlertOctagon,
  Info,
  CheckCircle2,
  XCircle,
  GitPullRequest,
  Bot,
} from 'lucide-react';
import type { AgentEvent, Severity } from '@/types';
import { StatusBadge } from '@/components/common/StatusBadge';
import { TimeAgo } from '@/components/common/TimeAgo';
import { cn } from '@/lib/utils';

interface EventRowProps {
  event: AgentEvent;
  compact?: boolean;
}

const severityIcon: Record<Severity, typeof Info> = {
  info: Info,
  warning: AlertTriangle,
  error: XCircle,
  critical: AlertOctagon,
};

function getEventIcon(eventType: string) {
  if (eventType.startsWith('agent.')) return Bot;
  if (eventType.includes('pr_created')) return GitPullRequest;
  if (eventType.includes('completed')) return CheckCircle2;
  if (eventType.includes('failed')) return XCircle;
  return Activity;
}

export function EventRow({ event, compact = false }: EventRowProps) {
  const SeverityIcon = severityIcon[event.severity] ?? Info;
  const EventIcon = getEventIcon(event.event_type);

  return (
    <div
      className={cn(
        'flex items-start gap-3 border-b border-zinc-700/50 px-4 py-3 transition-colors hover:bg-zinc-800/50',
        compact && 'py-2',
      )}
    >
      <div className="mt-0.5 flex-shrink-0">
        <EventIcon className="h-4 w-4 text-zinc-500" />
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-zinc-200">
            {event.event_type}
          </span>
          <StatusBadge status={event.severity} showDot={false} />
        </div>

        <p className="mt-0.5 text-sm text-zinc-400">{event.message}</p>

        {!compact && (
          <div className="mt-1 flex items-center gap-3 text-xs text-zinc-500">
            <span className="flex items-center gap-1">
              <Bot className="h-3 w-3" />
              {event.agent_name}
            </span>
            <TimeAgo date={event.created_at} />
          </div>
        )}
      </div>

      {compact && (
        <TimeAgo date={event.created_at} className="flex-shrink-0 text-xs text-zinc-600" />
      )}

      <SeverityIcon className="mt-0.5 h-4 w-4 flex-shrink-0 text-zinc-600" />
    </div>
  );
}
