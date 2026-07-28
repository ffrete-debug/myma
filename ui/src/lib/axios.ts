import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import Cookies from 'js-cookie';
import useAuthStore from '@/stores/auth';

const AUTH_COOKIE_NAME = 'auth-token';
const LOGIN_PATH = '/login';

/**
 * The shared HTTP client. Every browser-side call should go through this
 * instance so that auth and session expiry are handled in exactly one place.
 *
 * `baseURL` is the page origin on purpose: the browser talks to the Next route
 * handlers under `src/app/api/*`, and those handlers are the only thing that
 * knows how to reach the Go backend. `NEXT_PUBLIC_API_BASE`
 * (e.g. "http://localhost:8080/api") is the *server-side* address of that
 * backend - using it here would turn `/api/plugins` into
 * `http://localhost:8080/api/api/plugins` and bypass the proxy entirely.
 */
const instance = axios.create({
  baseURL: '/',
});

/**
 * Attach the bearer token to every request. The token lives in a non-httpOnly
 * cookie (see `stores/auth.ts`), which is why it is readable from here.
 */
instance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = Cookies.get(AUTH_COOKIE_NAME);
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  return config;
});

/**
 * A login/init rejection is a bad-credentials answer, not an expired session -
 * those forms render the server message themselves and must not be bounced.
 */
const isAuthRequest = (url?: string) => !!url && url.includes('/api/auth/');

/**
 * The JWT expires with no client-visible signal, so a 401 is the first hint the
 * session is gone. Without this every caller renders its own local "failed"
 * string and the user is stranded on a page that can never load again.
 */
instance.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (
      error.response?.status === 401 &&
      typeof window !== 'undefined' &&
      !isAuthRequest(error.config?.url) &&
      window.location.pathname !== LOGIN_PATH
    ) {
      // Clear the in-memory store *and* the cookie, otherwise `middleware.ts`
      // would see a stale token and bounce the user straight back out of /login.
      useAuthStore.getState().actions.logout();
      window.location.assign(LOGIN_PATH);
    }
    return Promise.reject(error);
  },
);

export default instance;
