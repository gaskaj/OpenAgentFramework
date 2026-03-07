import { vi } from 'vitest';
import type { Organization, User } from '@/types';

export const TEST_USER: User = {
  id: 'u-1',
  email: 'test@test.com',
  display_name: 'Test User',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

export const TEST_ORG: Organization = {
  id: 'org-1',
  name: 'Test Org',
  slug: 'test-org',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

export const mockUseAuth = () => ({
  useAuth: () => ({
    register: vi.fn(),
    login: vi.fn(),
    logout: vi.fn(),
    startOAuth: vi.fn(),
    user: TEST_USER,
    token: 'test-token',
    currentOrg: TEST_ORG,
    isAuthenticated: true,
    setCurrentOrg: vi.fn(),
  }),
});
