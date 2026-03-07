import { useCallback, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/auth-store';
import * as authApi from '@/api/auth';

export function useAuth() {
  const {
    user,
    token,
    currentOrg,
    login: storeLogin,
    logout: storeLogout,
    setUser,
    setCurrentOrg,
  } = useAuthStore();
  const navigate = useNavigate();

  // Fetch current user and orgs on mount if we have a token but no user
  useEffect(() => {
    if (token && !user) {
      authApi.getMe().then((me) => {
        setUser(me.user);
        if (!currentOrg && me.orgs && me.orgs.length > 0) {
          setCurrentOrg(me.orgs[0]);
        }
      }).catch(() => storeLogout());
    }
  }, [token, user, currentOrg, setUser, setCurrentOrg, storeLogout]);

  const login = useCallback(
    async (email: string, password: string) => {
      const result = await authApi.login(email, password);
      storeLogin(result.user, result.access_token, result.refresh_token);
      // Fetch orgs for the user and set the first one as current
      try {
        const me = await authApi.getMe();
        if (me.orgs && me.orgs.length > 0) {
          setCurrentOrg(me.orgs[0]);
        }
      } catch { /* org will be set on next page load via useEffect */ }
      navigate('/dashboard');
    },
    [storeLogin, setCurrentOrg, navigate],
  );

  const register = useCallback(
    async (email: string, password: string, displayName: string) => {
      const result = await authApi.register(email, password, displayName);
      storeLogin(result.user, result.access_token, result.refresh_token);
      // Fetch orgs for the newly registered user
      try {
        const me = await authApi.getMe();
        if (me.orgs && me.orgs.length > 0) {
          setCurrentOrg(me.orgs[0]);
        }
      } catch { /* org will be set on next page load via useEffect */ }
      navigate('/dashboard');
    },
    [storeLogin, setCurrentOrg, navigate],
  );

  const logout = useCallback(() => {
    storeLogout();
    navigate('/login');
  }, [storeLogout, navigate]);

  const startOAuth = useCallback(async (provider: string) => {
    const url = await authApi.getOAuthURL(provider);
    window.location.href = url;
  }, []);

  return {
    user,
    token,
    currentOrg,
    isAuthenticated: !!token,
    login,
    register,
    logout,
    startOAuth,
    setCurrentOrg,
  };
}
