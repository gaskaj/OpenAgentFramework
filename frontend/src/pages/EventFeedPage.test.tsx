import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { EventFeedPage } from './EventFeedPage';
import { useAuthStore } from '@/store/auth-store';
import { useAgentStore } from '@/store/agent-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/hooks/useEvents', () => ({
  useEvents: () => ({
    events: [],
    loading: false,
    error: null,
    total: 0,
    page: 1,
    refresh: vi.fn(),
  }),
}));

vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: vi.fn(),
}));

describe('EventFeedPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER, token: 'test-token', refreshToken: 'test-refresh', currentOrg: TEST_ORG,
    });
    useAgentStore.setState({ agents: [], loading: false, error: null, total: 0, page: 1 });
  });

  it('renders without crashing', () => {
    expect(() => render(<EventFeedPage />)).not.toThrow();
  });

  it('renders heading', () => {
    render(<EventFeedPage />);
    expect(screen.getByText('Event Feed')).toBeInTheDocument();
  });

  it('renders filters section', () => {
    render(<EventFeedPage />);
    expect(screen.getByText('Filters')).toBeInTheDocument();
  });

  it('renders empty state', () => {
    render(<EventFeedPage />);
    expect(screen.getByText('No events match your filters.')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() => render(<EventFeedPage />)).not.toThrow();
  });
});
