import { useUser } from '../context/UserContext';
import { useLocation, Navigate } from 'react-router-dom';
import React from 'react';

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, loading } = useUser();
  const location = useLocation();

  if (loading) return <div className="flex justify-center items-center min-h-screen text-white">Loading…</div>;
  if (!user) {
    return <Navigate to={`/login?reason=session_expired&redirect=${encodeURIComponent(location.pathname)}`} replace />;
  }
  return <>{children}</>;
} 