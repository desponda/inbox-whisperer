import React, { createContext, useContext, ReactNode, useCallback, useMemo } from 'react';
import useSWR from 'swr';
import { getUser } from '../api/generated/user/user';
import { components } from '../api/types';

export type User = components['schemas']['User'] | null;

interface UserContextValue {
  user: User;
  loading: boolean;
  error: any;
  mutate: () => void;
  logout: () => void;
}

const UserContext = createContext<UserContextValue | undefined>(undefined);

export function clearAllAuth() {
  // Expire all cookies (browser only allows path-level, so this is best effort)
  document.cookie.split(';').forEach(cookie => {
    const eqPos = cookie.indexOf('=');
    const name = eqPos > -1 ? cookie.substr(0, eqPos).trim() : cookie.trim();
    document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;SameSite=Strict`;
    // Also try root path
    document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/api;SameSite=Strict`;
  });
  localStorage.clear();
  sessionStorage.clear();
}

export const UserProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  // Use SWR for user session state
  const fetchUser = async () => {
    try {
      const res = await getUser().getApiUsersMe({ withCredentials: true });
      return res.data;
    } catch (error: any) {
      if (error?.response?.status === 401) {
        // Session expired, clear auth, but do not redirect
        clearAllAuth();
        return null;
      }
      throw error;
    }
  };

  const { data: user, error, isLoading, mutate } = useSWR('/api/users/me', fetchUser, {
    revalidateOnFocus: true,
    shouldRetryOnError: false,
  });

  const logout = useCallback(() => {
    clearAllAuth();
    mutate();
    window.location.href = '/login';
  }, [mutate]);

  const contextValue = useMemo(() => ({
    user: user ?? null,
    loading: isLoading,
    error,
    mutate,
    logout,
  }), [user, isLoading, error, mutate, logout]);

  return (
    <UserContext.Provider value={contextValue}>
      {children}
    </UserContext.Provider>
  );
};

export function useUser() {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}
