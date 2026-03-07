import apiClient from './client';
import type { AgentEvent, EventFilters, EventStats, PaginatedResponse } from '@/types';

export async function listEvents(
  orgSlug: string,
  filters?: EventFilters,
): Promise<PaginatedResponse<AgentEvent>> {
  const { data } = await apiClient.get<PaginatedResponse<AgentEvent>>(
    `/orgs/${orgSlug}/events`,
    { params: filters },
  );
  return { ...data, data: data.data ?? [] };
}

export async function getEventStats(orgSlug: string): Promise<EventStats> {
  const { data } = await apiClient.get<EventStats>(`/orgs/${orgSlug}/events/stats`);
  return data;
}

export async function getAgentEvents(
  orgSlug: string,
  agentId: string,
  filters?: EventFilters,
): Promise<PaginatedResponse<AgentEvent>> {
  const { data } = await apiClient.get<PaginatedResponse<AgentEvent>>(
    `/orgs/${orgSlug}/agents/${agentId}/events`,
    { params: filters },
  );
  return { ...data, data: data.data ?? [] };
}
