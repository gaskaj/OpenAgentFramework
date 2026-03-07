import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { AgentListPage } from './AgentListPage';
import { useAuthStore } from '@/store/auth-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/hooks/useAgents', () => ({
  useAgents: () => ({
    agents: [],
    loading: false,
    error: null,
    total: 0,
    page: 1,
    refresh: vi.fn(),
  }),
}));

describe('AgentListPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER, token: 'test-token', refreshToken: 'test-refresh', currentOrg: TEST_ORG,
    });
  });

  it('renders without crashing', () => {
    expect(() =>
      render(<MemoryRouter><AgentListPage /></MemoryRouter>)
    ).not.toThrow();
  });

  it('renders heading', () => {
    render(<MemoryRouter><AgentListPage /></MemoryRouter>);
    expect(screen.getByText('Agents')).toBeInTheDocument();
  });

  it('shows empty state when no agents', () => {
    render(<MemoryRouter><AgentListPage /></MemoryRouter>);
    expect(screen.getByText('No agents found.')).toBeInTheDocument();
  });

  it('renders search input', () => {
    render(<MemoryRouter><AgentListPage /></MemoryRouter>);
    expect(screen.getByPlaceholderText('Search agents...')).toBeInTheDocument();
  });

  it('renders status filter tabs', () => {
    render(<MemoryRouter><AgentListPage /></MemoryRouter>);
    expect(screen.getByText('All')).toBeInTheDocument();
    expect(screen.getByText('Online')).toBeInTheDocument();
    expect(screen.getByText('Offline')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() =>
      render(<MemoryRouter><AgentListPage /></MemoryRouter>)
    ).not.toThrow();
  });

});
