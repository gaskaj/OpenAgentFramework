import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AgentDetailPage } from './AgentDetailPage';
import { useAuthStore } from '@/store/auth-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/hooks/useAgents', () => ({
  useAgent: () => ({
    agent: null,
    loading: false,
    error: null,
  }),
}));

vi.mock('@/api/events', () => ({
  getAgentEvents: vi.fn().mockResolvedValue({ data: [] }),
}));

vi.mock('@/api/agents', () => ({
  deleteAgent: vi.fn(),
}));

describe('AgentDetailPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER, token: 'test-token', refreshToken: 'test-refresh', currentOrg: TEST_ORG,
    });
  });

  it('renders without crashing', () => {
    expect(() =>
      render(
        <MemoryRouter initialEntries={['/agents/some-id']}>
          <Routes>
            <Route path="/agents/:agentId" element={<AgentDetailPage />} />
          </Routes>
        </MemoryRouter>
      )
    ).not.toThrow();
  });

  it('shows not found when agent is null', () => {
    render(
      <MemoryRouter initialEntries={['/agents/some-id']}>
        <Routes>
          <Route path="/agents/:agentId" element={<AgentDetailPage />} />
        </Routes>
      </MemoryRouter>
    );
    expect(screen.getByText('Agent not found.')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() =>
      render(
        <MemoryRouter initialEntries={['/agents/some-id']}>
          <Routes>
            <Route path="/agents/:agentId" element={<AgentDetailPage />} />
          </Routes>
        </MemoryRouter>
      )
    ).not.toThrow();
  });
});
