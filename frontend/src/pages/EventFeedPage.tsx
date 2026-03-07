import { useState, useCallback } from 'react';
import { useEvents } from '@/hooks/useEvents';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useAuthStore } from '@/store/auth-store';
import { useEventStore } from '@/store/event-store';
import { useAgentStore } from '@/store/agent-store';
import { EventFeed } from '@/components/events/EventFeed';
import { EventFilter } from '@/components/events/EventFilter';
import type { EventFilters } from '@/types';

export function EventFeedPage() {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const agents = useAgentStore((s) => s.agents);
  const realtimeEvents = useEventStore((s) => s.realtimeEvents);
  const [filters, setFilters] = useState<EventFilters>({ page: 1, per_page: 50 });
  const { events, loading, total, page, refresh } = useEvents(filters);

  useWebSocket(currentOrg?.slug);

  const handleFilterChange = useCallback((newFilters: EventFilters) => {
    setFilters(newFilters);
  }, []);

  const totalPages = Math.ceil((total || 0) / (filters.per_page || 50));

  const agentOptions = agents.map((a) => ({ id: a.id, name: a.name }));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Event Feed</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Real-time and historical events from your agent fleet
        </p>
      </div>

      {/* Real-time events */}
      {realtimeEvents.length > 0 && (
        <div>
          <h2 className="mb-2 flex items-center gap-2 text-sm font-medium text-zinc-300">
            <span className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
            Live Events
          </h2>
          <EventFeed events={realtimeEvents} compact maxHeight="300px" />
        </div>
      )}

      {/* Filters */}
      <EventFilter
        filters={filters}
        onFilterChange={handleFilterChange}
        agentOptions={agentOptions}
      />

      {/* Historical events */}
      <div>
        <div className="mb-2 flex items-center justify-between">
          <h2 className="text-sm font-medium text-zinc-300">
            Historical Events
            {total ? (
              <span className="ml-2 text-zinc-500">({total} total)</span>
            ) : null}
          </h2>
          <button
            onClick={refresh}
            className="text-xs text-blue-400 transition-colors hover:text-blue-300"
          >
            Refresh
          </button>
        </div>

        {loading ? (
          <div className="rounded-lg border border-zinc-700 bg-zinc-800">
            <div className="space-y-3 p-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-12 animate-pulse rounded bg-zinc-700" />
              ))}
            </div>
          </div>
        ) : (
          <EventFeed events={events} maxHeight="600px" emptyMessage="No events match your filters." />
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="mt-4 flex items-center justify-between">
            <span className="text-xs text-zinc-500">
              Page {page} of {totalPages}
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => handleFilterChange({ ...filters, page: (page || 1) - 1 })}
                disabled={(page || 1) <= 1}
                className="rounded-lg border border-zinc-700 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:bg-zinc-800 disabled:opacity-30"
              >
                Previous
              </button>
              <button
                onClick={() => handleFilterChange({ ...filters, page: (page || 1) + 1 })}
                disabled={(page || 1) >= totalPages}
                className="rounded-lg border border-zinc-700 px-3 py-1.5 text-xs text-zinc-400 transition-colors hover:bg-zinc-800 disabled:opacity-30"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
