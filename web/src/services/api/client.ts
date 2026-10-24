import { APIError } from '@/types/api';
import { router } from '@/router';
import { useAuthStore } from '@/features/auth/store';

const BASE_URL = '/api/v1';

// Routes that legitimately may 401 without it meaning the session expired —
// these are the unauthenticated probes we make on app boot or as part of a
// flow that intentionally allows the unauth case. We don't want a probe to
// trigger a user-visible redirect.
const SILENT_401_ROUTES: ReadonlyArray<string> = [
  '/auth/login',     // bad creds; the form handles its own UX
  '/auth/register',  // not really 401 territory but keep symmetric
  '/auth/logout',    // already logging out; redirecting on 401 is moot
  '/auth/me',        // boot-time hydration probe; 401 = guest, not a redirect
  '/me/usage',       // fetched on guard mount, may 401 for guests
  '/config',
];

let redirecting = false; // de-bounce — many in-flight 401s should fire one nav
function handleUnauthorized(endpoint: string) {
  if (redirecting) return;
  if (SILENT_401_ROUTES.some((p) => endpoint.startsWith(p))) return;
  redirecting = true;
  // Drop client-side auth state so the next render sees a guest user.
  try { useAuthStore.getState().setUser(null); } catch { /* SSR / boot */ }
  // Use router.navigate (works outside the component tree) and stash the
  // current URL so the user lands back here after re-login.
  const here = window.location.pathname + window.location.search;
  router.navigate(`/login?redirect=${encodeURIComponent(here)}&reason=expired`);
  // Reset the flag after the navigation settles.
  setTimeout(() => { redirecting = false; }, 1000);
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
}

// Reads the CSRF cookie by suffix match so the frontend stays correct no matter
// what SESSION_COOKIE_NAME the deployment uses (diagram_session_csrf,
// clyra_session_csrf, etc.). Backend convention: cookie ends in "_csrf".
function readCsrfToken(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/(?:^|; )([A-Za-z0-9_-]+_csrf)=([^;]*)/);
  return match ? decodeURIComponent(match[2]!) : '';
}

export async function api<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { body, headers: customHeaders, method, ...rest } = options;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(customHeaders as Record<string, string> | undefined),
  };

  // Attach CSRF token on mutating methods. GETs don't need it server-side.
  const upperMethod = (method ?? 'GET').toUpperCase();
  if (upperMethod !== 'GET' && upperMethod !== 'HEAD') {
    const token = readCsrfToken();
    if (token) headers['X-CSRF-Token'] = token;
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    credentials: 'include',
    headers,
    method,
    body: body ? JSON.stringify(body) : undefined,
    ...rest,
  });

  if (!response.ok) {
    const data = await response.json().catch(() => null);
    if (response.status === 401) {
      handleUnauthorized(endpoint);
    }
    throw new APIError(
      response.status,
      data?.error ?? `Request failed (${response.status})`,
      data?.details,
    );
  }

  // Endpoints that legitimately return no body (e.g. 204 on DELETE, or a
  // bodyless 200) shouldn't be JSON-parsed — an empty string blows up
  // response.json() with 'Unexpected end of JSON input'. Callers of these
  // endpoints declare their return type as void.
  if (response.status === 204) return undefined as T;
  const raw = await response.text();
  if (!raw) return undefined as T;
  const data = JSON.parse(raw);
  return (data.data ?? data) as T;
}
