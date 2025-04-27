import React, { createContext, useContext, useEffect, useState, ReactNode, useCallback, useMemo } from 'react';
import { components } from '../api/types';

export type User = components['schemas']['User'] | null;

interface UserContextValue {
  user: User;
  setUser: (user: User) => void;
  loading: boolean;
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

// Custom error for API responses
class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

// Type guard for user data
function isValidUserData(data: unknown): data is components['schemas']['User'] {
  return data !== null && 
    typeof data === 'object' && 
    'id' in data && 
    typeof data['id'] === 'string';
}

export const UserProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User>(null);
  const [loading, setLoading] = useState(true);
  const [isRedirecting, setIsRedirecting] = useState(false);

  // Handle session expiry
  const handleSessionExpiry = useCallback((): void => {
    setIsRedirecting(true);
    clearAllAuth();
    const baseUrl = window.location.origin;
    window.location.href = `${baseUrl}/login?reason=session_expired`;
  }, []);

  // Fetch and validate user data
  const fetchUserData = useCallback(async (): Promise<User> => {
    const res = await fetch('/api/users/me', {
      credentials: 'include',
      headers: {
        'Accept': 'application/json',
      },
    });

    if (!res.ok) {
      throw new ApiError(res.status, `HTTP error! status: ${res.status}`);
    }

    const data = await res.json();
    
    if (!isValidUserData(data)) {
      throw new Error('Invalid user data received');
    }

    return data;
  }, []);

  // Fetch user effect
  useEffect(() => {
    // Skip fetch if on login page or already redirecting
    if (window.location.pathname === '/login' || isRedirecting) {
      setLoading(false);
      return;
    }

    let mounted = true;

    // Main fetch user function
    const fetchUser = async (): Promise<void> => {
      if (!mounted) return;
      setLoading(true);
      
      try {
        const userData = await fetchUserData();
        if (mounted) setUser(userData);
      } catch (error) {
        if (!mounted) return;

        // Handle specific error types
        if (error instanceof ApiError && error.status === 401) {
          handleSessionExpiry();
          return;
        }

        // Log other errors but don't expose to user
        console.error('Error fetching user:', error);
        setUser(null);
      } finally {
        // Only set loading false if we haven't redirected and component is mounted
        if (!isRedirecting && mounted) {
          setLoading(false);
        }
      }
    };

    void fetchUser();

    // Cleanup function to prevent state updates on unmounted component
    return () => {
      mounted = false;
    };
  }, [isRedirecting, fetchUserData, handleSessionExpiry]);

  const logout = useCallback(() => {
    clearAllAuth();
    setUser(null);
  }, []);

  const contextValue = useMemo(() => ({
    user,
    setUser,
    loading,
    logout,
  }), [user, loading, logout]);

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
