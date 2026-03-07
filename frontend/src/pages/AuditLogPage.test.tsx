import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AuditLogPage } from './AuditLogPage';
import { useAuthStore } from '@/store/auth-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn().mockResolvedValue({ data: { data: [], total: 0, limit: 25, offset: 0 } }),
  },
}));

describe('AuditLogPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER, token: 'test-token', refreshToken: 'test-refresh', currentOrg: TEST_ORG,
    });
  });

  it('renders without crashing', () => {
    expect(() => render(<AuditLogPage />)).not.toThrow();
  });

  it('renders heading', () => {
    render(<AuditLogPage />);
    expect(screen.getByText('Audit Log')).toBeInTheDocument();
  });

  it('renders filter inputs', () => {
    render(<AuditLogPage />);
    expect(screen.getByPlaceholderText('Filter by action...')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Filter by resource type...')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() => render(<AuditLogPage />)).not.toThrow();
  });
});
