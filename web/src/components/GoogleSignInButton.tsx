import React from 'react';

export const GoogleSignInButton: React.FC<{ className?: string }> = ({ className = '' }) => (
  <a
    href="/api/auth/login"
    className={`btn btn-accent w-full max-w-xs text-lg font-semibold flex items-center justify-center gap-2 shadow-md ${className}`}
    data-testid="google-login-btn"
  >
    <span className="inline-block w-6 h-6 mr-2 align-middle">
      <svg
        className="w-6 h-6"
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 48 48"
      >
        <g>
          <path
            fill="#4285F4"
            d="M24 9.5c3.54 0 6.7 1.22 9.19 3.22l6.85-6.85C36.64 2.42 30.79 0 24 0 14.82 0 6.51 5.8 2.23 14.09l7.98 6.2C12.06 13.36 17.57 9.5 24 9.5z"
          />
          <path
            fill="#34A853"
            d="M46.1 24.55c0-1.64-.15-3.22-.42-4.74H24v9.01h12.41c-.54 2.9-2.18 5.36-4.65 7.01l7.22 5.62C43.83 37.27 46.1 31.39 46.1 24.55z"
          />
          <path
            fill="#FBBC05"
            d="M10.21 28.29c-1.13-3.36-1.13-6.93 0-10.29l-7.98-6.2C-1.13 16.16-1.13 31.84 2.23 33.91l7.98-6.2z"
          />
          <path
            fill="#EA4335"
            d="M24 44c6.48 0 11.92-2.14 15.9-5.84l-7.22-5.62c-2.02 1.36-4.6 2.16-8.68 2.16-6.43 0-11.94-3.86-13.79-9.29l-7.98 6.2C6.51 42.2 14.82 48 24 48z"
          />
          <path fill="none" d="M0 0h48v48H0z" />
        </g>
      </svg>
    </span>
    Sign in with Google
  </a>
); 