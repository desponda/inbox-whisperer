import '@testing-library/jest-dom';
jest.mock('../fonts/inter.css', () => ({}));

import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Login from './Login';

describe('Login', () => {
  beforeEach(() => {
    // Clear query parameters before each test
    window.history.pushState({}, '', '/login');
  });

  it('renders login page without session expired message by default', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );
    
    // Should show login elements
    expect(screen.getByText(/Welcome to Inbox Whisperer/i)).toBeInTheDocument();
    expect(screen.getByText(/Sign in with Google/i)).toBeInTheDocument();
    
    // Should not show session expired message
    expect(screen.queryByText(/Your session has expired/i)).not.toBeInTheDocument();
  });

  it('shows session expired message when redirected with reason', async () => {
    // Set URL with session_expired reason
    window.history.pushState({}, '', '/login?reason=session_expired');

    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );

    // Should show session expired message
    expect(await screen.findByText(/Your session has expired/i)).toBeInTheDocument();
  });
});
