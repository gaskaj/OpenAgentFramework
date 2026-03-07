import { create } from 'zustand';
import type { Agent } from '@/types';
import * as agentsApi from '@/api/agents';

interface AgentState {
  agents: Agent[];
  selectedAgent: Agent | null;
  loading: boolean;
  error: string | null;
  total: number;
  page: number;

  fetchAgents: (orgSlug: string, filters?: Record<string, unknown>) => Promise<void>;
  selectAgent: (orgSlug: string, agentId: string) => Promise<void>;
  clearSelected: () => void;
}

export const useAgentStore = create<AgentState>()((set) => ({
  agents: [],
  selectedAgent: null,
  loading: false,
  error: null,
  total: 0,
  page: 1,

  fetchAgents: async (orgSlug, filters) => {
    set({ loading: true, error: null });
    try {
      const result = await agentsApi.listAgents(orgSlug, filters);
      set({
        agents: result.data,
        total: result.total,
        page: result.page,
        loading: false,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch agents',
        loading: false,
      });
    }
  },

  selectAgent: async (orgSlug, agentId) => {
    set({ loading: true, error: null });
    try {
      const agent = await agentsApi.getAgent(orgSlug, agentId);
      set({ selectedAgent: agent, loading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch agent',
        loading: false,
      });
    }
  },

  clearSelected: () => set({ selectedAgent: null }),
}));
