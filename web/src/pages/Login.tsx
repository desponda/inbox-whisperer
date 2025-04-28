import React from 'react';
import '/fonts/inter.css';
import { GoogleSignInButton } from '../components/GoogleSignInButton';

const Login: React.FC = () => {
  const searchParams = new URLSearchParams(window.location.search);
  const reason = searchParams.get('reason');
  const [message, setMessage] = React.useState<string>('');

  React.useEffect(() => {
    if (reason === 'session_expired') {
      setMessage('Your session has expired. Please log in again.');
    }
  }, [reason]);

  return (
    <main className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-b from-[#181a23] to-[#0e1015] font-sans" style={{ fontFamily: "'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif" }}>
      <style>{`
        body, .font-sans {
          font-family: 'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif;
        }
      `}</style>
      {/* Modern shiny dark gradient background */}
      <div
        aria-hidden
        className="fixed inset-0 z-0 pointer-events-none"
        style={{
          background:
            'radial-gradient(ellipse 120% 80% at 50% 0%, #23243a 80%, #0e1015 100%), linear-gradient(120deg,rgba(20,224,201,0.07) 0%,rgba(0,180,216,0.07) 100%)',
          opacity: 1,
        }}
      />
      {/* NAVBAR */}
      <nav
        className="w-full flex items-center justify-between px-4 py-3 bg-transparent z-10"
        style={{
          fontFamily: "'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif",
          boxShadow: 'none',
          border: 'none',
        }}
      >
        <div className="flex items-center gap-2">
          <svg width="28" height="28" fill="none" viewBox="0 0 24 24">
            <rect width="24" height="24" rx="6" fill="#14e0c9" />
            <path
              d="M7 12l5 5 5-5"
              stroke="#23272f"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <span
            className="text-xl font-normal tracking-tight text-white"
            style={{ fontFamily: 'inherit', fontWeight: 400, letterSpacing: '-0.01em' }}
          >
            Inbox Whisperer
          </span>
        </div>
        <div className="hidden md:flex items-center gap-8">
          <a
            href="/#features"
            className="text-base font-normal text-gray-300 hover:text-accent transition"
          >
            Features
          </a>
          <a
            href="/#about"
            className="text-base font-normal text-gray-300 hover:text-accent transition"
          >
            About
          </a>
          <a
            href="https://github.com/desponda/inbox-whisperer"
            target="_blank"
            rel="noopener noreferrer"
            className="text-base font-normal text-gray-300 hover:text-accent transition"
          >
            GitHub
          </a>
          <a href="/login" className="ml-4 btn btn-accent font-semibold rounded-full px-6">
            Sign in
          </a>
        </div>
        <div className="md:hidden">{/* Mobile menu placeholder */}</div>
      </nav>
      {/* LOGIN SECTION */}
      <section className="flex-1 flex flex-col items-center justify-center px-4 z-10 relative">
        <div className="w-full max-w-md mx-auto p-8 rounded-2xl shadow-xl bg-[#181a23]/80 mt-24 border border-[#23243a]">
          <h1 className="text-3xl font-extrabold mb-6 text-center text-white tracking-tight">Welcome to Inbox Whisperer</h1>
          {message && (
            <div className="mb-4 px-4 py-3 rounded-xl bg-gradient-to-r from-red-500/80 to-pink-500/80 text-white text-center font-semibold border border-red-400 shadow-lg">
              <span className="inline-flex items-center gap-2">
                <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9 9 4.03 9 9z" /></svg>
                {message}
              </span>
            </div>
          )}
          <div className="flex flex-col gap-4 items-center">
            <GoogleSignInButton />
          </div>
          <p className="mt-6 text-xs text-gray-400 text-center">
            We never see your password. Google login is required for inbox access.
          </p>
        </div>
      </section>
      {/* FOOTER */}
      <footer className="footer footer-center p-4 bg-neutral bg-opacity-90 text-gray-400 border-t border-neutral-content/10 shadow-inner z-10">
        <div>
          <p>
            &copy; {new Date().getFullYear()} Inbox Whisperer &mdash; Built with{' '}
            <span className="text-accent font-semibold">DaisyUI</span> and{' '}
            <span className="text-primary font-semibold">Windsurf</span>
          </p>
        </div>
      </footer>
      {/* Custom keyframes for background and button glow */}
      <style>{`
        @keyframes bg-move {
          0% { background-position: 0 0, 0 0; }
          100% { background-position: 200px 400px, 400px 200px; }
        }
        .animate-glow {
          box-shadow: 0 0 16px 4px #14e0c977, 0 8px 32px 0 #14e0c955;
          transition: box-shadow 0.25s;
        }
        .animate-glow:hover, .animate-glow:focus {
          box-shadow: 0 0 32px 8px #14e0c9cc, 0 8px 32px 0 #14e0c955;
        }
        .animate-fade-in {
          animation: fadeIn 1.2s cubic-bezier(.22,1,.36,1) both;
        }
        @keyframes fadeIn {
          from { opacity: 0; transform: translateY(30px); }
          to { opacity: 1; transform: none; }
        }
      `}</style>
    </main>
  );
};

export default Login;
