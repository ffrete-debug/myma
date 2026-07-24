import Cookies from 'js-cookie';

export const locales = ['en', 'zh'] as const;
export type Locale = typeof locales[number];
export const defaultLocale: Locale = 'en';

const LOCALE_COOKIE_NAME = 'NEXT_LOCALE';

// Get locale on client side
export function getClientLocale(): Locale {
  if (typeof window === 'undefined') {
    return defaultLocale;
  }
  
  const locale = Cookies.get(LOCALE_COOKIE_NAME);
  
  if (locale && locales.includes(locale as Locale)) {
    return locale as Locale;
  }
  
  return defaultLocale;
}

// Set locale on client side
export function setClientLocale(locale: Locale) {
  if (typeof window === 'undefined') {
    return;
  }
  
  Cookies.set(LOCALE_COOKIE_NAME, locale, {
    expires: 365, // 1 year
    path: '/',
    sameSite: 'lax'
  });
}
