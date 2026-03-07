import { Link } from 'react-router-dom';
import { Bot, GitBranch, Clock } from 'lucide-react';
import type { Agent } from '@/types';
import { StatusBadge } from '@/components/common/StatusBadge';
import { TimeAgo } from '@/components/common/TimeAgo';

interface AgentCardProps {
  agent: Agent;
}

export function AgentCard({ agent }: AgentCardProps) {
  return (
    <Link
      to={`/agents/${agent.id}`}
      className="block rounded-lg border border-zinc-700 bg-zinc-800 p-5 transition-colors hover:border-zinc-600 hover:bg-zinc-800/80"
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-600/20">
            <Bot className="h-5 w-5 text-blue-400" />
          </div>
          <div>
            <h3 className="font-medium text-zinc-100">{agent.name}</h3>
            <p className="text-xs text-zinc-500">{agent.agent_type}</p>
          </div>
        </div>
        <StatusBadge status={agent.status} />
      </div>

      <div className="mt-4 space-y-2">
        {agent.github_repo && (
          <div className="flex items-center gap-2 text-xs text-zinc-400">
            <GitBranch className="h-3.5 w-3.5" />
            <span className="truncate">{agent.github_repo}</span>
          </div>
        )}
        {agent.last_heartbeat && (
          <div className="flex items-center gap-2 text-xs text-zinc-500">
            <Clock className="h-3.5 w-3.5" />
            <span>
              Last heartbeat: <TimeAgo date={agent.last_heartbeat} />
            </span>
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
    </Link>
  );
}
