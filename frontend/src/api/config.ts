import apiClient from './client';
import type { AgentTypeConfig, AgentConfigOverride, ConfigAuditEntry, PaginatedResponse } from '@/types';

export async function listAgentTypeConfigs(orgSlug: string): Promise<AgentTypeConfig[]> {
  const { data } = await apiClient.get<{ data: AgentTypeConfig[] }>(
    `/orgs/${orgSlug}/config/types`,
  );
  return data.data ?? [];
}

export async function getAgentTypeConfig(
  orgSlug: string,
  agentType: string,
): Promise<AgentTypeConfig> {
  const { data } = await apiClient.get<{ data: AgentTypeConfig }>(
    `/orgs/${orgSlug}/config/types/${agentType}`,
  );
  return data.data;
}

export async function upsertAgentTypeConfig(
  orgSlug: string,
  agentType: string,
  config: Record<string, unknown>,
  description?: string,
): Promise<AgentTypeConfig> {
  const { data } = await apiClient.put<{ data: AgentTypeConfig }>(
    `/orgs/${orgSlug}/config/types/${agentType}`,
    { config, description },
  );
  return data.data;
}

export async function getAgentOverride(
  orgSlug: string,
  agentId: string,
): Promise<AgentConfigOverride> {
  const { data } = await apiClient.get<{ data: AgentConfigOverride }>(
    `/orgs/${orgSlug}/config/agents/${agentId}`,
  );
  return data.data;
}

export async function upsertAgentOverride(
  orgSlug: string,
  agentId: string,
  config: Record<string, unknown>,
  description?: string,
): Promise<AgentConfigOverride> {
  const { data } = await apiClient.put<{ data: AgentConfigOverride }>(
    `/orgs/${orgSlug}/config/agents/${agentId}`,
    { config, description },
  );
  return data.data;
}

export async function deleteAgentOverride(
  orgSlug: string,
  agentId: string,
): Promise<void> {
  await apiClient.delete(`/orgs/${orgSlug}/config/agents/${agentId}`);
}

export async function getMergedConfig(
  orgSlug: string,
  agentId: string,
): Promise<{ config: Record<string, unknown>; version: number }> {
  const { data } = await apiClient.get<{ data: { config: Record<string, unknown>; version: number } }>(
    `/orgs/${orgSlug}/config/agents/${agentId}/merged`,
  );
  return data.data;
}

export async function getConfigAudit(
  orgSlug: string,
  params?: { limit?: number; offset?: number },
): Promise<PaginatedResponse<ConfigAuditEntry>> {
  const { data } = await apiClient.get<PaginatedResponse<ConfigAuditEntry>>(
    `/orgs/${orgSlug}/config/audit`,
    { params },
  );
  return { ...data, data: data.data ?? [] };
}
