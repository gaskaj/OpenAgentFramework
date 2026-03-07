import apiClient from './client';
import type { Organization, OrgMember, Invitation, OrgRole } from '@/types';

export async function listOrgs(): Promise<Organization[]> {
  const { data } = await apiClient.get<{ data: Organization[] }>('/orgs');
  return data.data ?? [];
}

export async function getOrg(slug: string): Promise<Organization> {
  const { data } = await apiClient.get<{ data: Organization }>(`/orgs/${slug}`);
  return data.data;
}

export async function createOrg(orgData: {
  name: string;
  slug: string;
}): Promise<Organization> {
  const { data } = await apiClient.post<{ data: Organization }>('/orgs', orgData);
  return data.data;
}

export async function updateOrg(
  slug: string,
  orgData: Partial<Organization>,
): Promise<Organization> {
  const { data } = await apiClient.patch<{ data: Organization }>(`/orgs/${slug}`, orgData);
  return data.data;
}

export async function listMembers(slug: string): Promise<OrgMember[]> {
  const { data } = await apiClient.get<{ data: OrgMember[] }>(`/orgs/${slug}/members`);
  return data.data ?? [];
}

export async function updateMemberRole(
  slug: string,
  userId: string,
  role: OrgRole,
): Promise<OrgMember> {
  const { data } = await apiClient.patch<OrgMember>(
    `/orgs/${slug}/members/${userId}`,
    { role },
  );
  return data;
}

export async function removeMember(slug: string, userId: string): Promise<void> {
  await apiClient.delete(`/orgs/${slug}/members/${userId}`);
}

export async function listInvitations(slug: string): Promise<Invitation[]> {
  const { data } = await apiClient.get<{ data: Invitation[] }>(`/orgs/${slug}/invitations`);
  return data.data ?? [];
}

export async function createInvitation(
  slug: string,
  email: string,
  role: OrgRole,
): Promise<Invitation> {
  const { data } = await apiClient.post<{ data: Invitation }>(`/orgs/${slug}/invitations`, {
    email,
    role,
  });
  return data.data;
}

export async function cancelInvitation(slug: string, invId: string): Promise<void> {
  await apiClient.delete(`/orgs/${slug}/invitations/${invId}`);
}

export async function acceptInvitation(token: string): Promise<{ org: Organization }> {
  const { data } = await apiClient.post<{ org: Organization }>(
    `/invitations/${token}/accept`,
  );
  return data;
}
