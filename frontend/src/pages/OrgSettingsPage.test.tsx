import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { OrgSettingsPage } from './OrgSettingsPage';
import { useAuthStore } from '@/store/auth-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/api/orgs', () => ({
  listMembers: vi.fn().mockResolvedValue([]),
  listInvitations: vi.fn().mockResolvedValue([]),
  updateOrg: vi.fn(),
  createInvitation: vi.fn(),
  updateMemberRole: vi.fn(),
  removeMember: vi.fn(),
  cancelInvitation: vi.fn(),
}));

describe('OrgSettingsPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER, token: 'test-token', refreshToken: 'test-refresh', currentOrg: TEST_ORG,
    });
  });

  it('renders without crashing', () => {
    expect(() => render(<OrgSettingsPage />)).not.toThrow();
  });

  it('renders heading', () => {
    render(<OrgSettingsPage />);
    expect(screen.getByText('Organization Settings')).toBeInTheDocument();
  });

  it('renders org details section', () => {
    render(<OrgSettingsPage />);
    expect(screen.getByText('Organization Details')).toBeInTheDocument();
  });

  it('renders members section', () => {
    render(<OrgSettingsPage />);
    expect(screen.getByText('Members')).toBeInTheDocument();
  });

  it('renders invite section', () => {
    render(<OrgSettingsPage />);
    expect(screen.getByText('Invite Member')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() => render(<OrgSettingsPage />)).not.toThrow();
  });
});
