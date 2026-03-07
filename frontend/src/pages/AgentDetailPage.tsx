import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Trash2, Bot, Server, Clock, GitBranch, Tag } from 'lucide-react';
import { useAgent } from '@/hooks/useAgents';
import { useAuthStore } from '@/store/auth-store';
import { deleteAgent } from '@/api/agents';
import { getAgentEvents } from '@/api/events';
import { StatusBadge } from '@/components/common/StatusBadge';
import { TimeAgo } from '@/components/common/TimeAgo';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { EventFeed } from '@/components/events/EventFeed';
import type { AgentEvent } from '@/types';
import { useEffect } from 'react';

export function AgentDetailPage() {
  const { agentId } = useParams<{ agentId: string }>();
  const navigate = useNavigate();
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { agent, loading } = useAgent(agentId);
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (currentOrg && agentId) {
      getAgentEvents(currentOrg.slug, agentId, { per_page: 50 })
        .then((res) => setEvents(res.data))
        .catch(() => {});
    }
  }, [currentOrg, agentId]);

  const handleDelete = async () => {
    if (!currentOrg || !agentId) return;
    if (!window.confirm('Are you sure you want to deregister this agent?')) return;
    setDeleting(true);
    try {
      await deleteAgent(currentOrg.slug, agentId);
      navigate('/agents');
    } catch {
      setDeleting(false);
    }
  };

  if (loading) {
    return <LoadingSpinner size="lg" className="py-20" />;
  }

  if (!agent) {
    return (
      <div className="py-20 text-center text-sm text-zinc-500">Agent not found.</div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Back button */}
      <button
        onClick={() => navigate('/agents')}
        className="flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-200"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to agents
      </button>

      {/* Agent header */}
      <div className="flex items-start justify-between rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <div className="flex items-start gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-blue-600/20">
            <Bot className="h-7 w-7 text-blue-400" />
          </div>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold text-zinc-100">{agent.name}</h1>
              <StatusBadge status={agent.status} />
            </div>
            <p className="mt-1 text-sm text-zinc-400">{agent.agent_type}</p>

            <div className="mt-4 grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
              <div className="flex items-center gap-2 text-zinc-400">
                <Tag className="h-4 w-4 text-zinc-600" />
                <span>Version: </span>
                <span className="font-mono text-zinc-300">{agent.version}</span>
              </div>
              <div className="flex items-center gap-2 text-zinc-400">
                <Server className="h-4 w-4 text-zinc-600" />
                <span>Host: </span>
                <span className="text-zinc-300">{agent.hostname}</span>
              </div>
              {agent.github_repo && (
                <div className="flex items-center gap-2 text-zinc-400">
                  <GitBranch className="h-4 w-4 text-zinc-600" />
                  <span>Repo: </span>
                  <span className="text-zinc-300">{agent.github_repo}</span>
                </div>
              )}
              {agent.last_heartbeat && (
                <div className="flex items-center gap-2 text-zinc-400">
                  <Clock className="h-4 w-4 text-zinc-600" />
                  <span>Last heartbeat: </span>
                  <TimeAgo date={agent.last_heartbeat} className="text-zinc-300" />
                </div>
              )}
            </div>

            {agent.tags && agent.tags.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-1">
                {agent.tags.map((tag) => (
                  <span
                    key={tag}
                    className="rounded-md bg-zinc-700 px-2 py-0.5 text-xs text-zinc-400"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        <button
          onClick={handleDelete}
          disabled={deleting}
          className="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-400 transition-colors hover:bg-red-500/20 disabled:opacity-50"
        >
          <Trash2 className="h-4 w-4" />
          Deregister
        </button>
      </div>

      {/* Config snapshot */}
      {agent.config_snapshot && Object.keys(agent.config_snapshot).length > 0 && (
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-5">
          <h2 className="mb-3 text-sm font-medium text-zinc-300">Configuration Snapshot</h2>
          <pre className="overflow-x-auto rounded-lg bg-zinc-900 p-4 text-xs text-zinc-400">
            {JSON.stringify(agent.config_snapshot, null, 2)}
          </pre>
        </div>
      )}

      {/* Event timeline */}
      <div>
        <h2 className="mb-3 text-sm font-medium text-zinc-300">Event Timeline</h2>
        <EventFeed events={events} maxHeight="600px" emptyMessage="No events for this agent." />
      </div>
    </div>
  );
}
