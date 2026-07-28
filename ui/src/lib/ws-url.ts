/**
 * WebSocket URL construction.
 *
 * `NEXT_PUBLIC_API_BASE` is an HTTP origin with an `/api` suffix
 * (e.g. `https://example.com/api`), or a relative `/api` when the backend is
 * reverse-proxied onto the same origin as the UI.
 *
 * Two things make this trickier than string concatenation:
 *
 * 1. `NEXT_PUBLIC_*` is inlined by Next.js at BUILD time. The published image
 *    is built with the default `http://localhost:8080/api`, which is correct
 *    only when the browser runs on the Docker host. Opened from any other
 *    machine, `localhost` is *the viewer's own computer* and every socket
 *    fails. See resolveApiBase.
 * 2. The scheme must track the page: an `ws://` socket on an `https://` page is
 *    blocked as mixed content.
 */

/**
 * Hosts that mean "this machine". A base pinned to one of these cannot be
 * right for a browser loaded from anywhere else.
 */
const LOOPBACK_HOSTS = new Set(['localhost', '127.0.0.1', '::1', '[::1]', '0.0.0.0']);

/** The subset of `window.location` this module needs, so it can be tested. */
export interface PageLocation {
  hostname: string;
  protocol: string;
  origin: string;
}

/**
 * Reconcile the build-time API base with where the page is actually served from.
 *
 * A base naming a real host is always trusted — an operator who configured it
 * knows better than this function. The rewrite applies only to the specific
 * broken case: a loopback base viewed from a non-loopback page.
 */
export function resolveApiBase(configured: string, page: PageLocation | null): string {
  // Server-side render: there is no page origin, and a loopback base is
  // actually correct for container-to-container calls.
  if (!page) return configured;

  // Relative base — already same-origin, nothing to reconcile.
  if (configured.startsWith('/')) return configured;

  let url: URL;
  try {
    url = new URL(configured);
  } catch {
    // Not parseable; hand it back rather than inventing something.
    return configured;
  }

  if (!LOOPBACK_HOSTS.has(url.hostname)) return configured;
  // Genuinely local development: loopback base, loopback page. Correct as-is.
  if (LOOPBACK_HOSTS.has(page.hostname)) return configured;

  // Loopback base but a remote page. The baked value cannot work here, so fall
  // back to the host the page came from, keeping the configured port. The
  // scheme follows the page: an http:// API on an https:// page is blocked as
  // mixed content, so matching is the only browser-viable option. Operators
  // terminating TLS should reverse-proxy the backend and set
  // NEXT_PUBLIC_API_BASE=/api instead of relying on this.
  url.hostname = page.hostname;
  url.protocol = page.protocol;
  return url.toString().replace(/\/$/, '');
}

/** The API base to use from the browser, reconciled with the current page. */
export function getApiBase(): string {
  const configured = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080/api';
  const page = typeof window === 'undefined' ? null : window.location;
  return resolveApiBase(configured, page);
}

/**
 * Convert an HTTP(S) API base into a WebSocket origin.
 *
 * Note the two anchored, mutually exclusive patterns. A single
 * `/^https?:\/\//` would match `https://` and hand the callback the whole
 * `"https://"` string, so a comparison against `"https:"` is never true and
 * every deployment silently gets plaintext `ws://`. That exact bug shipped here
 * once; keep these separate.
 */
export function toWebSocketBase(apiBase: string): string {
  const origin = apiBase.replace(/\/api\/?$/, '');

  if (/^https:\/\//i.test(origin)) return origin.replace(/^https:\/\//i, 'wss://');
  if (/^http:\/\//i.test(origin)) return origin.replace(/^http:\/\//i, 'ws://');

  // Relative base (e.g. NEXT_PUBLIC_API_BASE="/api"): resolve against the page
  // origin so the socket inherits its host and scheme.
  if (typeof window !== 'undefined') {
    return `${window.location.origin.replace(/^http/i, 'ws')}${origin}`;
  }
  return origin;
}

/** Build a full WebSocket URL for a backend path, with the auth token. */
export function buildWebSocketUrl(path: string, token: string): string {
  const base = toWebSocketBase(getApiBase());
  const sep = path.startsWith('/') ? '' : '/';
  return `${base}${sep}${path}?token=${encodeURIComponent(token)}`;
}
