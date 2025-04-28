import '@testing-library/jest-dom';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { UserProvider, useUser, clearAllAuth } from './UserContext';
import { MemoryRouter } from 'react-router-dom';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock window.location
const mockLocation = {
  href: 'http://localhost',
  pathname: '/',
  origin: 'http://localhost',
  assign: jest.fn(),
  replace: jest.fn(),
  reload: jest.fn(),
} as Partial<Location>;

Object.defineProperty(window, 'location', {
  writable: true,
  value: new Proxy(mockLocation, {
    set: (obj: any, prop: string, value: any) => {
      obj[prop] = value;
      return true;
    },
  }),
});

// Mock document.cookie
Object.defineProperty(document, 'cookie', {
  writable: true,
  value: 'session_id=test-session',
});

describe('UserContext', () => {
  beforeEach(() => {
    // Reset mocks and cookies before each test
    jest.clearAllMocks();
    document.cookie = 'session_id=test-session';
    window.location.href = 'http://localhost';
    localStorage.clear();
    sessionStorage.clear();
  });

  it('fetches and sets user data on mount', async () => {
    const mockUser = { id: '123', email: 'test@example.com' };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockUser),
    });

    const TestComponent = () => {
      const { user } = useUser();
      return <div>{user?.email}</div>;
    };

    render(
      <MemoryRouter>
        <UserProvider>
          <TestComponent />
        </UserProvider>
      </MemoryRouter>
    );

    // Should show loading state initially
    expect(screen.queryByText(mockUser.email)).not.toBeInTheDocument();

    // Should show user email after loading
    await waitFor(() => {
      expect(screen.getByText(mockUser.email)).toBeInTheDocument();
    });
  });

  it('handles session expiry correctly', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
    });

    const TestComponent = () => {
      const { user } = useUser();
      return <div>{user ? 'Logged in' : 'Not logged in'}</div>;
    };

    render(
      <MemoryRouter>
        <UserProvider>
          <TestComponent />
        </UserProvider>
      </MemoryRouter>
    );

    await waitFor(() => {
      // Should redirect to login with session_expired reason
      expect(window.location.href).toBe('http://localhost/login?reason=session_expired');
      // Should clear session cookie
      expect(document.cookie).not.toContain('session_id=test-session');
    });
  });

  it('clears all auth data properly', () => {
    // Set up test data
    document.cookie = 'session_id=test-session; path=/';
    document.cookie = 'other_cookie=value; path=/api';
    localStorage.setItem('test', 'value');
    sessionStorage.setItem('test', 'value');

    // Clear auth
    clearAllAuth();

    // Verify cookies are cleared
    expect(document.cookie).not.toContain('session_id=test-session');
    expect(document.cookie).not.toContain('other_cookie=value');

    // Verify storage is cleared
    expect(localStorage.getItem('test')).toBeNull();
    expect(sessionStorage.getItem('test')).toBeNull();
  });

  it('skips user fetch on login page', async () => {
    // Set current path to /login
    window.location.pathname = '/login';

    render(
      <MemoryRouter>
        <UserProvider>
          <div>Test</div>
        </UserProvider>
      </MemoryRouter>
    );

    // Should not make fetch request
    await waitFor(() => {
      expect(mockFetch).not.toHaveBeenCalled();
    });
  });

  it('logs out and resets user context', async () => {
    const mockUser = { id: '123', email: 'test@example.com' };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockUser),
    });
    const TestComponent = () => {
      const { user, logout } = useUser();
      return (
        <div>
          <span>{user?.email || 'No user'}</span>
          <button onClick={logout}>Logout</button>
        </div>
      );
    };
    render(
      <MemoryRouter>
        <UserProvider>
          <TestComponent />
        </UserProvider>
      </MemoryRouter>
    );
    await waitFor(() => {
      expect(screen.getByText('test@example.com')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Logout'));
    await waitFor(() => {
      expect(window.location.href).toBe('/login');
    });
  });
});
