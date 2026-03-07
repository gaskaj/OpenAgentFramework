import apiClient from './client';
import type { Agent, AgentFilters, PaginatedResponse } from '@/types';

export async function listAgents(
  orgSlug: string,
  filters?: AgentFilters,
): Promise<PaginatedResponse<Agent>> {
  const { data } = await apiClient.get<PaginatedResponse<Agent>>(
    `/orgs/${orgSlug}/agents`,
    { params: filters },
  );
  return { ...data, data: data.data ?? [] };
}

export async function getAgent(orgSlug: string, agentId: string): Promise<Agent> {
  const { data } = await apiClient.get<{ data: Agent }>(`/orgs/${orgSlug}/agents/${agentId}`);
  return data.data;
}

export async function registerAgent(
  orgSlug: string,
  agentData: Partial<Agent>,
): Promise<Agent> {
  const { data } = await apiClient.post<{ data: Agent }>(`/orgs/${orgSlug}/agents`, agentData);
  return data.data;
}

export async function updateAgent(
  orgSlug: string,
  agentId: string,
  agentData: Partial<Agent>,
): Promise<Agent> {
  const { data } = await apiClient.patch<{ data: Agent }>(
    `/orgs/${orgSlug}/agents/${agentId}`,
    agentData,
  );
  return data.data;
}

export async function deleteAgent(orgSlug: string, agentId: string): Promise<void> {
  await apiClient.delete(`/orgs/${orgSlug}/agents/${agentId}`);
}
