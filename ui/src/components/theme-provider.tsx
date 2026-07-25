"use client";

import { useEffect } from 'react';
import { initTheme, setupThemeChangeListener } from '@/stores/theme';

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    initTheme();
    return setupThemeChangeListener();
  }, []);

  return <>{children}</>;
}