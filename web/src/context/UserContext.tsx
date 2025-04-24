import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { components } from '../api/types';

export type User = components['schemas']['User'] | null;

interface UserContextValue {
  user: User;
  setUser: (user: User) => void;
  loading: boolean;
  logout: () => void;
}

const UserContext = createContext<UserContextValue | undefined>(undefined);

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Fetch current user info from backend (e.g., /api/users/me or session endpoint)
    async function fetchUser() {
      setLoading(true);
      try {
        const res = await fetch('/api/users/me', { credentials: 'include' });
        if (res.ok) {
          const data = await res.json();
          setUser(data);
        } else {
          setUser(null);
        }
      } catch {
        setUser(null);
      }
      setLoading(false);
    }
    fetchUser();
  }, []);

  function logout() {
    // For MVP, just clear user and optionally call a logout endpoint
    setUser(null);
    fetch('/api/auth/logout', { method: 'POST', credentials: 'include' });
  }

  return (
    <UserContext.Provider value={{ user, setUser, loading, logout }}>
      {children}
    </UserContext.Provider>
  );
};

export function useUser() {
  const ctx = useContext(UserContext);
  if (!ctx) throw new Error('useUser must be used within a UserProvider');
  return ctx;
}
