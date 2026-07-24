import { create } from 'zustand';
import axios from 'axios';
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
        Cookies.set('auth-token', token, { expires: 7 });
        return { success: true, message };
      } catch (error) {
        if (axios.isAxiosError(error) && error.response) {
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
        Cookies.set('auth-token', token, { expires: 7 });
        return { success: true, message };
      } catch (error) {
        if (axios.isAxiosError(error) && error.response) {
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
        console.error('Failed to get user info:', error);
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
