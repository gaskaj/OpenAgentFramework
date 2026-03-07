import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ApiKeysPage } from './ApiKeysPage';
import { useAuthStore } from '@/store/auth-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/api/apikeys', () => ({
  listAPIKeys: vi.fn().mockResolvedValue([]),
  createAPIKey: vi.fn(),
  revokeAPIKey: vi.fn(),
}));

describe('ApiKeysPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER, token: 'test-token', refreshToken: 'test-refresh', currentOrg: TEST_ORG,
    });
  });

  it('renders without crashing', () => {
    expect(() => render(<ApiKeysPage />)).not.toThrow();
  });

  it('renders heading', () => {
    render(<ApiKeysPage />);
    expect(screen.getByText('API Keys')).toBeInTheDocument();
  });

  it('renders create form', () => {
    render(<ApiKeysPage />);
    expect(screen.getByText('Create New Key')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Key name (e.g., production-agent)')).toBeInTheDocument();
  });

  it('renders active keys section', () => {
    render(<ApiKeysPage />);
    expect(screen.getByText('Active Keys')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() => render(<ApiKeysPage />)).not.toThrow();
  });
});
