import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { LoginPage } from './LoginPage';
import { useAuthStore } from '@/store/auth-store';

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    login: vi.fn(),
    startOAuth: vi.fn(),
    user: null,
    token: null,
    currentOrg: null,
    isAuthenticated: false,
    setCurrentOrg: vi.fn(),
  }),
}));

describe('LoginPage', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null, token: null, refreshToken: null, currentOrg: null });
  });

  it('renders without crashing', () => {
    expect(() =>
      render(<MemoryRouter><LoginPage /></MemoryRouter>)
    ).not.toThrow();
  });

  it('renders sign in heading', () => {
    render(<MemoryRouter><LoginPage /></MemoryRouter>);
    expect(screen.getByRole('heading', { name: 'Sign in' })).toBeInTheDocument();
  });

  it('renders email and password fields', () => {
    render(<MemoryRouter><LoginPage /></MemoryRouter>);
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
  });

  it('renders sign up link', () => {
    render(<MemoryRouter><LoginPage /></MemoryRouter>);
    expect(screen.getByText('Sign up')).toBeInTheDocument();
  });

  it('redirects when already authenticated', () => {
    useAuthStore.setState({ token: 'existing-token' });
    render(<MemoryRouter initialEntries={['/login']}><LoginPage /></MemoryRouter>);
    expect(screen.queryByText('Sign in')).not.toBeInTheDocument();
  });
});
