import { useState, useMemo } from 'react';
import { Search, LayoutGrid, List } from 'lucide-react';
import * as Tabs from '@radix-ui/react-tabs';
import { useAgents } from '@/hooks/useAgents';
import { AgentCard } from '@/components/agents/AgentCard';
import { AgentTable } from '@/components/agents/AgentTable';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import type { AgentStatus } from '@/types';
import { cn } from '@/lib/utils';

type ViewMode = 'grid' | 'table';

const statusTabs: { value: string; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'online', label: 'Online' },
  { value: 'offline', label: 'Offline' },
  { value: 'error', label: 'Error' },
];

export function AgentListPage() {
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const { agents, loading } = useAgents();

  const filteredAgents = useMemo(() => {
    let result = agents;
    if (statusFilter !== 'all') {
      result = result.filter((a) => a.status === (statusFilter as AgentStatus));
    }
    if (search) {
      const q = search.toLowerCase();
      result = result.filter(
        (a) =>
          a.name.toLowerCase().includes(q) ||
          a.agent_type.toLowerCase().includes(q) ||
          a.github_repo?.toLowerCase().includes(q),
      );
    }
    return result;
  }, [agents, statusFilter, search]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">Agents</h1>
          <p className="mt-1 text-sm text-zinc-400">
            Manage and monitor your agent fleet
          </p>
        </div>

        {/* View toggle */}
        <div className="flex items-center gap-1 rounded-lg border border-zinc-700 bg-zinc-800 p-1">
          <button
            onClick={() => setViewMode('grid')}
            className={cn(
              'rounded-md p-1.5 transition-colors',
              viewMode === 'grid'
                ? 'bg-zinc-700 text-zinc-100'
                : 'text-zinc-500 hover:text-zinc-300',
            )}
          >
            <LayoutGrid className="h-4 w-4" />
          </button>
          <button
            onClick={() => setViewMode('table')}
            className={cn(
              'rounded-md p-1.5 transition-colors',
              viewMode === 'table'
                ? 'bg-zinc-700 text-zinc-100'
                : 'text-zinc-500 hover:text-zinc-300',
            )}
          >
            <List className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Search and filters */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-500" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search agents..."
            className="w-full rounded-lg border border-zinc-700 bg-zinc-800 py-2 pl-9 pr-3 text-sm text-zinc-300 placeholder-zinc-600 outline-none focus:border-blue-500"
          />
        </div>

        <Tabs.Root value={statusFilter} onValueChange={setStatusFilter}>
          <Tabs.List className="flex rounded-lg border border-zinc-700 bg-zinc-800 p-1">
            {statusTabs.map((tab) => (
              <Tabs.Trigger
                key={tab.value}
                value={tab.value}
                className={cn(
                  'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                  statusFilter === tab.value
                    ? 'bg-zinc-700 text-zinc-100'
                    : 'text-zinc-500 hover:text-zinc-300',
                )}
              >
                {tab.label}
              </Tabs.Trigger>
            ))}
          </Tabs.List>
        </Tabs.Root>
      </div>

      {/* Agent list */}
      {loading ? (
        <LoadingSpinner size="lg" className="py-20" />
      ) : viewMode === 'grid' ? (
        filteredAgents.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-lg border border-zinc-700 bg-zinc-800 py-20">
            <p className="text-sm text-zinc-500">No agents found.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {filteredAgents.map((agent) => (
              <AgentCard key={agent.id} agent={agent} />
            ))}
          </div>
        )
      ) : (
        <AgentTable agents={filteredAgents} />
      )}
    </div>
  );
}
