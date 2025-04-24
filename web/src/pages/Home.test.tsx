jest.mock('../fonts/inter.css', () => ({}));
import React from 'react';
import { render, screen } from '@testing-library/react';
import Home from './Home';
import { UserProvider } from '../context/UserContext';

describe('Home', () => {
  it('renders the Inbox Whisperer hero headline', () => {
    render(
      <UserProvider>
        <Home />
      </UserProvider>
    );
    expect(screen.getAllByText(/Inbox Whisperer/i).length).toBeGreaterThan(0);
  });

  it('renders the Sign in button when not logged in', () => {
    render(
      <UserProvider>
        <Home />
      </UserProvider>
    );
    expect(screen.getAllByText(/Sign in/i)[0]).toBeInTheDocument();
  });
});
