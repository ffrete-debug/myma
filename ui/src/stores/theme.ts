import { create } from 'zustand';

export type ThemeMode = 'dark' | 'light' | 'auto';

interface ThemeState {
  theme: ThemeMode;
  resolvedTheme: 'dark' | 'light';
  setTheme: (theme: ThemeMode) => void;
}

function detectSystemTheme(): 'dark' | 'light' {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme: ThemeMode): 'dark' | 'light' {
  const resolved = theme === 'auto' ? detectSystemTheme() : theme;
  document.documentElement.setAttribute('data-theme', resolved);
  return resolved;
}

export const useThemeStore = create<ThemeState>((set) => ({
  theme: (typeof window !== 'undefined'
    ? (localStorage.getItem('theme') as ThemeMode)
    : null) || 'dark',
  resolvedTheme: 'dark',

  setTheme: (theme) => {
    localStorage.setItem('theme', theme);
    const resolved = applyTheme(theme);
    set({ theme, resolvedTheme: resolved });
  },
}));

export function initTheme(): void {
  const theme = (localStorage.getItem('theme') as ThemeMode) || 'dark';
  const resolved = applyTheme(theme);
  useThemeStore.setState({ theme, resolvedTheme: resolved });
}

export function setupThemeChangeListener(): () => void {
  const mql = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = () => {
    const { theme } = useThemeStore.getState();
    if (theme === 'auto') {
      const resolved = detectSystemTheme();
      document.documentElement.setAttribute('data-theme', resolved);
      useThemeStore.setState({ resolvedTheme: resolved });
    }
  };
  mql.addEventListener('change', handler);
  return () => mql.removeEventListener('change', handler);
}