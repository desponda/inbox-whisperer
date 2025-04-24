import React from 'react';

const Login: React.FC = () => (
  <main className="min-h-screen flex items-center justify-center bg-base-200">
    <div className="p-8 rounded-box shadow-xl bg-base-100 flex flex-col items-center">
      <h2 className="text-2xl font-bold mb-4 text-primary">Sign in to Inbox Whisperer</h2>
      <a
        href="/auth/login"
        className="btn btn-primary btn-wide text-lg mb-2"
        data-testid="google-login-btn"
      >
        <svg className="w-5 h-5 mr-2 inline-block" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><g><path fill="#4285F4" d="M24 9.5c3.54 0 6.7 1.22 9.19 3.22l6.85-6.85C36.64 2.42 30.79 0 24 0 14.82 0 6.51 5.8 2.23 14.09l7.98 6.2C12.06 13.36 17.57 9.5 24 9.5z"/><path fill="#34A853" d="M46.1 24.55c0-1.64-.15-3.22-.42-4.74H24v9.01h12.41c-.54 2.9-2.18 5.36-4.65 7.01l7.22 5.62C43.83 37.27 46.1 31.39 46.1 24.55z"/><path fill="#FBBC05" d="M10.21 28.29c-1.13-3.36-1.13-6.93 0-10.29l-7.98-6.2C-1.13 16.16-1.13 31.84 2.23 33.91l7.98-6.2z"/><path fill="#EA4335" d="M24 44c6.48 0 11.92-2.14 15.9-5.84l-7.22-5.62c-2.02 1.36-4.6 2.16-8.68 2.16-6.43 0-11.94-3.86-13.79-9.29l-7.98 6.2C6.51 42.2 14.82 48 24 48z"/><path fill="none" d="M0 0h48v48H0z"/></g></svg>
        Sign in with Google
      </a>
      <p className="text-sm text-base-content/60 mt-2">We never see your password. Google login is required for inbox access.</p>
    </div>
  </main>
);

export default Login;
