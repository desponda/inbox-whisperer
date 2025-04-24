import React from 'react';
import '/fonts/inter.css';
import { useUser } from '../context/UserContext';
import { Link } from 'react-router-dom';

// Mocked email data for preview
const MOCK_EMAILS = [
  {
    id: '1',
    subject: 'Welcome to Inbox Whisperer!',
    from: 'team@inboxwhisperer.com',
    snippet: 'Get started with your AI-powered inbox assistant...',
  },
  {
    id: '2',
    subject: 'Your AI Summary is Ready',
    from: 'ai@inboxwhisperer.com',
    snippet: 'Here is a summary of your latest emails...',
  },
  {
    id: '3',
    subject: 'Weekly Productivity Insights',
    from: 'insights@inboxwhisperer.com',
    snippet: 'See how you managed your inbox this week.',
  },
];

const Dashboard: React.FC = () => {
  const { user, loading } = useUser();

  return (
    <main
      className="min-h-screen flex flex-col text-gray-100 bg-[#0e1015]"
      style={{ fontFamily: "'DM Sans', 'Inter', ui-sans-serif, system-ui, sans-serif" }}
    >
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
      <nav className="w-full flex items-center justify-between px-4 py-3 bg-transparent z-10">
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
          <span className="text-xl font-normal tracking-tight text-white">Inbox Whisperer</span>
        </div>
        <div className="hidden md:flex items-center gap-8">
          <Link to="/" className="text-base font-normal text-gray-300 hover:text-accent transition">
            Home
          </Link>
          <Link to="/dashboard" className="text-base font-bold text-accent">
            Dashboard
          </Link>
        </div>
      </nav>
      {/* DASHBOARD HEADER */}
      <section className="flex flex-col items-center justify-center pt-8 pb-2 px-4 z-10 relative">
        <h1 className="text-3xl md:text-4xl font-black mb-2 tracking-tight leading-tight text-white">
          {loading ? 'Loading…' : user ? `Welcome, ${user.email}` : 'Welcome!'}
        </h1>
        <p className="text-lg text-gray-300 mb-6">
          Here’s a preview of your inbox, prioritized by AI.
        </p>
      </section>
      {/* INBOX PREVIEW */}
      <section className="flex flex-col items-center justify-center px-4 z-10 relative">
        <div className="w-full max-w-2xl bg-[#181f2a] bg-opacity-90 rounded-2xl shadow-2xl border border-gray-800 p-6">
          <h2 className="text-2xl font-bold mb-4 text-accent">Your Inbox</h2>
          <ul className="divide-y divide-gray-800">
            {MOCK_EMAILS.map((email) => (
              <li
                key={email.id}
                className="py-4 flex flex-col md:flex-row md:items-center md:justify-between group hover:bg-[#23243a] rounded-lg px-2 transition"
              >
                <div>
                  <div className="font-semibold text-white group-hover:text-accent transition text-lg">
                    {email.subject}
                  </div>
                  <div className="text-sm text-gray-400">From: {email.from}</div>
                  <div className="text-base text-gray-300 mt-1">{email.snippet}</div>
                </div>
                <button className="btn btn-outline btn-accent mt-2 md:mt-0 md:ml-4">Open</button>
              </li>
            ))}
          </ul>
        </div>
      </section>
      {/* FOOTER */}
      <footer className="footer footer-center p-4 bg-neutral bg-opacity-90 text-gray-400 border-t border-neutral-content/10 shadow-inner z-10 mt-8">
        <div>
          <p>
            &copy; {new Date().getFullYear()} Inbox Whisperer &mdash; Built with{' '}
            <span className="text-accent font-semibold">DaisyUI</span> and{' '}
            <span className="text-primary font-semibold">Windsurf</span>
          </p>
        </div>
      </footer>
    </main>
  );
};

export default Dashboard;
