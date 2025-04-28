// Explicitly mock the auth module first
jest.mock('../api/generated/auth/auth');
jest.mock('../context/UserContext');

import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import AuthCallback from './AuthCallback';
import * as UserContext from '../context/UserContext';
import * as authModule from '../api/generated/auth/auth';

// Mock UserContext
const mockedUseUser = UserContext as jest.Mocked<typeof UserContext>;

describe('AuthCallback', () => {
  let mutateMock: jest.Mock;
  let originalLocation: Location;
  let getApiAuthCallbackMock: jest.Mock;

  beforeEach(() => {
    // Create a fresh mock function for each test
    getApiAuthCallbackMock = jest.fn();
    
    // Set up the auth module mock
    (authModule.getAuth as jest.Mock).mockReturnValue({
      getApiAuthCallback: getApiAuthCallbackMock
    });

    mutateMock = jest.fn().mockResolvedValue(undefined);
    // Set the mock implementation for useUser
    mockedUseUser.useUser.mockReturnValue({
      user: { id: '123', email: 'test@example.com' },
      loading: false,
      error: null,
      logout: jest.fn(),
      mutate: mutateMock,
    });
    originalLocation = window.location;
    delete (window as any).location;
    (window as any).location = { ...originalLocation, href: '', assign: jest.fn() };
    jest.clearAllMocks();
  });

  afterEach(() => {
    (window as any).location = originalLocation;
    jest.resetAllMocks();
  });

  it('handles successful callback and redirects', async () => {
    getApiAuthCallbackMock.mockResolvedValueOnce({});
    window.location.search = '?code=abc&state=xyz';
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/" element={<div>Home</div>} />
        </Routes>
      </MemoryRouter>
    );
    expect(screen.getByText(/signing you in/i)).toBeInTheDocument();
    await waitFor(() => expect(mutateMock).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText(/success/i)).toBeInTheDocument());
  });

  it('handles onboarding flow', async () => {
    getApiAuthCallbackMock.mockResolvedValueOnce({});
    window.location.search = '?code=abc&state=xyz&first=1';
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz&first=1']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => expect(mutateMock).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText(/welcome to/i)).toBeInTheDocument());
  });

  it('shows error if code or state is missing', async () => {
    window.location.search = '?code=abc';
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => expect(screen.getByText(/missing code or state/i)).toBeInTheDocument());
  });

  it('handles backend error and shows error message', async () => {
    getApiAuthCallbackMock.mockRejectedValueOnce({ response: { data: { error: 'invalid_state' } } });
    const clearAllAuthSpy = jest.spyOn(UserContext, 'clearAllAuth');
    window.location.search = '?code=abc&state=xyz';
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
        </Routes>
      </MemoryRouter>
    );
    
    // Wait for the API call to be made with correct parameters
    await waitFor(() => expect(getApiAuthCallbackMock).toHaveBeenCalledWith({ code: 'abc', state: 'xyz' }));
    
    // Wait for clearAllAuth to be called (which happens in error handling)
    await waitFor(() => expect(clearAllAuthSpy).toHaveBeenCalled());
  });

  it('handles generic backend error', async () => {
    getApiAuthCallbackMock.mockRejectedValueOnce(new Error('fail'));
    const clearAllAuthSpy = jest.spyOn(UserContext, 'clearAllAuth');
    window.location.search = '?code=abc&state=xyz';
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
        </Routes>
      </MemoryRouter>
    );
    
    // Wait for the API call to be made with correct parameters
    await waitFor(() => expect(getApiAuthCallbackMock).toHaveBeenCalledWith({ code: 'abc', state: 'xyz' }));
    
    // Wait for clearAllAuth to be called (which happens in error handling)
    await waitFor(() => expect(clearAllAuthSpy).toHaveBeenCalled());
  });
}); 