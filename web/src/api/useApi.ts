import useSWR from 'swr';
import { paths } from './types';

// --- Basic fetcher for SWR ---
const fetcher = (url: string) =>
  fetch(url, { credentials: 'include' }).then((res) => {
    if (!res.ok) throw new Error('API error');
    return res.json();
  });

// --- Typed hooks for MVP endpoints ---

// Get current user info (assumes /api/users/me returns User)
export function useCurrentUser() {
  return useSWR<paths['/users/{id}']['get']['responses']['200']>('/api/users/me', fetcher);
}

// Add more hooks as needed, e.g. for onboarding, emails, etc.

// Dummy mutator for Orval compatibility
export const useApi = () => ({});
