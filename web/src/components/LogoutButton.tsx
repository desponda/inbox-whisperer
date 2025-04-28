import React from 'react';
import { useUser } from '../context/UserContext';

const LogoutButton: React.FC = () => {
  const { logout, loading } = useUser();
  return (
    <button
      className="btn btn-outline btn-error"
      onClick={logout}
      disabled={loading}
      data-testid="logout-btn"
    >
      Log out
    </button>
  );
};

export default LogoutButton; 