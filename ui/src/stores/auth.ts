import { create } from 'zustand';
// Use the configured instance, not bare axios: it carries the 401 handler
// that clears a dead session and sends the user to /login. Calls made with
// bare axios skipped it, so a token for a user that no longer exists (after
// the database is reset, say) produced a console error per caller and left
// the page stranded instead of redirecting once.
import { isAxiosError } from 'axios';
import axios from '@/lib/axios';
import Cookies from 'js-cookie';

// Define user object type
interface User {
  id: number;
  username: string;
  is_admin: boolean;
  created_at: string;
}

// Define auth state type
interface AuthState {
  token: string | null;
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isInitialized: boolean; // Add initialization state
  actions: AuthActions;
}

// Define auth actions type
interface AuthActions {
  checkInit: () => Promise<boolean>;
  init: (credentials: Credentials) => Promise<{ success: boolean; message: string }>;
  login: (credentials: Credentials) => Promise<{ success: boolean; message: string }>;
  getProfile: () => Promise<void>;
  logout: () => void;
  initFromStorage: () => void;
}

// Define login credentials type
interface Credentials {
  username?: string;
  password?: string;
}

// Define API response type
interface AuthResponse {
  token: string;
  user: User;
  message: string;
}

// Cookie options for the auth token.
// NOTE: js-cookie writes from JavaScript, so the cookie can never be httpOnly and
// stays readable by any script on the page. The real fix is for the backend to
// issue auth-token as an httpOnly cookie on /api/auth/login and /api/auth/init;
// these flags are the best hardening available on the client until then.
const authCookieOptions = () => ({
  expires: 7,
  path: '/',
  sameSite: 'strict' as const,
  // Only an https origin can carry a secure cookie - on plain http (including
  // http://localhost during development) the browser would drop it silently.
  secure: typeof window !== 'undefined' && window.location.protocol === 'https:',
});

const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isLoading: false,
  isAuthenticated: false,
  isInitialized: false, // Initialize to false
  actions: {
    checkInit: async () => {
      try {
        const response = await axios.get('/api/auth/check-init');
        return response.data.initialized;
      } catch (error) {
        console.error('Failed to check initialization status:', error);
        return false;
      }
    },
    init: async (credentials) => {
      set({ isLoading: true });
      try {
        const response = await axios.post<AuthResponse>('/api/auth/init', credentials);
        const { token, user, message } = response.data;
        set({ token, user, isAuthenticated: true });
        Cookies.set('auth-token', token, authCookieOptions());
        return { success: true, message };
      } catch (error) {
        if (isAxiosError(error) && error.response) {
          return { success: false, message: error.response.data?.error || 'Initialization failed' };
        }
        return { success: false, message: 'Initialization failed' };
      } finally {
        set({ isLoading: false });
      }
    },
    login: async (credentials) => {
      set({ isLoading: true });
      try {
        const response = await axios.post<AuthResponse>('/api/auth/login', credentials);
        const { token, user, message } = response.data;
        set({ token, user, isAuthenticated: true });
        Cookies.set('auth-token', token, authCookieOptions());
        return { success: true, message };
      } catch (error) {
        if (isAxiosError(error) && error.response) {
          return { success: false, message: error.response.data?.error || 'Login failed' };
        }
        return { success: false, message: 'Login failed' };
      } finally {
        set({ isLoading: false });
      }
    },
    getProfile: async () => {
      try {
        const token = get().token;
        if (!token) return;
        const response = await axios.get<{ user: User }>('/api/profile', {
          headers: { Authorization: `Bearer ${token}` },
        });
        set({ user: response.data.user });
      } catch (error) {
        // A 401 here means the session is dead - the token outlived its user,
        // which is exactly what happens after the database is reset. The shared
        // interceptor already clears the cookie and moves the user on, so this
        // is an expected, handled outcome and must not be logged as a failure:
        // every caller doing so is what produced the console full of errors on a
        // fresh install.
        if (!isAxiosError(error) || error.response?.status !== 401) {
          console.error('Failed to get user info:', error);
        }
        get().actions.logout();
      }
    },
    logout: () => {
      set({ token: null, user: null, isAuthenticated: false });
      Cookies.remove('auth-token');
    },
    initFromStorage: () => {
      const token = Cookies.get('auth-token');
      if (token) {
        const currentState = get();
        set({ token, isAuthenticated: true, isInitialized: true });
        // Only call getProfile if user info is missing
        if (!currentState.user) {
          get().actions.getProfile();
        }
      } else {
        set({ isInitialized: true });
      }
    },
  },
}));

export const useAuthActions = () => useAuthStore((state) => state.actions);
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated);
export const useAuthUser = () => useAuthStore((state) => state.user);
export const useAuthIsLoading = () => useAuthStore((state) => state.isLoading);
export const useAuthIsInitialized = () => useAuthStore((state) => state.isInitialized);

export default useAuthStore;
