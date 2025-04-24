import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import '/fonts/inter.css';

// You can import types from your OpenAPI-generated types if needed
// import { components } from '../api/types';

const AuthCallback: React.FC = () => {
  const [status, setStatus] = useState<'loading' | 'error' | 'onboarding' | 'success'>('loading');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Parse query params
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');
    // Simulate API call for demo; replace with real call to your backend
    async function handleCallback() {
      if (!code || !state) {
        setStatus('error');
        setError('Missing code or state in callback URL.');
        return;
      }
      try {
        // Call your backend to finish OAuth (could be a fetch to /api/auth/callback)
        // For MVP, assume backend sets session cookie and redirects here
        // Optionally fetch user info to determine if onboarding is needed
        // Simulate onboarding for first-time users
        const isFirstTime = params.get('first') === '1';
        if (isFirstTime) {
          setStatus('onboarding');
        } else {
          setStatus('success');
          setTimeout(() => navigate('/'), 1200);
        }
      } catch {
        setStatus('error');
        setError('Authentication failed.');
      }
    }
    handleCallback();
  }, [navigate]);

  return (
    <main className="min-h-screen flex flex-col text-gray-100 bg-[#0e1015]" style={{ fontFamily: "'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif" }}>
      <div aria-hidden className="fixed inset-0 z-0 pointer-events-none" style={{background:'radial-gradient(ellipse 120% 80% at 50% 0%, #23243a 80%, #0e1015 100%), linear-gradient(120deg,rgba(20,224,201,0.07) 0%,rgba(0,180,216,0.07) 100%)',opacity:1}} />
      <nav className="w-full flex items-center justify-between px-4 py-3 bg-transparent z-10">
        <div className="flex items-center gap-2">
          <svg width="28" height="28" fill="none" viewBox="0 0 24 24"><rect width="24" height="24" rx="6" fill="#14e0c9"/><path d="M7 12l5 5 5-5" stroke="#23272f" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/></svg>
          <span className="text-xl font-normal tracking-tight text-white">Inbox Whisperer</span>
        </div>
      </nav>
      <section className="flex-1 flex flex-col items-center justify-center px-4 z-10 relative">
        <div className="w-full max-w-md flex flex-col items-center justify-center text-center py-16">
          {status === 'loading' && (
            <>
              <div className="mb-6 animate-spin rounded-full h-12 w-12 border-t-4 border-accent border-solid" />
              <h2 className="text-2xl font-bold mb-2">Signing you in…</h2>
              <p className="text-base text-gray-300">Please wait while we complete your sign-in.</p>
            </>
          )}
          {status === 'onboarding' && (
            <>
              <h2 className="text-3xl font-black mb-4 tracking-tight leading-tight">
                Welcome to <span className="bg-gradient-to-r from-cyan-400 via-blue-400 to-accent bg-clip-text text-transparent" style={{ WebkitTextStroke: '0.5px #14e0c9' }}>Inbox Whisperer</span>!
              </h2>
              <p className="text-lg text-gray-300 mb-6">Your account is ready. Let's get started.</p>
              <button className="btn btn-accent btn-lg font-bold rounded-full px-8 shadow-xl" onClick={() => navigate('/')}>Continue</button>
            </>
          )}
          {status === 'success' && (
            <>
              <h2 className="text-2xl font-bold mb-2 text-accent">Success!</h2>
              <p className="text-base text-gray-300">You are now signed in. Redirecting…</p>
            </>
          )}
          {status === 'error' && (
            <>
              <h2 className="text-2xl font-bold mb-2 text-error">Authentication Error</h2>
              <p className="text-base text-gray-300">{error}</p>
              <a href="/login" className="btn btn-accent mt-6">Back to Login</a>
            </>
          )}
        </div>
      </section>
      <footer className="footer footer-center p-4 bg-neutral bg-opacity-90 text-gray-400 border-t border-neutral-content/10 shadow-inner z-10">
        <div>
          <p>
            &copy; {new Date().getFullYear()} Inbox Whisperer &mdash; Built with <span className="text-accent font-semibold">DaisyUI</span> and <span className="text-primary font-semibold">Windsurf</span>
          </p>
        </div>
      </footer>
    </main>
  );
};

export default AuthCallback;
