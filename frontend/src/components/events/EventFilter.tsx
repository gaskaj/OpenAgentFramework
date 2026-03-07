import { useState } from 'react';
import { Search, Filter, X } from 'lucide-react';
import type { EventFilters, EventType, Severity } from '@/types';

interface EventFilterProps {
  filters: EventFilters;
  onFilterChange: (filters: EventFilters) => void;
  agentOptions?: { id: string; name: string }[];
}

const eventTypes: EventType[] = [
  'agent.registered',
  'agent.deregistered',
  'agent.heartbeat',
  'agent.status_change',
  'agent.error',
  'issue.claimed',
  'issue.analyzed',
  'issue.decomposed',
  'issue.implemented',
  'issue.pr_created',
  'issue.completed',
  'issue.failed',
  'workflow.started',
  'workflow.step_completed',
  'workflow.completed',
  'workflow.failed',
];

const severities: Severity[] = ['info', 'warning', 'error', 'critical'];

export function EventFilter({
  filters,
  onFilterChange,
  agentOptions = [],
}: EventFilterProps) {
  const [search, setSearch] = useState(filters.search ?? '');

  const updateFilter = (key: keyof EventFilters, value: string | undefined) => {
    onFilterChange({ ...filters, [key]: value || undefined, page: 1 });
  };

  const handleSearch = () => {
    updateFilter('search', search || undefined);
  };

  const clearFilters = () => {
    setSearch('');
    onFilterChange({ page: 1, per_page: filters.per_page });
  };

  const hasActiveFilters =
    filters.event_type || filters.severity || filters.agent_id || filters.search;

  return (
    <div className="space-y-3 rounded-lg border border-zinc-700 bg-zinc-800 p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm font-medium text-zinc-300">
          <Filter className="h-4 w-4" />
          Filters
        </div>
        {hasActiveFilters && (
          <button
            onClick={clearFilters}
            className="flex items-center gap-1 text-xs text-zinc-500 hover:text-zinc-300"
          >
            <X className="h-3 w-3" />
            Clear
          </button>
        )}
      </div>

      {/* Search */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-500" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Search events..."
            className="w-full rounded-lg border border-zinc-700 bg-zinc-900 py-2 pl-9 pr-3 text-sm text-zinc-300 placeholder-zinc-600 outline-none focus:border-blue-500"
          />
        </div>
        <button
          onClick={handleSearch}
          className="rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          Search
        </button>
      </div>

      {/* Dropdowns */}
      <div className="flex flex-wrap gap-3">
        <select
          value={filters.event_type ?? ''}
          onChange={(e) => updateFilter('event_type', e.target.value)}
          className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 outline-none focus:border-blue-500"
        >
          <option value="">All event types</option>
          {eventTypes.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>

        <select
          value={filters.severity ?? ''}
          onChange={(e) => updateFilter('severity', e.target.value)}
          className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 outline-none focus:border-blue-500"
        >
          <option value="">All severities</option>
          {severities.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>

        {agentOptions.length > 0 && (
          <select
            value={filters.agent_id ?? ''}
            onChange={(e) => updateFilter('agent_id', e.target.value)}
            className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 outline-none focus:border-blue-500"
          >
            <option value="">All agents</option>
            {agentOptions.map((a) => (
              <option key={a.id} value={a.id}>
                {a.name}
              </option>
            ))}
          </select>
        )}

        <input
          type="date"
          value={filters.from ?? ''}
          onChange={(e) => updateFilter('from', e.target.value)}
          className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 outline-none focus:border-blue-500"
        />
        <input
          type="date"
          value={filters.to ?? ''}
          onChange={(e) => updateFilter('to', e.target.value)}
          className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-300 outline-none focus:border-blue-500"
        />
      </div>
    </div>
  );
}
