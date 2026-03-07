import axios, { AxiosError } from 'axios';
import type { ApiError } from '@/types';
import { useAuthStore } from '@/store/auth-store';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach JWT token to every request
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 responses: attempt token refresh or redirect to login
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
      originalRequest._retry = true;

      const refreshToken = useAuthStore.getState().refreshToken;
      if (refreshToken) {
        try {
          const { data } = await axios.post(
            `${apiClient.defaults.baseURL}/auth/refresh`,
            { refresh_token: refreshToken },
          );
          useAuthStore.getState().setToken(data.access_token, data.refresh_token);
          originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
          return apiClient(originalRequest);
        } catch {
          useAuthStore.getState().logout();
          window.location.href = '/login';
          return Promise.reject(error);
        }
      }

      useAuthStore.getState().logout();
      window.location.href = '/login';
    }

    // Normalize error
    const normalized: ApiError = error.response?.data ?? {
      error: 'network_error',
      message: error.message || 'An unexpected error occurred',
      status: error.response?.status ?? 0,
    };

    return Promise.reject(normalized);
  },
);

// Extend AxiosRequestConfig to support _retry flag
declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _retry?: boolean;
  }
}

export default apiClient;
