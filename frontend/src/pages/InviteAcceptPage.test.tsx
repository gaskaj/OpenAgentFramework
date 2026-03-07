import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { InviteAcceptPage } from './InviteAcceptPage';
import { useAuthStore } from '@/store/auth-store';

vi.mock('@/api/orgs', () => ({
  acceptInvitation: vi.fn().mockRejectedValue(new Error('test')),
}));

describe('InviteAcceptPage', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null, token: null, refreshToken: null, currentOrg: null });
  });

  it('renders without crashing', () => {
    expect(() =>
      render(
        <MemoryRouter initialEntries={['/invite/some-token']}>
          <Routes>
            <Route path="/invite/:token" element={<InviteAcceptPage />} />
          </Routes>
        </MemoryRouter>
      )
    ).not.toThrow();
  });

  it('shows login required when not authenticated', () => {
    render(
      <MemoryRouter initialEntries={['/invite/some-token']}>
        <Routes>
          <Route path="/invite/:token" element={<InviteAcceptPage />} />
        </Routes>
      </MemoryRouter>
    );
    expect(screen.getByText('Sign in required')).toBeInTheDocument();
  });

  it('renders sign in and create account links', () => {
    render(
      <MemoryRouter initialEntries={['/invite/some-token']}>
        <Routes>
          <Route path="/invite/:token" element={<InviteAcceptPage />} />
        </Routes>
      </MemoryRouter>
    );
    expect(screen.getByText('Sign in')).toBeInTheDocument();
    expect(screen.getByText('Create account')).toBeInTheDocument();
  });
});
