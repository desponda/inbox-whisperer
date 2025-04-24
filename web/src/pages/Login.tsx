import React from 'react';
import '/fonts/inter.css';

const Login: React.FC = () => (
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
        background: 'radial-gradient(ellipse 120% 80% at 50% 0%, #23243a 80%, #0e1015 100%), linear-gradient(120deg,rgba(20,224,201,0.07) 0%,rgba(0,180,216,0.07) 100%)',
        opacity: 1
      }}
    />
    {/* NAVBAR */}
    <nav className="w-full flex items-center justify-between px-4 py-3 bg-transparent z-10" style={{ fontFamily: "'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif", boxShadow: 'none', border: 'none' }}>
      <div className="flex items-center gap-2">
        <svg width="28" height="28" fill="none" viewBox="0 0 24 24"><rect width="24" height="24" rx="6" fill="#14e0c9"/><path d="M7 12l5 5 5-5" stroke="#23272f" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/></svg>
        <span className="text-xl font-normal tracking-tight text-white" style={{ fontFamily: 'inherit', fontWeight: 400, letterSpacing: '-0.01em' }}>Inbox Whisperer</span>
      </div>
      <div className="hidden md:flex items-center gap-8">
        <a href="/#features" className="text-base font-normal text-gray-300 hover:text-accent transition">Features</a>
        <a href="/#about" className="text-base font-normal text-gray-300 hover:text-accent transition">About</a>
        <a href="https://github.com/desponda/inbox-whisperer" target="_blank" rel="noopener noreferrer" className="text-base font-normal text-gray-300 hover:text-accent transition">GitHub</a>
        <a href="/login" className="ml-4 btn btn-accent font-semibold rounded-full px-6">Sign in</a>
      </div>
      <div className="md:hidden">
        {/* Mobile menu placeholder */}
      </div>
    </nav>
    {/* LOGIN SECTION */}
    <section className="flex-1 flex flex-col items-center justify-center px-4 z-10 relative">
      <div className="w-full max-w-md flex flex-col items-center justify-center text-center py-16">
        <h2 className="text-3xl md:text-4xl font-black mb-6 tracking-tight leading-tight">
          <span className="text-white">Sign in to </span>
          <span className="bg-gradient-to-r from-cyan-400 via-blue-400 to-accent bg-clip-text text-transparent" style={{ WebkitTextStroke: '0.5px #14e0c9' }}>Inbox Whisperer</span>
        </h2>
        <a
          href="/api/auth/login"
          className="btn btn-accent btn-lg font-bold rounded-full px-8 shadow-xl transition-transform hover:scale-105 focus:scale-105 mb-4"
          style={{ boxShadow: '0 4px 24px 0 #14e0c955' }}
          data-testid="google-login-btn"
        >
          <svg className="w-5 h-5 mr-2 inline-block" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><g><path fill="#4285F4" d="M24 9.5c3.54 0 6.7 1.22 9.19 3.22l6.85-6.85C36.64 2.42 30.79 0 24 0 14.82 0 6.51 5.8 2.23 14.09l7.98 6.2C12.06 13.36 17.57 9.5 24 9.5z"/><path fill="#34A853" d="M46.1 24.55c0-1.64-.15-3.22-.42-4.74H24v9.01h12.41c-.54 2.9-2.18 5.36-4.65 7.01l7.22 5.62C43.83 37.27 46.1 31.39 46.1 24.55z"/><path fill="#FBBC05" d="M10.21 28.29c-1.13-3.36-1.13-6.93 0-10.29l-7.98-6.2C-1.13 16.16-1.13 31.84 2.23 33.91l7.98-6.2z"/><path fill="#EA4335" d="M24 44c6.48 0 11.92-2.14 15.9-5.84l-7.22-5.62c-2.02 1.36-4.6 2.16-8.68 2.16-6.43 0-11.94-3.86-13.79-9.29l-7.98 6.2C6.51 42.2 14.82 48 24 48z"/><path fill="none" d="M0 0h48v48H0z"/></g></svg>
          Sign in with Google
        </a>
        <p className="text-base text-gray-300 mt-2">We never see your password. Google login is required for inbox access.</p>
      </div>
    </section>
    {/* FOOTER */}
    <footer className="footer footer-center p-4 bg-neutral bg-opacity-90 text-gray-400 border-t border-neutral-content/10 shadow-inner z-10">
      <div>
        <p>
          &copy; {new Date().getFullYear()} Inbox Whisperer &mdash; Built with <span className="text-accent font-semibold">DaisyUI</span> and <span className="text-primary font-semibold">Windsurf</span>
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

export default Login;
