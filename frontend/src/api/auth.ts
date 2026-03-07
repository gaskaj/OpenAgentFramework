import apiClient from './client';
import type { User, Organization } from '@/types';

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

interface MeResponse {
  user: User;
  orgs: Organization[];
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/login', { email, password });
  return data;
}

export async function register(
  email: string,
  password: string,
  displayName: string,
): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/register', {
    email,
    password,
    display_name: displayName,
  });
  return data;
}

export async function refreshToken(token: string): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/auth/refresh', {
    refresh_token: token,
  });
  return data;
}

export async function getMe(): Promise<MeResponse> {
  const { data } = await apiClient.get<MeResponse>('/auth/me');
  return data;
}

export async function getOAuthURL(provider: string): Promise<string> {
  const { data } = await apiClient.get<{ url: string }>(`/auth/oauth/${provider}`);
  return data.url;
}
