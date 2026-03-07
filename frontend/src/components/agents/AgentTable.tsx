import { Link } from 'react-router-dom';
import { DataTable, type Column } from '@/components/common/DataTable';
import { StatusBadge } from '@/components/common/StatusBadge';
import { TimeAgo } from '@/components/common/TimeAgo';
import type { Agent } from '@/types';

interface AgentTableProps {
  agents: Agent[];
  loading?: boolean;
  page?: number;
  totalPages?: number;
  onPageChange?: (page: number) => void;
}

const columns: Column<Agent>[] = [
  {
    key: 'name',
    header: 'Name',
    sortable: true,
    render: (agent) => (
      <Link
        to={`/agents/${agent.id}`}
        className="font-medium text-blue-400 hover:text-blue-300"
      >
        {agent.name}
      </Link>
    ),
  },
  {
    key: 'type',
    header: 'Type',
    render: (agent) => <span className="text-zinc-400">{agent.agent_type}</span>,
  },
  {
    key: 'status',
    header: 'Status',
    sortable: true,
    render: (agent) => <StatusBadge status={agent.status} />,
  },
  {
    key: 'version',
    header: 'Version',
    render: (agent) => <span className="font-mono text-xs text-zinc-500">{agent.version}</span>,
  },
  {
    key: 'github_repo',
    header: 'Repository',
    render: (agent) => (
      <span className="truncate text-zinc-400">{agent.github_repo || '--'}</span>
    ),
  },
  {
    key: 'last_heartbeat',
    header: 'Last Heartbeat',
    sortable: true,
    render: (agent) =>
      agent.last_heartbeat ? (
        <TimeAgo date={agent.last_heartbeat} className="text-zinc-500" />
      ) : (
        <span className="text-zinc-600">--</span>
      ),
  },
];

export function AgentTable({
  agents,
  loading,
  page,
  totalPages,
  onPageChange,
}: AgentTableProps) {
  return (
    <DataTable
      columns={columns}
      data={agents}
      loading={loading}
      rowKey={(a) => a.id}
      emptyMessage="No agents registered."
      page={page}
      totalPages={totalPages}
      onPageChange={onPageChange}
    />
  );
}
