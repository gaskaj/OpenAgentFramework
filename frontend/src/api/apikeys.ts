import apiClient from './client';
import type { APIKey, APIKeyCreateResponse } from '@/types';

export async function listAPIKeys(orgSlug: string): Promise<APIKey[]> {
  const { data } = await apiClient.get<{ data: APIKey[] }>(`/orgs/${orgSlug}/apikeys`);
  return data.data ?? [];
}

export async function createAPIKey(
  orgSlug: string,
  name: string,
): Promise<APIKeyCreateResponse> {
  const { data } = await apiClient.post<APIKeyCreateResponse>(
    `/orgs/${orgSlug}/apikeys`,
    { name },
  );
  return data;
}

export async function revokeAPIKey(orgSlug: string, keyId: string): Promise<void> {
  await apiClient.delete(`/orgs/${orgSlug}/apikeys/${keyId}`);
}
