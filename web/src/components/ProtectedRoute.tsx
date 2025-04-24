import React from 'react';
import { Navigate } from 'react-router-dom';
import { useUser } from '../context/UserContext';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, loading } = useUser();

  if (loading) {
    // Optionally show a loading spinner
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#0e1015]">
        <span className="loading loading-spinner loading-lg text-accent"></span>
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

export default ProtectedRoute;
