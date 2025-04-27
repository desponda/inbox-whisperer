import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
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
    const name = eqPos > -1 ? cookie.substr(0, eqPos) : cookie;
    document.cookie = name + '=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;SameSite=Strict';
  });
  localStorage.clear();
  sessionStorage.clear();
}

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
          // On 401, clear cookies, localStorage, and redirect to login with message
          if (res.status === 401) {
            clearAllAuth();
            window.location.href = '/login?reason=session_expired';
            return;
          }
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
