import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { RegisterPage } from './RegisterPage';
import { useAuthStore } from '@/store/auth-store';

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    register: vi.fn(),
    login: vi.fn(),
    logout: vi.fn(),
    startOAuth: vi.fn(),
    user: null,
    token: null,
    currentOrg: null,
    isAuthenticated: false,
    setCurrentOrg: vi.fn(),
  }),
}));

describe('RegisterPage', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null, token: null, refreshToken: null, currentOrg: null });
  });

  it('renders without crashing', () => {
    expect(() =>
      render(<MemoryRouter><RegisterPage /></MemoryRouter>)
    ).not.toThrow();
  });

  it('renders registration form fields', () => {
    render(<MemoryRouter><RegisterPage /></MemoryRouter>);
    expect(screen.getByRole('heading', { name: 'Create account' })).toBeInTheDocument();
    expect(screen.getByLabelText('Display name')).toBeInTheDocument();
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Create account' })).toBeInTheDocument();
  });

  it('shows sign in link', () => {
    render(<MemoryRouter><RegisterPage /></MemoryRouter>);
    expect(screen.getByText('Sign in')).toBeInTheDocument();
  });

  it('redirects to dashboard when already authenticated', () => {
    useAuthStore.setState({ token: 'existing-token' });
    render(<MemoryRouter initialEntries={['/register']}><RegisterPage /></MemoryRouter>);
    expect(screen.queryByRole('heading', { name: 'Create account' })).not.toBeInTheDocument();
  });
});
