"use client";

import { useThemeStore } from '@/stores/theme';
import { useTranslations } from 'next-intl';
import { DropdownMenu, DropdownMenuContent, DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { Sun, Moon, Monitor } from 'lucide-react';


export function ThemeToggle() {
  const t = useTranslations('common');
  const { theme, setTheme } = useThemeStore();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 px-0">
          {theme === 'dark' ? (
            <Moon className="h-4 w-4" />
          ) : theme === 'light' ? (
            <Sun className="h-4 w-4" />
          ) : (
            <Monitor className="h-4 w-4" />
          )}
          <span className="sr-only">{t('theme')}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-36">
        <DropdownMenuRadioGroup value={theme} onValueChange={(v) => setTheme(v as 'dark' | 'light' | 'auto')}>
          <DropdownMenuRadioItem value="dark">
            <Moon className="mr-2 h-3.5 w-3.5" />
            {t('dark')}
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="light">
            <Sun className="mr-2 h-3.5 w-3.5" />
            {t('light')}
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="auto">
            <Monitor className="mr-2 h-3.5 w-3.5" />
            {t('auto')}
          </DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
