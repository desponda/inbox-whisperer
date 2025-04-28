import React from 'react';
import '../fonts/inter.css';
import { useUser } from '../context/UserContext';
import { GoogleSignInButton } from '../components/GoogleSignInButton';

const InboxIcon = () => (
  <svg
    width="160"
    height="160"
    viewBox="0 0 160 160"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <radialGradient id="glow" cx="50%" cy="50%" r="70%" fx="50%" fy="50%">
        <stop offset="0%" stopColor="#14e0c9" stopOpacity="0.25" />
        <stop offset="100%" stopColor="#14e0c9" stopOpacity="0" />
      </radialGradient>
      <linearGradient
        id="icon-gradient"
        x1="0"
        y1="0"
        x2="160"
        y2="160"
        gradientUnits="userSpaceOnUse"
      >
        <stop stopColor="#14e0c9" />
        <stop offset="1" stopColor="#00b4d8" />
      </linearGradient>
    </defs>
    <circle cx="80" cy="80" r="78" fill="url(#glow)" />
    <rect
      x="30"
      y="50"
      width="100"
      height="60"
      rx="16"
      fill="#202e3a"
      stroke="url(#icon-gradient)"
      strokeWidth="4"
    />
    <polyline
      points="40,60 80,95 120,60"
      fill="none"
      stroke="url(#icon-gradient)"
      strokeWidth="6"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

const Home: React.FC = () => {
  const { user, loading, logout } = useUser();

  return (
    <main
      className="min-h-screen flex flex-col text-gray-100 bg-[#0e1015]"
      style={{ fontFamily: "'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif" }}
    >
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
            href="#features"
            className="text-base font-normal text-gray-300 hover:text-accent transition"
          >
            Features
          </a>
          <a
            href="#about"
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
          {loading ? null : user ? (
            <>
              <span className="ml-4 text-base font-semibold text-accent">{user.email}</span>
              <button
                className="ml-2 btn btn-outline btn-accent rounded-full px-6 font-semibold"
                onClick={logout}
              >
                Sign out
              </button>
            </>
          ) : (
            <GoogleSignInButton />
          )}
        </div>
        <div className="md:hidden">{/* Mobile menu placeholder */}</div>
      </nav>

      {/* HERO SECTION */}
      <section className="flex-1 flex flex-col items-center justify-center px-4 z-10 relative">
        {/* HERO CONTENT */}
        <div className="w-full max-w-2xl flex flex-col items-center justify-center text-center py-20">
          {/* Logo/Icon */}
          <div className="mb-8">
            <InboxIcon />
          </div>
          {/* Headline with gradient accent */}
          <h1 className="text-5xl md:text-6xl font-black mb-4 tracking-tight leading-tight">
            <span className="text-white">Free Your </span>
            <span
              className="bg-gradient-to-r from-cyan-400 via-blue-400 to-accent bg-clip-text text-transparent"
              style={{ WebkitTextStroke: '0.5px #14e0c9' }}
            >
              Inbox
            </span>
          </h1>
          {/* Subheadline */}
          <p className="mb-8 text-lg md:text-xl text-gray-300 font-medium">
            Meet Inbox Whisperer, the <span className="font-bold text-accent">AI-powered</span>{' '}
            inbox assistant
            <br className="hidden md:inline" /> that maximizes your productivity and privacy.
          </p>
          {/* CTA Button */}
          <div className="flex flex-col items-center gap-4 w-full">
            {loading ? null : user ? (
              <a
                href="/dashboard"
                className="btn btn-accent btn-lg font-bold rounded-full px-8 shadow-xl transition-transform hover:scale-105 focus:scale-105"
                style={{ boxShadow: '0 4px 24px 0 #14e0c955' }}
              >
                Go to Dashboard
              </a>
            ) : (
              <GoogleSignInButton />
            )}
          </div>
          {/* Optional: Download/alt CTA or secondary links here */}
        </div>
        {/* Optional: Floating screenshot/illustration placeholder */}
        <div className="w-full max-w-4xl mx-auto mt-10 flex items-center justify-center">
          <div className="rounded-2xl overflow-hidden shadow-2xl border border-gray-800 bg-[#181f2a] bg-opacity-90 p-2 flex items-center justify-center min-h-[200px] min-w-[320px]">
            {/* Replace this with a real screenshot or illustration */}
            <span className="text-gray-500 text-lg">
              [ App screenshot or illustration coming soon ]
            </span>
          </div>
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

export default Home;
