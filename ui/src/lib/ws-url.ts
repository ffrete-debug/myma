/**
 * Helpers for turning the HTTP API base URL into a WebSocket base URL.
 *
 * `NEXT_PUBLIC_API_BASE` is an HTTP origin with an `/api` suffix
 * (e.g. `https://example.com/api`), while the WebSocket endpoints live under
 * `<origin>/api/ws/...`. So the suffix is dropped and the scheme swapped.
 *
 * The scheme swap is deliberately done with two anchored, mutually exclusive
 * patterns instead of one `^https?:\/\//` match: a combined pattern captures
 * the `://` as part of the match, so comparing the matched text against
 * `'https:'` is never true and every page - including one served over TLS -
 * silently falls back to plaintext `ws://`, which a browser blocks as mixed
 * content.
 */

/** Base URL of the API, defaulting to the local dev backend. */
export function getApiBase(): string {
  return process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080/api';
}

/**
 * Convert an HTTP(S) API base URL into the matching `ws://` / `wss://` origin,
 * with the trailing `/api` path segment removed.
 */
export function toWebSocketBase(apiBase: string): string {
  const origin = apiBase.replace(/\/api\/?$/, '');

  if (/^https:\/\//i.test(origin)) {
    return origin.replace(/^https:\/\//i, 'wss://');
  }
  if (/^http:\/\//i.test(origin)) {
    return origin.replace(/^http:\/\//i, 'ws://');
  }
  if (/^wss?:\/\//i.test(origin)) {
    return origin;
  }

  // Relative base (e.g. NEXT_PUBLIC_API_BASE="/api"). Returning it untouched
  // yields a scheme-less URL that `new WebSocket()` rejects with a SyntaxError,
  // so resolve it against the page origin: an HTTPS page then yields wss://.
  if (typeof window !== 'undefined') {
    return `${window.location.origin.replace(/^http/i, 'ws')}${origin}`;
  }
  return origin;
}

/** Build a full WebSocket URL for `path` (which must start with a slash). */
export function buildWebSocketUrl(path: string, token?: string): string {
  const url = `${toWebSocketBase(getApiBase())}${path}`;
  return token ? `${url}?token=${encodeURIComponent(token)}` : url;
}
