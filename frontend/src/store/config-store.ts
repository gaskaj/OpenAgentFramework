import { create } from 'zustand';
import type { AgentTypeConfig, AgentConfigOverride } from '@/types';
import * as configApi from '@/api/config';

interface ConfigState {
  typeConfigs: AgentTypeConfig[];
  selectedTypeConfig: AgentTypeConfig | null;
  agentOverride: AgentConfigOverride | null;
  mergedPreview: Record<string, unknown> | null;
  mergedVersion: number;
  loading: boolean;
  saving: boolean;
  error: string | null;

  fetchTypeConfigs: (orgSlug: string) => Promise<void>;
  fetchTypeConfig: (orgSlug: string, agentType: string) => Promise<void>;
  saveTypeConfig: (orgSlug: string, agentType: string, config: Record<string, unknown>, description?: string) => Promise<void>;
  fetchAgentOverride: (orgSlug: string, agentId: string) => Promise<void>;
  saveAgentOverride: (orgSlug: string, agentId: string, config: Record<string, unknown>, description?: string) => Promise<void>;
  deleteAgentOverride: (orgSlug: string, agentId: string) => Promise<void>;
  fetchMergedPreview: (orgSlug: string, agentId: string) => Promise<void>;
  clearError: () => void;
}

export const useConfigStore = create<ConfigState>()((set) => ({
  typeConfigs: [],
  selectedTypeConfig: null,
  agentOverride: null,
  mergedPreview: null,
  mergedVersion: 0,
  loading: false,
  saving: false,
  error: null,

  fetchTypeConfigs: async (orgSlug) => {
    set({ loading: true, error: null });
    try {
      const configs = await configApi.listAgentTypeConfigs(orgSlug);
      set({ typeConfigs: configs, loading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to fetch configs', loading: false });
    }
  },

  fetchTypeConfig: async (orgSlug, agentType) => {
    set({ loading: true, error: null });
    try {
      const config = await configApi.getAgentTypeConfig(orgSlug, agentType);
      set({ selectedTypeConfig: config, loading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to fetch config', loading: false });
    }
  },

  saveTypeConfig: async (orgSlug, agentType, config, description) => {
    set({ saving: true, error: null });
    try {
      const result = await configApi.upsertAgentTypeConfig(orgSlug, agentType, config, description);
      set((state) => ({
        selectedTypeConfig: result,
        typeConfigs: state.typeConfigs.map((c) =>
          c.agent_type === agentType ? result : c,
        ),
        saving: false,
      }));
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to save config', saving: false });
    }
  },

  fetchAgentOverride: async (orgSlug, agentId) => {
    set({ loading: true, error: null });
    try {
      const override = await configApi.getAgentOverride(orgSlug, agentId);
      set({ agentOverride: override, loading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to fetch override', loading: false });
    }
  },

  saveAgentOverride: async (orgSlug, agentId, config, description) => {
    set({ saving: true, error: null });
    try {
      const result = await configApi.upsertAgentOverride(orgSlug, agentId, config, description);
      set({ agentOverride: result, saving: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to save override', saving: false });
    }
  },

  deleteAgentOverride: async (orgSlug, agentId) => {
    set({ saving: true, error: null });
    try {
      await configApi.deleteAgentOverride(orgSlug, agentId);
      set({ agentOverride: null, saving: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to delete override', saving: false });
    }
  },

  fetchMergedPreview: async (orgSlug, agentId) => {
    set({ loading: true, error: null });
    try {
      const result = await configApi.getMergedConfig(orgSlug, agentId);
      set({ mergedPreview: result.config, mergedVersion: result.version, loading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to fetch merged config', loading: false });
    }
  },

  clearError: () => set({ error: null }),
}));
