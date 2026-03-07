import { describe, it, expect, beforeEach } from 'vitest';
import { useAuthStore } from './auth-store';

describe('auth-store', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      token: null,
      refreshToken: null,
      currentOrg: null,
    });
  });

  it('starts with null state', () => {
    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.currentOrg).toBeNull();
  });

  it('login sets user, token, and refreshToken', () => {
    const user = { id: '1', email: 'test@example.com', display_name: 'Test', created_at: '', updated_at: '' };
    useAuthStore.getState().login(user, 'tok', 'ref');

    const state = useAuthStore.getState();
    expect(state.user).toEqual(user);
    expect(state.token).toBe('tok');
    expect(state.refreshToken).toBe('ref');
  });

  it('setCurrentOrg sets the current org', () => {
    const org = { id: '1', name: 'Test Org', slug: 'test-org', created_at: '', updated_at: '' };
    useAuthStore.getState().setCurrentOrg(org);

    expect(useAuthStore.getState().currentOrg).toEqual(org);
  });

  it('logout clears all state', () => {
    const user = { id: '1', email: 'test@example.com', display_name: 'Test', created_at: '', updated_at: '' };
    const org = { id: '1', name: 'Test Org', slug: 'test-org', created_at: '', updated_at: '' };

    useAuthStore.getState().login(user, 'tok', 'ref');
    useAuthStore.getState().setCurrentOrg(org);
    useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.currentOrg).toBeNull();
  });
});
