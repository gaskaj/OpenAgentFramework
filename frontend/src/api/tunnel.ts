import apiClient from './client';

export interface TunnelStatus {
  enabled: boolean;
  public_url?: string;
  error?: string;
  has_auth_token: boolean;
}

export async function getTunnelStatus(): Promise<TunnelStatus> {
  const { data } = await apiClient.get<TunnelStatus>('/tunnel');
  return data;
}

export async function toggleTunnel(enabled: boolean): Promise<TunnelStatus> {
  const { data } = await apiClient.post<TunnelStatus>('/tunnel', { enabled });
  return data;
}

export async function saveAuthToken(authToken: string): Promise<TunnelStatus> {
  const { data } = await apiClient.put<TunnelStatus>('/tunnel/token', { auth_token: authToken });
  return data;
}
