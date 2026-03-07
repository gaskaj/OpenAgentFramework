import { useEffect, useCallback } from 'react';
import { useAgentStore } from '@/store/agent-store';
import { useAuthStore } from '@/store/auth-store';
import type { AgentFilters } from '@/types';

export function useAgents(filters?: AgentFilters) {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { agents, loading, error, total, page, fetchAgents } = useAgentStore();

  const refresh = useCallback(() => {
    if (currentOrg) {
      fetchAgents(currentOrg.slug, filters as unknown as Record<string, unknown>);
    }
  }, [currentOrg, filters, fetchAgents]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { agents, loading, error, total, page, refresh };
}

export function useAgent(agentId: string | undefined) {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { selectedAgent, loading, error, selectAgent, clearSelected } = useAgentStore();

  useEffect(() => {
    if (currentOrg && agentId) {
      selectAgent(currentOrg.slug, agentId);
    }
    return () => clearSelected();
  }, [currentOrg, agentId, selectAgent, clearSelected]);

  return { agent: selectedAgent, loading, error };
}
