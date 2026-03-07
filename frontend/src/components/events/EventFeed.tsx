import { useEffect, useRef } from 'react';
import type { AgentEvent } from '@/types';
import { EventRow } from './EventRow';

interface EventFeedProps {
  events: AgentEvent[];
  compact?: boolean;
  maxHeight?: string;
  autoScroll?: boolean;
  emptyMessage?: string;
}

export function EventFeed({
  events,
  compact = false,
  maxHeight = '500px',
  autoScroll = true,
  emptyMessage = 'No events yet.',
}: EventFeedProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const safeEvents = events ?? [];

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = 0;
    }
  }, [safeEvents.length, autoScroll]);

  if (safeEvents.length === 0) {
    return (
      <div className="flex items-center justify-center rounded-lg border border-zinc-700 bg-zinc-800 p-12">
        <p className="text-sm text-zinc-500">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className="overflow-y-auto rounded-lg border border-zinc-700 bg-zinc-800"
      style={{ maxHeight }}
    >
      {safeEvents.map((event) => (
        <EventRow key={event.id} event={event} compact={compact} />
      ))}
    </div>
  );
}
