import '@testing-library/jest-dom';
jest.mock('../fonts/inter.css', () => ({}));

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import Home from './Home';
import { UserProvider } from '../context/UserContext';

describe('Home', () => {
  it('renders the Inbox Whisperer hero headline', () => {
    render(
      <UserProvider>
        <Home />
      </UserProvider>,
    );
    expect(screen.getAllByText(/Inbox Whisperer/i).length).toBeGreaterThan(0);
  });

  it('shows sign in button when not logged in', async () => {
    // Mock fetch to return null user
    (global.fetch as jest.Mock).mockImplementationOnce(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve(null),
      } as Response)
    );

    render(
      <UserProvider>
        <Home />
      </UserProvider>,
    );

    // Wait for async state updates
    await waitFor(() => {
      expect(screen.getAllByText(/Sign in/i)[0]).toBeInTheDocument();
    });
    expect(screen.getAllByText(/Sign in/i)[0]).toBeInTheDocument();
  });
});
