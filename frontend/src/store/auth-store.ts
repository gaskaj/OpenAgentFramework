import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, Organization } from '@/types';

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  currentOrg: Organization | null;

  setUser: (user: User) => void;
  setToken: (token: string, refreshToken: string) => void;
  setCurrentOrg: (org: Organization) => void;
  login: (user: User, token: string, refreshToken: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      refreshToken: null,
      currentOrg: null,

      setUser: (user) => set({ user }),

      setToken: (token, refreshToken) => set({ token, refreshToken }),

      setCurrentOrg: (org) => set({ currentOrg: org }),

      login: (user, token, refreshToken) =>
        set({ user, token, refreshToken }),

      logout: () =>
        set({
          user: null,
          token: null,
          refreshToken: null,
          currentOrg: null,
        }),
    }),
    {
      name: 'oaf-auth',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        currentOrg: state.currentOrg,
      }),
    },
  ),
);
